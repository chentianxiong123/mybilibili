package com.mybilibili.search.service.impl;

import com.mybilibili.common.vo.VideoSearchVO;
import com.mybilibili.search.document.ManuscriptDocument;
import com.mybilibili.search.service.HotSearchService;
import com.mybilibili.search.service.VideoSearchService;
import lombok.extern.slf4j.Slf4j;
import co.elastic.clients.elasticsearch._types.SortOptions;
import co.elastic.clients.elasticsearch._types.SortOrder;
import co.elastic.clients.elasticsearch._types.query_dsl.Query;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.elasticsearch.client.elc.ElasticsearchTemplate;
import org.springframework.data.elasticsearch.client.elc.NativeQuery;
import org.springframework.data.elasticsearch.client.elc.NativeQueryBuilder;
import org.springframework.data.elasticsearch.client.elc.QueryBuilders;
import org.springframework.data.elasticsearch.core.SearchHit;
import org.springframework.data.elasticsearch.core.SearchHits;
import org.springframework.data.elasticsearch.core.query.HighlightQuery;
import org.springframework.data.elasticsearch.core.query.highlight.Highlight;
import org.springframework.data.elasticsearch.core.query.highlight.HighlightField;
import org.springframework.data.elasticsearch.core.query.highlight.HighlightFieldParameters;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@Primary
public class VideoSearchServiceImpl implements VideoSearchService {

    @Autowired
    private ElasticsearchTemplate elasticsearchTemplate;

    @Autowired
    private HotSearchService hotSearchService;

    private static final Integer PUBLISHED_STATUS = 3;
    private static final String SORT_RELEVANCE = "relevance";
    private static final String SORT_TIME = "time";
    private static final String SORT_HOT = "hot";

    @Override
    public Page<VideoSearchVO> search(String keyword, Integer categoryId, String tag, Integer userId,
                                      String sort, int page, int size) {
        try {
            Pageable pageable = PageRequest.of(Math.max(0, page - 1), size);

            NativeQueryBuilder queryBuilder = new NativeQueryBuilder()
                    .withQuery(q -> q.bool(b -> {
                        // 暂时移除status过滤，测试数据
                        // b.must(m -> m.term(t -> t.field("status").value(PUBLISHED_STATUS)));

                        if (StringUtils.hasText(keyword)) {
                            b.must(m -> m.bool(kb -> {
                                kb.should(s -> s.match(mt -> mt.field("title").query(keyword).boost(3.0f)));
                                kb.should(s -> s.match(mt -> mt.field("description").query(keyword).boost(1.5f)));
                                kb.should(s -> s.match(mt -> mt.field("videoTitles").query(keyword).boost(2.0f)));
                                kb.minimumShouldMatch("1");
                                return kb;
                            }));
                        }

                        if (categoryId != null) {
                            b.filter(f -> f.term(t -> t.field("categoryId").value(categoryId.toString())));
                        }

                        if (StringUtils.hasText(tag)) {
                            b.filter(f -> f.term(t -> t.field("tags").value(tag)));
                        }

                        if (userId != null) {
                            b.filter(f -> f.term(t -> t.field("userId").value(userId.toString())));
                        }

                        return b;
                    }))
                    .withPageable(pageable);

            // 添加高亮
            HighlightQuery highlightQuery = buildHighlight();
            queryBuilder.withHighlightQuery(highlightQuery);

            // 添加排序
            if (!SORT_RELEVANCE.equals(sort)) {
                SortOptions sortOptions = buildSort(sort);
                queryBuilder.withSort(sortOptions);
            }

            if (StringUtils.hasText(keyword)) {
                try {
                    hotSearchService.incrementHotSearch(keyword);
                } catch (Exception e) {
                    log.warn("更新热搜榜失败，keyword: {}, error: {}", keyword, e.getMessage());
                }
            }

            NativeQuery searchQuery = queryBuilder.build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchTemplate.search(
                    searchQuery, ManuscriptDocument.class);

            log.info("搜索命中 {} 条记录", searchHits.getTotalHits());

            List<VideoSearchVO> voList = new ArrayList<>();
            for (SearchHit<ManuscriptDocument> hit : searchHits.getSearchHits()) {
                VideoSearchVO vo = convertToVO(hit);
                if (vo != null) {
                    voList.add(vo);
                }
            }

            return new PageImpl<>(voList, pageable, searchHits.getTotalHits());

        } catch (Exception e) {
            log.error("搜索视频失败，keyword: {}, error: {}", keyword, e.getMessage(), e);
            return Page.empty();
        }
    }

