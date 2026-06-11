package com.mybilibili.search.service.impl;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoRecommendVO;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.common.document.ManuscriptDocument;
import com.mybilibili.search.entity.RecommendConfig;
import com.mybilibili.search.feign.UserProfileClient;
import com.mybilibili.search.service.VideoRecommendService;
import lombok.extern.slf4j.Slf4j;
import co.elastic.clients.elasticsearch._types.SortOptions;
import co.elastic.clients.elasticsearch._types.SortOrder;
import co.elastic.clients.elasticsearch._types.query_dsl.Query;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.elasticsearch.client.elc.ElasticsearchTemplate;
import org.springframework.data.elasticsearch.client.elc.NativeQuery;
import org.springframework.data.elasticsearch.client.elc.NativeQueryBuilder;
import org.springframework.data.elasticsearch.core.SearchHit;
import org.springframework.data.elasticsearch.core.SearchHits;
import org.springframework.stereotype.Service;

import jakarta.servlet.http.HttpServletRequest;
import java.util.*;
import java.util.stream.Collectors;

@Slf4j
@Service
public class VideoRecommendServiceImpl implements VideoRecommendService {

    @Autowired
    private ElasticsearchTemplate elasticsearchRestTemplate;

    @Autowired
    private UserProfileClient userProfileClient;

    @Autowired
    private com.mybilibili.search.service.RecommendConfigService recommendConfigService;

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

            // Build similarity sub-query
            List<Query> similarityClauses = new ArrayList<>();

            if (hasText(currentManuscript.getTitle())) {
                similarityClauses.add(Query.of(q -> q.match(m ->
                    m.field("title").query(currentManuscript.getTitle()).boost(3.0f))));
            }

            if (hasText(currentManuscript.getDescription())) {
                similarityClauses.add(Query.of(q -> q.match(m ->
                    m.field("description").query(currentManuscript.getDescription()).boost(1.5f))));
            }

            if (hasText(currentManuscript.getVideoTitles())) {
                similarityClauses.add(Query.of(q -> q.match(m ->
                    m.field("videoTitles").query(currentManuscript.getVideoTitles()).boost(2.0f))));
            }

            if (hasTags(currentManuscript.getTags())) {
                for (String tag : currentManuscript.getTags()) {
                    similarityClauses.add(Query.of(q -> q.term(t ->
                        t.field("tags").value(tag).boost(2.0f))));
                }
            }

            if (currentManuscript.getCategoryId() != null) {
                similarityClauses.add(Query.of(q -> q.term(t ->
                    t.field("categoryId").value(currentManuscript.getCategoryId().toString()).boost(1.0f))));
            }

            Query similarityQuery = Query.of(q -> q.bool(b -> {
                for (Query clause : similarityClauses) {
                    b.should(clause);
                }
                return b;
            }));

            // Build main query
            NativeQuery searchQuery = new NativeQueryBuilder()
                .withQuery(q -> q.bool(b -> b
                    .must(m -> m.term(t -> t.field("status").value(PUBLISHED_STATUS)))
                    .mustNot(m -> m.term(t -> t.field("manuscriptId").value(currentManuscript.getManuscriptId().toString())))
                    .must(similarityQuery)
                ))
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
            RecommendConfig rcfg = recommendConfigService.getConfig();
            int candidateSize = calculateCandidateSize(size, rcfg);
            double randomWeight = clampDouble(rcfg.getHotRandomWeight(), 0.0, 1.0, 0.0);

            NativeQueryBuilder queryBuilder = new NativeQueryBuilder()
                .withQuery(q -> q.bool(b -> {
                    b.must(m -> m.term(t -> t.field("status").value(PUBLISHED_STATUS)));
                    if (categoryId != null) {
                        b.filter(f -> f.term(t -> t.field("categoryId").value(categoryId.toString())));
                    }
                    return b;
                }))
                .withSort(SortOptions.of(s -> s.field(f -> f.field("viewCount").order(SortOrder.Desc))))
                .withPageable(PageRequest.of(0, candidateSize));

