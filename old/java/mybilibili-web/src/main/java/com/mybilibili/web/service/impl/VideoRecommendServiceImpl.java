package com.mybilibili.web.service.impl;

import com.mybilibili.common.document.ManuscriptDocument;
import com.mybilibili.common.vo.VideoRecommendVO;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.web.service.VideoRecommendService;
import com.mybilibili.web.service.WatchHistoryService;
import lombok.extern.slf4j.Slf4j;
import org.elasticsearch.index.query.BoolQueryBuilder;
import org.elasticsearch.index.query.QueryBuilders;
import org.elasticsearch.search.sort.FieldSortBuilder;
import org.elasticsearch.search.sort.SortBuilders;
import org.elasticsearch.search.sort.SortOrder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.elasticsearch.core.ElasticsearchRestTemplate;
import org.springframework.data.elasticsearch.core.SearchHit;
import org.springframework.data.elasticsearch.core.SearchHits;
import org.springframework.data.elasticsearch.core.query.NativeSearchQuery;
import org.springframework.data.elasticsearch.core.query.NativeSearchQueryBuilder;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

/**
 * 视频推荐服务实现 - Elasticsearch版本
 * 以稿件为单位进行推荐
 */
@Slf4j
@Service
@ConditionalOnProperty(name = "spring.data.elasticsearch.repositories.enabled", havingValue = "true", matchIfMissing = false)
public class VideoRecommendServiceImpl implements VideoRecommendService {

    @Autowired
    private ElasticsearchRestTemplate elasticsearchRestTemplate;

    @Autowired
    private WatchHistoryService watchHistoryService;

    /**
     * 已上架稿件状态
     */
    private static final Integer PUBLISHED_STATUS = 3;

    /**
     * 默认推荐数量
     */
    private static final int DEFAULT_SIZE = 10;

    /**
     * 最大推荐数量
     */
    private static final int MAX_SIZE = 50;

