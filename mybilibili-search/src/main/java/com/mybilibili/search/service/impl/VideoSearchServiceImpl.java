package com.mybilibili.search.service.impl;

import com.mybilibili.common.vo.VideoSearchVO;
import com.mybilibili.search.document.ManuscriptDocument;
import com.mybilibili.search.service.HotSearchService;
import com.mybilibili.search.service.VideoSearchService;
import lombok.extern.slf4j.Slf4j;
import org.elasticsearch.index.query.BoolQueryBuilder;
import org.elasticsearch.index.query.QueryBuilders;
import org.elasticsearch.search.fetch.subphase.highlight.HighlightBuilder;
import org.elasticsearch.search.sort.FieldSortBuilder;
import org.elasticsearch.search.sort.SortBuilder;
import org.elasticsearch.search.sort.SortBuilders;
import org.elasticsearch.search.sort.SortOrder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.elasticsearch.core.ElasticsearchRestTemplate;
import org.springframework.data.elasticsearch.core.SearchHit;
import org.springframework.data.elasticsearch.core.SearchHits;
import org.springframework.data.elasticsearch.core.query.NativeSearchQuery;
import org.springframework.data.elasticsearch.core.query.NativeSearchQueryBuilder;
import org.springframework.context.annotation.Primary;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@Primary
public class VideoSearchServiceImpl implements VideoSearchService {

    @Autowired
    private ElasticsearchRestTemplate elasticsearchRestTemplate;

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
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();

            // 暂时移除status过滤，测试数据
            // boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            if (StringUtils.hasText(keyword)) {
                BoolQueryBuilder keywordQuery = QueryBuilders.boolQuery();
                keywordQuery.should(QueryBuilders.matchQuery("title", keyword).boost(3.0f));
                keywordQuery.should(QueryBuilders.matchQuery("description", keyword).boost(1.5f));
                keywordQuery.should(QueryBuilders.matchQuery("videoTitles", keyword).boost(2.0f));
                keywordQuery.minimumShouldMatch(1);
                boolQuery.must(keywordQuery);
            }

            if (categoryId != null) {
                boolQuery.filter(QueryBuilders.termQuery("categoryId", categoryId));
            }

            if (StringUtils.hasText(tag)) {
                boolQuery.filter(QueryBuilders.termQuery("tags", tag));
            }

            if (userId != null) {
                boolQuery.filter(QueryBuilders.termQuery("userId", userId));
            }

            if (StringUtils.hasText(keyword)) {
                try {
                    hotSearchService.incrementHotSearch(keyword);
                } catch (Exception e) {
                    log.warn("更新热搜榜失败，keyword: {}, error: {}", keyword, e.getMessage());
                }
            }

            SortBuilder<?> sortBuilder = buildSort(sort);
            HighlightBuilder highlightBuilder = buildHighlight();
            Pageable pageable = PageRequest.of(Math.max(0, page - 1), size);

            NativeSearchQueryBuilder queryBuilder = new NativeSearchQueryBuilder()
                    .withQuery(boolQuery)
                    .withPageable(pageable)
                    .withHighlightBuilder(highlightBuilder);

            if (!SORT_RELEVANCE.equals(sort)) {
                queryBuilder.withSort(sortBuilder);
            }

            NativeSearchQuery searchQuery = queryBuilder.build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
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
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery()
                    .must(QueryBuilders.prefixQuery("title", keyword))
                    .must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                    .withQuery(boolQuery)
                    .withPageable(PageRequest.of(0, 10))
                    .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                    searchQuery, ManuscriptDocument.class);

            return searchHits.getSearchHits().stream()
                    .map(hit -> hit.getContent().getTitle())
                    .distinct()
                    .limit(10)
                    .collect(Collectors.toList());

        } catch (Exception e) {
            log.error("获取搜索建议失败，keyword: {}, error: {}", keyword, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    private SortBuilder<?> buildSort(String sort) {
        if (!StringUtils.hasText(sort)) {
            sort = SORT_RELEVANCE;
        }

        switch (sort) {
            case SORT_TIME:
                return SortBuilders.fieldSort("uploadTime").order(SortOrder.DESC);
            case SORT_HOT:
                return SortBuilders.fieldSort("viewCount").order(SortOrder.DESC);
            case SORT_RELEVANCE:
            default:
                return SortBuilders.scoreSort().order(SortOrder.DESC);
        }
    }

    private HighlightBuilder buildHighlight() {
        HighlightBuilder highlightBuilder = new HighlightBuilder();

        HighlightBuilder.Field titleField = new HighlightBuilder.Field("title");
        titleField.preTags("<em>");
        titleField.postTags("</em>");
        highlightBuilder.field(titleField);

        HighlightBuilder.Field descField = new HighlightBuilder.Field("description");
        descField.preTags("<em>");
        descField.postTags("</em>");
        highlightBuilder.field(descField);

        HighlightBuilder.Field videoTitlesField = new HighlightBuilder.Field("videoTitles");
        videoTitlesField.preTags("<em>");
        videoTitlesField.postTags("</em>");
        highlightBuilder.field(videoTitlesField);

        return highlightBuilder;
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