            NativeQuery searchQuery = queryBuilder.build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                searchQuery, ManuscriptDocument.class);

            List<SearchHit<ManuscriptDocument>> rankedHits = selectCandidateHits(
                    searchHits.getSearchHits(),
                    size,
                    randomWeight,
                    rcfg.getShuffleWindowSize()
            );

            int rank = 1;
            List<VideoRecommendVO> result = new ArrayList<>();
            for (SearchHit<ManuscriptDocument> hit : rankedHits) {
                VideoRecommendVO vo = convertToVO(hit);
                vo.setRecommendReason(randomWeight > 0 ? "热门推荐" : "热门排行第" + rank + "名");
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
            RecommendConfig rcfg = recommendConfigService.getConfig();
            Map<String, Object> profile = fetchUserProfile(userId);
            if (profile == null || profile.isEmpty()) {
                return getHotVideos(null, size);
            }

            @SuppressWarnings("unchecked")
            Map<String, Number> categoryWeights = (Map<String, Number>) profile.get("categoryWeights");
            @SuppressWarnings("unchecked")
            Map<String, Number> tagWeights = (Map<String, Number>) profile.get("tagWeights");

            if ((categoryWeights == null || categoryWeights.isEmpty()) &&
                (tagWeights == null || tagWeights.isEmpty())) {
                return getHotVideos(null, size);
            }

            List<Query> profileClauses = new ArrayList<>();
            int candidateSize = calculateCandidateSize(size, rcfg);
            double randomWeight = clampDouble(rcfg.getPersonalizedRandomWeight(), 0.0, 1.0, 0.0);

            if (categoryWeights != null) {
                List<Map.Entry<String, Number>> topCategories = categoryWeights.entrySet().stream()
                        .sorted(Map.Entry.<String, Number>comparingByValue(
                                Comparator.comparingDouble(Number::doubleValue)).reversed())
                        .limit(rcfg.getTopCategoryCount() != null ? rcfg.getTopCategoryCount() : 5)
                        .toList();
                for (Map.Entry<String, Number> entry : topCategories) {
                    float boost = Math.max(1.0f, entry.getValue().floatValue() * (rcfg.getCategoryBoost() != null ? rcfg.getCategoryBoost() : 1.0f));
                    String catId = entry.getKey();
                    profileClauses.add(Query.of(q -> q.term(t ->
                            t.field("categoryId").value(catId).boost(boost))));
                }
            }

            if (tagWeights != null) {
                List<Map.Entry<String, Number>> topTags = tagWeights.entrySet().stream()
                        .sorted(Map.Entry.<String, Number>comparingByValue(
                                Comparator.comparingDouble(Number::doubleValue)).reversed())
                        .limit(rcfg.getTopTagCount() != null ? rcfg.getTopTagCount() : 10)
                        .toList();
                for (Map.Entry<String, Number> entry : topTags) {
                    float boost = Math.max(1.0f, entry.getValue().floatValue() * (rcfg.getTagBoost() != null ? rcfg.getTagBoost() : 0.5f));
                    String tag = entry.getKey();
                    profileClauses.add(Query.of(q -> q.term(t ->
                            t.field("tags").value(tag).boost(boost))));
                }
            }

            Query profileQuery = Query.of(q -> q.bool(b -> {
                for (Query clause : profileClauses) {
                    b.should(clause);
                }
                return b;
            }));

            NativeQuery searchQuery = new NativeQueryBuilder()
                    .withQuery(q -> q.bool(b -> b
                            .must(m -> m.term(t -> t.field("status").value(PUBLISHED_STATUS)))
                            .should(profileQuery)
                    ))
                    .withSort(SortOptions.of(s -> s.score(sc -> sc.order(SortOrder.Desc))))
                    .withPageable(PageRequest.of(0, candidateSize))
                    .build();

            SearchHits<ManuscriptDocument> searchHits = elasticsearchRestTemplate.search(
                    searchQuery, ManuscriptDocument.class);

            List<SearchHit<ManuscriptDocument>> rankedHits = selectCandidateHits(
                    searchHits.getSearchHits(),
                    size,
                    randomWeight,
                    rcfg.getShuffleWindowSize()
            );

            List<VideoRecommendVO> result = new ArrayList<>();
            for (SearchHit<ManuscriptDocument> hit : rankedHits) {
                VideoRecommendVO vo = convertToVO(hit);
                vo.setRecommendReason(generateProfileRecommendReason(hit.getContent(), categoryWeights, tagWeights));
                vo.setScore((double) hit.getScore());
                result.add(vo);
            }

            if (result.size() < size) {
                List<VideoRecommendVO> hotFill = getHotVideos(null, size - result.size());
                Set<Integer> existingIds = result.stream()
                        .map(VideoRecommendVO::getManuscriptId).collect(Collectors.toSet());
                for (VideoRecommendVO hot : hotFill) {
                    if (!existingIds.contains(hot.getManuscriptId())) {
                        result.add(hot);
                    }
                }
            }

            return result;
        } catch (Exception e) {
            log.error("获取个性化推荐失败，userId: {}, error: {}", userId, e.getMessage(), e);
            return getHotVideos(null, size);
        }
    }

    private Map<String, Object> fetchUserProfile(Integer userId) {
        try {
            Result<Map<String, Object>> result = userProfileClient.getProfile(userId);
            if (result != null && result.getCode() == 200 && result.getData() != null) {
                return result.getData();
            }
        } catch (Exception e) {
            log.debug("获取用户画像失败: {}", e.getMessage());
        }
        return null;
    }

    private String generateProfileRecommendReason(ManuscriptDocument doc,
                                                   Map<String, Number> categoryWeights,
                                                   Map<String, Number> tagWeights) {
        if (tagWeights != null && hasTags(doc.getTags())) {
            for (String tag : doc.getTags()) {
                if (tagWeights.containsKey(tag)) {
                    return "你可能感兴趣: " + tag;
                }
            }
        }
        if (categoryWeights != null && doc.getCategoryId() != null) {
            if (categoryWeights.containsKey(doc.getCategoryId().toString())) {
                return "基于你常看的" + (doc.getCategoryName() != null ? doc.getCategoryName() : "分类");
            }
        }
        return "为你推荐";
    }

    private ManuscriptDocument getManuscriptByVideoId(Integer videoId) {
        try {
            NativeQuery searchQuery = new NativeQueryBuilder()
                .withQuery(q -> q.bool(b -> b
                    .must(m -> m.term(t -> t.field("status").value(PUBLISHED_STATUS)))
                    .must(m -> m.term(t -> t.field("videoIds").value(videoId.toString())))
                ))
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
        vo.setUserAvatar(document.getUserAvatar());
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

    private int calculateCandidateSize(int size, RecommendConfig config) {
        int multiplier = clampInt(config != null ? config.getCandidateMultiplier() : null, 1, 10, 1);
        return Math.min(MAX_SIZE, Math.max(size, size * multiplier));
    }

    private List<SearchHit<ManuscriptDocument>> selectCandidateHits(List<SearchHit<ManuscriptDocument>> hits,
                                                                    int size,
                                                                    double randomWeight,
                                                                    Integer shuffleWindowSize) {
        if (hits == null || hits.isEmpty()) {
            return Collections.emptyList();
        }
        if (randomWeight <= 0) {
            return new ArrayList<>(hits.subList(0, Math.min(size, hits.size())));
        }

        int windowSize = clampInt(shuffleWindowSize, size, MAX_SIZE, Math.max(size, 10));
        int window = Math.min(hits.size(), windowSize);
        List<RankedHit> ranked = new ArrayList<>();

        for (int i = 0; i < window; i++) {
            double baseScore = window <= 1 ? 1.0 : 1.0 - ((double) i / (window - 1));
            double finalScore = baseScore * (1.0 - randomWeight) + Math.random() * randomWeight;
            ranked.add(new RankedHit(hits.get(i), finalScore));
        }

        ranked.sort(Comparator.comparingDouble(RankedHit::score).reversed());

        List<SearchHit<ManuscriptDocument>> result = ranked.stream()
                .limit(size)
                .map(RankedHit::hit)
                .collect(Collectors.toCollection(ArrayList::new));

        for (int i = window; result.size() < size && i < hits.size(); i++) {
            result.add(hits.get(i));
        }
        return result;
    }

    private int clampInt(Integer value, int min, int max, int fallback) {
        int actual = value != null ? value : fallback;
        return Math.max(min, Math.min(max, actual));
    }

    private double clampDouble(Double value, double min, double max, double fallback) {
        double actual = value != null ? value : fallback;
        return Math.max(min, Math.min(max, actual));
    }

    private static class RankedHit {
        private final SearchHit<ManuscriptDocument> hit;
        private final double score;

        private RankedHit(SearchHit<ManuscriptDocument> hit, double score) {
            this.hit = hit;
            this.score = score;
        }

        private SearchHit<ManuscriptDocument> hit() {
            return hit;
        }

        private double score() {
            return score;
        }
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
