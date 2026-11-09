package com.mybilibili.web.service.impl;

import com.mybilibili.common.document.ManuscriptDocument;
import com.mybilibili.common.vo.VideoSearchVO;
import com.mybilibili.web.service.HotSearchService;
import com.mybilibili.web.service.VideoSearchService;
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
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

/**
 * 视频搜索服务实现 - Elasticsearch版本
 * 以稿件为单位进行搜索
 */
@Slf4j
@Service
@ConditionalOnProperty(name = "spring.data.elasticsearch.repositories.enabled", havingValue = "true", matchIfMissing = false)
public class VideoSearchServiceImpl implements VideoSearchService {

    @Autowired
    private ElasticsearchRestTemplate elasticsearchRestTemplate;

    @Autowired
    private HotSearchService hotSearchService;

    /**
     * 已上架稿件状态
     */
    private static final Integer PUBLISHED_STATUS = 3;

    /**
     * 排序方式：相关度
     */
    private static final String SORT_RELEVANCE = "relevance";

    /**
     * 排序方式：时间
     */
    private static final String SORT_TIME = "time";

    /**
     * 排序方式：热度
     */
    private static final String SORT_HOT = "hot";

    @Override
    public Page<VideoSearchVO> search(String keyword, Integer categoryId, String tag, Integer userId,
                                      String sort, int page, int size) {
        try {
            // 构建布尔查询
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();

            // 1. 必须条件：只搜索已上架稿件（status=3)
            boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            // 2. 关键词搜索（title、description、videoTitles)
            if (StringUtils.hasText(keyword)) {
                BoolQueryBuilder keywordQuery = QueryBuilders.boolQuery();
                keywordQuery.should(QueryBuilders.matchQuery("title", keyword).boost(3.0f));
                keywordQuery.should(QueryBuilders.matchQuery("description", keyword).boost(1.5f));
                keywordQuery.should(QueryBuilders.matchQuery("videoTitles", keyword).boost(2.0f));
                keywordQuery.minimumShouldMatch(1);
                boolQuery.must(keywordQuery);
            }

            // 3. 分类过滤
            if (categoryId != null) {
                boolQuery.filter(QueryBuilders.termQuery("categoryId", categoryId));
            }

            // 4. 标签过滤
            if (StringUtils.hasText(tag)) {
                boolQuery.filter(QueryBuilders.termQuery("tags", tag));
            }

            // 5. UP主过滤
            if (userId != null) {
                boolQuery.filter(QueryBuilders.termQuery("userId", userId));
            }

            // 6. 更新热搜榜（异步更新，不影响搜索性能）
            if (StringUtils.hasText(keyword)) {
                try {
                    hotSearchService.incrementHotSearch(keyword);
                } catch (Exception e) {
                    log.warn("更新热搜榜失败，keyword: {}, error: {}", keyword, e.getMessage());
                }
            }

            // 构建排序
            SortBuilder<?> sortBuilder = buildSort(sort);

            // 构建高亮
            HighlightBuilder highlightBuilder = buildHighlight();

            // 构建分页（前端page从1开始，Spring Data从0开始）
            Pageable pageable = PageRequest.of(Math.max(0, page - 1), size);

            // 构建搜索查询
            NativeSearchQueryBuilder queryBuilder = new NativeSearchQueryBuilder()
                    .withQuery(boolQuery)
                    .withPageable(pageable)
                    .withHighlightBuilder(highlightBuilder);

            // 添加排序（如果不是按相关度排序）
            if (!SORT_RELEVANCE.equals(sort)) {
                queryBuilder.withSort(sortBuilder);
            }

            NativeSearchQuery searchQuery = queryBuilder.build();

            // 执行搜索
            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                    searchQuery, ManuscriptDocument.class);

            log.info("搜索命中 {} 条记录", searchHits.getTotalHits());
            for (SearchHit<ManuscriptDocument> hit : searchHits.getSearchHits()) {
                log.info("文档 ID: {}, title: {}", hit.getContent().getManuscriptId(), hit.getContent().getTitle());
            }

            // 转换为VO
            log.info("开始转换VO...");
            List<VideoSearchVO> voList = new ArrayList<>();
            for (SearchHit<ManuscriptDocument> hit : searchHits.getSearchHits()) {
                log.info("处理文档: {}", hit.getContent().getManuscriptId());
                VideoSearchVO vo = convertToVO(hit);
                log.info("转换结果: {}", vo != null ? vo.getManuscriptId() : "null");
                if (vo != null) {
                    voList.add(vo);
                }
            }

            log.info("转换后 VO 数量: {}", voList.size());

            // 构建分页结果
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
            // 使用前缀匹配获取建议
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

    /**
     * 构建排序
     *
     * @param sort 排序方式
     * @return 排序构建器
     */
    private SortBuilder<?> buildSort(String sort) {
        if (!StringUtils.hasText(sort)) {
            sort = SORT_RELEVANCE;
        }

        switch (sort) {
            case SORT_TIME:
                return SortBuilders.fieldSort("uploadTime").order(SortOrder.DESC);
            case SORT_HOT:
                // 热度排序：综合播放量
                return SortBuilders.fieldSort("viewCount").order(SortOrder.DESC);
            case SORT_RELEVANCE:
            default:
                // 相关度排序由ES默认处理，不需要额外指定
                return SortBuilders.scoreSort().order(SortOrder.DESC);
        }
    }

    /**
     * 构建高亮
     *
     * @return 高亮构建器
     */
    private HighlightBuilder buildHighlight() {
        HighlightBuilder highlightBuilder = new HighlightBuilder();

        // 标题高亮
        HighlightBuilder.Field titleField = new HighlightBuilder.Field("title");
        titleField.preTags("<em>");
        titleField.postTags("</em>");
        highlightBuilder.field(titleField);

        // 描述高亮
        HighlightBuilder.Field descField = new HighlightBuilder.Field("description");
        descField.preTags("<em>");
        descField.postTags("</em>");
        highlightBuilder.field(descField);

        // 视频标题高亮
        HighlightBuilder.Field videoTitlesField = new HighlightBuilder.Field("videoTitles");
        videoTitlesField.preTags("<em>");
        videoTitlesField.postTags("</em>");
        highlightBuilder.field(videoTitlesField);

        return highlightBuilder;
    }

    /**
     * 将搜索结果转换为VO
     *
     * @param searchHit 搜索结果
     * @return VideoSearchVO
     */
    private VideoSearchVO convertToVO(SearchHit<ManuscriptDocument> searchHit) {
        try {
            ManuscriptDocument document = searchHit.getContent();
            log.info("转换文档: manuscriptId={}, title={}", document.getManuscriptId(), document.getTitle());
            VideoSearchVO vo = new VideoSearchVO();

            // 设置基本信息 - 以稿件为单位
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

            // 设置视频信息
            vo.setVideoIds(document.getVideoIds());
            vo.setVideoCount(document.getVideoCount());

            log.info("转换成功: manuscriptId={}", document.getManuscriptId());
            return vo;
        } catch (Exception e) {
            log.error("转换VO失败: {}", e.getMessage(), e);
            return null;
        }
    }

    /**
     * 格式化视频时长
     *
     * @param seconds 秒数
     * @return 格式化字符串（如 05:30）
     */
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
