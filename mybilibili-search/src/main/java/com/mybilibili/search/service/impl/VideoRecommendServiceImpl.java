package com.mybilibili.search.service.impl;

import com.mybilibili.common.document.ManuscriptDocument;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.VideoRecommendVO;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.search.service.VideoRecommendService;
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
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpServletRequest;
import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class VideoRecommendServiceImpl implements VideoRecommendService {

    @Autowired
    private ElasticsearchRestTemplate elasticsearchRestTemplate;

    private static final Integer PUBLISHED_STATUS = 3;
    private static final int DEFAULT_SIZE = 10;
    private static final int MAX_SIZE = 50;

    @Override
    public List<VideoRecommendVO> getRelatedVideos(Integer videoId, int size) {
        if (videoId == null) {
            return new ArrayList<>();
        }

        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            ManuscriptDocument currentManuscript = getManuscriptByVideoId(videoId);
            if (currentManuscript == null) {
                log.warn("未找到视频对应的稿件: {}", videoId);
                return new ArrayList<>();
            }

            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();
            boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));
            boolQuery.mustNot(QueryBuilders.termQuery("manuscriptId", currentManuscript.getManuscriptId()));

            BoolQueryBuilder similarityQuery = QueryBuilders.boolQuery();

            if (hasText(currentManuscript.getTitle())) {
                similarityQuery.should(
                    QueryBuilders.matchQuery("title", currentManuscript.getTitle()).boost(3.0f)
                );
            }

            if (hasText(currentManuscript.getDescription())) {
                similarityQuery.should(
                    QueryBuilders.matchQuery("description", currentManuscript.getDescription()).boost(1.5f)
                );
            }

            if (hasText(currentManuscript.getVideoTitles())) {
                similarityQuery.should(
                    QueryBuilders.matchQuery("videoTitles", currentManuscript.getVideoTitles()).boost(2.0f)
                );
            }

            if (hasTags(currentManuscript.getTags())) {
                for (String tag : currentManuscript.getTags()) {
                    similarityQuery.should(
                        QueryBuilders.termQuery("tags", tag).boost(2.0f)
                    );
                }
            }

            if (currentManuscript.getCategoryId() != null) {
                similarityQuery.should(
                    QueryBuilders.termQuery("categoryId", currentManuscript.getCategoryId()).boost(1.0f)
                );
            }

            boolQuery.must(similarityQuery);

            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                .withQuery(boolQuery)
                .withPageable(PageRequest.of(0, size))
                .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

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
        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            BoolQueryBuilder boolQuery = QueryBuilders.boolQuery();
            boolQuery.must(QueryBuilders.termQuery("status", PUBLISHED_STATUS));

            if (categoryId != null) {
                boolQuery.filter(QueryBuilders.termQuery("categoryId", categoryId));
            }

            FieldSortBuilder sortBuilder = SortBuilders
                .fieldSort("viewCount")
                .order(SortOrder.DESC);

            NativeSearchQuery searchQuery = new NativeSearchQueryBuilder()
                .withQuery(boolQuery)
                .withSort(sortBuilder)
                .withPageable(PageRequest.of(0, size))
                .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

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

        size = Math.max(1, Math.min(size, MAX_SIZE));

        try {
            return getHotVideos(null, size);
        } catch (Exception e) {
            log.error("获取个性化推荐失败，userId: {}, error: {}", userId, e.getMessage(), e);
            return getHotVideos(null, size);
        }
    }

    private ManuscriptDocument getManuscriptByVideoId(Integer videoId) {
        try {
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

    private VideoRecommendVO convertToVO(SearchHit<ManuscriptDocument> searchHit, ManuscriptDocument currentManuscript) {
        VideoRecommendVO vo = convertToVO(searchHit);
        ManuscriptDocument document = searchHit.getContent();
        String reason = generateRecommendReason(document, currentManuscript);
        vo.setRecommendReason(reason);
        return vo;
    }

    private VideoRecommendVO convertToVO(SearchHit<ManuscriptDocument> searchHit) {
        ManuscriptDocument document = searchHit.getContent();
        VideoRecommendVO vo = new VideoRecommendVO();

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

    private String generateRecommendReason(ManuscriptDocument document, ManuscriptDocument currentManuscript) {
        if (hasTags(document.getTags()) && hasTags(currentManuscript.getTags())) {
            Set<String> commonTags = new HashSet<>(document.getTags());
            commonTags.retainAll(currentManuscript.getTags());
            if (!commonTags.isEmpty()) {
                return "相似标签: " + String.join(", ", commonTags.stream().limit(3).collect(Collectors.toList()));
            }
        }

        if (document.getCategoryId() != null &&
            document.getCategoryId().equals(currentManuscript.getCategoryId())) {
            return "同分类视频";
        }

        return "相关推荐";
    }

    private boolean hasText(String text) {
        return text != null && !text.trim().isEmpty();
    }

    private boolean hasTags(List<String> tags) {
        return tags != null && !tags.isEmpty();
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