    @Override
    public List<VideoRecommendVO> getRelatedVideos(Integer videoId, int size) {
        if (videoId == null) {
            return new ArrayList<>();
        }

        // 限制size范围
        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            // 1. 先获取当前稿件的信息
            ManuscriptDocument currentManuscript = getManuscriptByVideoId(videoId);
            if (currentManuscript == null) {
                log.warn("未找到视频对应的稿件: {}", videoId);
                return new ArrayList<>();
            }

            // 2. 构建 More Like This 查询
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();

            // 必须条件：只搜索已上架稿件
            boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            // 排除当前稿件
            boolQuery.mustNot(QueryBuilders.termQuery("manuscriptId", currentManuscript.getManuscriptId()));

            // 3. 构建相似度查询（基于标题、描述、标签、视频标题）
            BoolQueryBuilder similarityQuery = QueryBuilders.boolQuery();

            // 标题相似度
            if (hasText(currentManuscript.getTitle())) {
                similarityQuery.should(
                    QueryBuilders.matchQuery("title", currentManuscript.getTitle()).boost(3.0f)
                );
            }

            // 描述相似度
            if (hasText(currentManuscript.getDescription())) {
                similarityQuery.should(
                    QueryBuilders.matchQuery("description", currentManuscript.getDescription()).boost(1.5f)
                );
            }

            // 视频标题相似度
            if (hasText(currentManuscript.getVideoTitles())) {
                similarityQuery.should(
                    QueryBuilders.matchQuery("videoTitles", currentManuscript.getVideoTitles()).boost(2.0f)
                );
            }

            // 标签相似度
            if (hasTags(currentManuscript.getTags())) {
                for (String tag : currentManuscript.getTags()) {
                    similarityQuery.should(
                        QueryBuilders.termQuery("tags", tag).boost(2.0f)
                    );
                }
            }

            // 同分类视频加分
            if (currentManuscript.getCategoryId() != null) {
                similarityQuery.should(
                    QueryBuilders.termQuery("categoryId", currentManuscript.getCategoryId()).boost(1.0f)
                );
            }

            boolQuery.must(similarityQuery);

            // 4. 执行搜索
            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                .withQuery(boolQuery)
                .withPageable(PageRequest.of(0, size))
                .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

            // 5. 转换为VO并设置推荐理由
            return searchHits.getSearchHits().stream()
                .map(hit -> convertToVO(hit, currentManuscript))
                .collect(Collectors.toList());

        } catch (Exception e) {
            log.error("获取相关视频失败，videoId: {}, error: {}", videoId, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    @Override
    public List<VideoRecommendVO> getHotVideos(Integer categoryId, int size) {
        // 限制size范围
        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            // 1. 构建查询
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();

            // 必须条件：只搜索已上架稿件
            boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            // 分类过滤
            if (categoryId != null) {
                boolQuery.filter(QueryBuilders.termQuery("categoryId", categoryId));
            }

            // 2. 按播放量排序
            FieldSortBuilder sortBuilder = SortBuilders
                .fieldSort("viewCount")
                .order(SortOrder.DESC);

            // 3. 执行搜索
            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                .withQuery(boolQuery)
                .withSort(sortBuilder)
                .withPageable(PageRequest.of(0, size))
                .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

            // 4. 转换为VO并设置推荐理由
            int rank = 1;
            List<VideoRecommendVO> result = new ArrayList<>();
            for (SearchHit<ManuscriptDocument> hit : searchHits.getSearchHits()) {
                VideoRecommendVO vo = convertToVO(hit);
                vo.setRecommendReason("热门排行第" + rank + "名");
                vo.setScore((double) hit.getScore());
                result.add(vo);
                rank++;
            }

            return result;

        } catch (Exception e) {
            log.error("获取热门视频失败，categoryId: {}, error: {}", categoryId, e.getMessage(), e);
            return new ArrayList<>();
        }
    }

    @Override
    public List<VideoRecommendVO> getRecommendedVideosForUser(Integer userId, int size) {
        if (userId == null) {
            return new ArrayList<>();
        }

        // 限制size范围
        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            // 1. 获取用户浏览历史
            List<WatchHistoryVO> historyList = watchHistoryService.getWatchHistoryList(userId, 0, 20);

            if (historyList == null || historyList.isEmpty()) {
                // 没有浏览历史，返回热门视频
                return getHotVideos(null, size);
            }

            // 2. 提取用户感兴趣的标签和分类
            Set<String> interestedTags = new HashSet<>();
            Set<Integer> interestedCategories = new HashSet<>();
            Set<Integer> watchedManuscriptIds = new HashSet<>();

            for (WatchHistoryVO history : historyList) {
                if (history.getVideo() != null) {
                    VideoVO video = history.getVideo();
                    // 从视频信息中获取稿件ID
                    if (video.getManuscriptId() != null) {
                        watchedManuscriptIds.add(video.getManuscriptId());
                    }
                    if (video.getTags() != null) {
                        interestedTags.addAll(video.getTags());
                    }
                    if (video.getCategoryId() != null) {
                        interestedCategories.add(video.getCategoryId());
                    }
                }
            }

            // 3. 构建个性化推荐查询
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();

            // 必须条件：只搜索已上架稿件
            boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            // 排除已观看的稿件
            for (Integer manuscriptId : watchedManuscriptIds) {
                boolQuery.mustNot(QueryBuilders.termQuery("manuscriptId", manuscriptId));
            }

            // 4. 构建兴趣匹配查询
            BoolQueryBuilder interestQuery = QueryBuilders.boolQuery();

            // 标签匹配
            for (String tag : interestedTags) {
                if (hasText(tag)) {
                    interestQuery.should(
                        QueryBuilders.termQuery("tags", tag).boost(2.0f)
                    );
                }
            }

            // 分类匹配
            for (Integer categoryId : interestedCategories) {
                if (categoryId != null) {
                    interestQuery.should(
                        QueryBuilders.termQuery("categoryId", categoryId).boost(1.5f)
                    );
                }
            }

            // 热度因子（播放量）
            interestQuery.should(
                QueryBuilders.rangeQuery("viewCount").gt(1000).boost(0.5f)
            );

            boolQuery.must(interestQuery);

            // 5. 执行搜索
            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                .withQuery(boolQuery)
                .withPageable(PageRequest.of(0, size))
                .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

            // 6. 转换为VO并设置推荐理由
            return searchHits.getSearchHits().stream()
                .map(hit -> {
                    VideoRecommendVO vo = convertToVO(hit);
                    vo.setRecommendReason("猜你喜欢");
                    vo.setScore((double) hit.getScore());
                    return vo;
                })
                .collect(Collectors.toList());

        } catch (Exception e) {
            log.error("获取个性化推荐失败，userId: {}, error: {}", userId, e.getMessage(), e);
            // 出错时返回热门视频作为兜底
            return getHotVideos(null, size);
        }
    }