    @Override
    public List<String> suggest(String keyword) {
        if (!StringUtils.hasText(keyword)) {
            return new ArrayList<>();
        }

        try {
            NativeQuery searchQuery = new NativeQueryBuilder()
                    .withQuery(q -> q.bool(b -> b
                            .must(m -> m.term(t -> t.field("status").value(PUBLISHED_STATUS)))
                            .must(m -> m.bool(inner -> inner
                                    .should(s -> s.matchPhrasePrefix(mp -> mp.field("title").query(keyword).boost(3.0f)))
                                    .should(s -> s.prefix(p -> p.field("tags").value(keyword).boost(2.0f)))
                                    .should(s -> s.matchPhrasePrefix(mp -> mp.field("description").query(keyword).boost(1.0f)))
                            ))
                    ))
                    .withPageable(PageRequest.of(0, 10))
                    .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchTemplate.search(
                    searchQuery, ManuscriptDocument.class);

            List<String> suggestions = new ArrayList<>();
            for (var hit : searchHits.getSearchHits()) {
                ManuscriptDocument doc = hit.getContent();
                if (doc.getTitle() != null && !suggestions.contains(doc.getTitle())) {
                    suggestions.add(doc.getTitle());
                }
                if (suggestions.size() >= 8) break;
            }

            if (suggestions.size() < 8) {
                for (var hit : searchHits.getSearchHits()) {
                    ManuscriptDocument doc = hit.getContent();
                    if (doc.getTags() != null) {
                        for (String tag : doc.getTags()) {
                            if (tag.contains(keyword) && !suggestions.contains(tag)) {
                                suggestions.add(tag);
                                if (suggestions.size() >= 10) break;
                            }
                        }
                    }
                    if (suggestions.size() >= 10) break;
                }
            }

            return suggestions;

        } catch (Exception e) {
            log.error("获取搜索建议失败, keyword: {}, error: {}", keyword, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    private SortOptions buildSort(String sort) {
        if (!StringUtils.hasText(sort)) {
            sort = SORT_RELEVANCE;
        }

        switch (sort) {
            case SORT_TIME:
                return SortOptions.of(s -> s.field(f -> f.field("uploadTime").order(SortOrder.Desc)));
            case SORT_HOT:
                return SortOptions.of(s -> s.field(f -> f.field("viewCount").order(SortOrder.Desc)));
            case SORT_RELEVANCE:
            default:
                return SortOptions.of(s -> s.score(sc -> sc.order(SortOrder.Desc)));
        }
    }

    private HighlightQuery buildHighlight() {
        HighlightFieldParameters params = HighlightFieldParameters.builder()
                .withPreTags("<em>")
                .withPostTags("</em>")
                .build();

        List<HighlightField> fields = Arrays.asList(
                new HighlightField("title", params),
                new HighlightField("description", params),
                new HighlightField("videoTitles", params)
        );

        Highlight highlight = new Highlight(fields);
        return new HighlightQuery(highlight, ManuscriptDocument.class);
    }

    private VideoSearchVO convertToVO(SearchHit<ManuscriptDocument> searchHit) {
        try {
            ManuscriptDocument document = searchHit.getContent();
            VideoSearchVO vo = new VideoSearchVO();

            vo.setManuscriptId(document.getManuscriptId());
            vo.setTitle(document.getTitle());
            vo.setDescription(document.getDescription());
            vo.setUserId(document.getUserId());
            vo.setUserName(document.getUserName());
            vo.setCategoryId(document.getCategoryId());
            vo.setCategoryName(document.getCategoryName());
            vo.setTags(document.getTags());
            vo.setViewCount(document.getViewCount());
            vo.setLikeCount(document.getLikeCount());
            vo.setCommentCount(document.getCommentCount());
            vo.setShareCount(document.getShareCount());
            vo.setCollectCount(document.getCollectCount());
            vo.setCoinCount(document.getCoinCount());
            vo.setDurationSeconds(document.getDurationSeconds());
            vo.setDuration(formatDuration(document.getDurationSeconds()));
            vo.setUploadTime(document.getUploadTime());
            vo.setStatus(document.getStatus());
            vo.setCoverUrl(document.getCoverUrl());
            vo.setVideoIds(document.getVideoIds());
            vo.setVideoCount(document.getVideoCount());

            return vo;
        } catch (Exception e) {
            log.error("转换VO失败: {}", e.getMessage(), e);
            return null;
        }
    }

    private String formatDuration(Integer seconds) {
        if (seconds == null || seconds <= 0) {
            return "00:00";
        }

        int hours = seconds / 3600;
        int minutes = (seconds % 3600) / 60;
        int secs = seconds % 60;

        if (hours > 0) {
            return String.format("%02d:%02d:%02d", hours, minutes, secs);
        } else {
            return String.format("%02d:%02d", minutes, secs);
        }
    }
}