    /**
     * 根据视频ID获取对应的稿件文档
     *
     * @param videoId 视频ID
     * @return ManuscriptDocument
     */
    private ManuscriptDocument getManuscriptByVideoId(Integer videoId) {
        try {
            // 通过videoIds字段查询包含该视频的稿件
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery()
                .must(QueryBuilders.termQuery("status", PUBLISHED_STATUS))
                .must(QueryBuilders.termQuery("videoIds", videoId));

            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                .withQuery(boolQuery)
                .withPageable(PageRequest.of(0, 1))
                .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

            if (searchHits.hasSearchHits()) {
                return searchHits.getSearchHit(0).getContent();
            }
            return null;
        } catch (Exception e) {
            log.error("获取稿件信息失败，videoId: {}, error: {}", videoId, e.getMessage());
            return null;
        }
    }

    /**
     * 转换为VO（带推荐理由）
     *
     * @param searchHit          搜索结果
     * @param currentManuscript  当前稿件（用于生成推荐理由）
     * @return VideoRecommendVO
     */
    private VideoRecommendVO convertToVO(SearchHit<ManuscriptDocument> searchHit, ManuscriptDocument currentManuscript) {
        VideoRecommendVO vo = convertToVO(searchHit);

        // 生成推荐理由
        ManuscriptDocument document = searchHit.getContent();
        String reason = generateRecommendReason(document, currentManuscript);
        vo.setRecommendReason(reason);

        return vo;
    }

    /**
     * 转换为VO
     *
     * @param searchHit 搜索结果
     * @return VideoRecommendVO
     */
    private VideoRecommendVO convertToVO(SearchHit<ManuscriptDocument> searchHit) {
        ManuscriptDocument document = searchHit.getContent();
        VideoRecommendVO vo = new VideoRecommendVO();

        // 以稿件为单位返回，取第一个视频ID作为代表
        if (document.getVideoIds() != null && !document.getVideoIds().isEmpty()) {
            vo.setVideoId(document.getVideoIds().get(0));
        }
        vo.setManuscriptId(document.getManuscriptId());
        vo.setTitle(document.getTitle());
        vo.setDescription(document.getDescription());
        vo.setCoverUrl(document.getCoverUrl());
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
        vo.setScore((double) searchHit.getScore());

        return vo;
    }

    /**
     * 生成推荐理由
     *
     * @param document          推荐稿件
     * @param currentManuscript 当前稿件
     * @return 推荐理由
     */
    private String generateRecommendReason(ManuscriptDocument document, ManuscriptDocument currentManuscript) {
        // 检查标签匹配
        if (hasTags(document.getTags()) && hasTags(currentManuscript.getTags())) {
            Set<String> commonTags = new HashSet<>(document.getTags());
            commonTags.retainAll(currentManuscript.getTags());
            if (!commonTags.isEmpty()) {
                return "相似标签: " + String.join(", ", commonTags.stream().limit(3).collect(Collectors.toList()));
            }
        }

        // 检查分类匹配
        if (document.getCategoryId() != null &&
            document.getCategoryId().equals(currentManuscript.getCategoryId())) {
            return "同分类视频";
        }

        return "相关推荐";
    }

    /**
     * 检查文本是否有内容
     *
     * @param text 文本
     * @return 是否有内容
     */
    private boolean hasText(String text) {
        return text != null && !text.trim().isEmpty();
    }

    /**
     * 检查是否有标签
     *
     * @param tags 标签列表
     * @return 是否有标签
     */
    private boolean hasTags(List<String> tags) {
        return tags != null && !tags.isEmpty();
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
