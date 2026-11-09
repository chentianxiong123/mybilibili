package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Video;
import com.mybilibili.web.mapper.ManuscriptMapper;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.LikeHistoryService;
import com.mybilibili.web.service.RandomRecommendService;
import com.mybilibili.web.service.VideoTagService;
import com.mybilibili.web.service.WatchHistoryService;
import lombok.Data;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.stream.Collectors;

/**
 * 随机推荐服务实现
 * 基于加权随机算法，结合用户历史行为和热度进行推荐
 */
@Slf4j
@Service
public class RandomRecommendServiceImpl implements RandomRecommendService {

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private WatchHistoryService watchHistoryService;

    @Autowired
    private LikeHistoryService likeHistoryService;

    @Autowired
    private VideoTagService videoTagService;

    /**
     * 已上架稿件状态
     */
    private static final Integer PUBLISHED_STATUS = 3;

    /**
     * 基础分
     */
    private static final double BASE_SCORE = 10.0;

    /**
     * 兴趣分封顶值
     */
    private static final double MAX_INTEREST_SCORE = 50.0;

    /**
     * 热度分封顶值
     */
    private static final double MAX_HOT_SCORE = 20.0;

    /**
     * 随机分封顶值
     */
    private static final double MAX_RANDOM_SCORE = 30.0;

    /**
     * 浏览历史分类匹配加分
     */
    private static final double WATCH_CATEGORY_SCORE = 5.0;

    /**
     * 浏览历史标签匹配加分
     */
    private static final double WATCH_TAG_SCORE = 3.0;

    /**
     * 点赞历史分类匹配加分
     */
    private static final double LIKE_CATEGORY_SCORE = 10.0;

    /**
     * 点赞历史标签匹配加分
     */
    private static final double LIKE_TAG_SCORE = 6.0;

    /**
     * 热度分除数（每1000播放算1分）
     */
    private static final double HOT_SCORE_DIVISOR = 1000.0;

    @Override
    public List<Manuscript> getRandomRecommendedManuscripts(Integer userId, int size) {
        // 1. 获取所有已上架稿件作为候选集
        List<Manuscript> candidates = manuscriptMapper.selectByStatus(PUBLISHED_STATUS);

        if (candidates.isEmpty()) {
            log.warn("没有已上架的稿件可供推荐");
            return Collections.emptyList();
        }

        // 2. 获取用户行为数据（如果用户已登录）
        UserBehavior behavior = null;
        if (userId != null) {
            behavior = getUserBehavior(userId);
        }

        // 3. 为每个候选计算权重（同时加载标签信息）
        List<WeightedManuscript> weightedList = new ArrayList<>();
        for (Manuscript manuscript : candidates) {
            // 加载稿件的标签信息
            List<String> tags = loadManuscriptTags(manuscript.getId());
            manuscript.setTags(tags);

            double weight = calculateWeight(manuscript, behavior);
            weightedList.add(new WeightedManuscript(manuscript, weight));
        }

        // 4. 加权随机排序
        List<Manuscript> result = weightedRandomSort(weightedList, size);

        log.info("为用户 {} 生成推荐列表，候选集大小: {}, 推荐数量: {}",
                userId != null ? userId : "匿名", candidates.size(), result.size());

        return result;
    }

    /**
     * 加载稿件的标签信息（从Redis）
     *
     * @param manuscriptId 稿件ID
     * @return 标签列表
     */
    private List<String> loadManuscriptTags(Integer manuscriptId) {
        try {
            // 获取稿件的第一个视频
            List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
            if (videos.isEmpty()) {
                return Collections.emptyList();
            }

            // 从Redis获取第一个视频的标签
            Video firstVideo = videos.get(0);
            return videoTagService.getVideoTags(firstVideo.getId());
        } catch (Exception e) {
            log.warn("加载稿件标签失败，manuscriptId: {}", manuscriptId, e);
            return Collections.emptyList();
        }
    }

    /**
     * 获取用户行为数据（从Redis）
     *
     * @param userId 用户ID
     * @return 用户行为数据
     */
    private UserBehavior getUserBehavior(Integer userId) {
        UserBehavior behavior = new UserBehavior();

        // 获取最近7天的浏览历史（从Redis）
        List<Integer> watchedVideoIds = watchHistoryService.getRecentWatchVideoIds(userId, 7, 20);
        List<ManuscriptBehaviorInfo> watchHistory = buildWatchHistoryFromVideoIds(watchedVideoIds);
        behavior.setWatchHistory(watchHistory);

        // 获取最近7天的点赞记录（从Redis）
        List<Integer> likedVideoIds = likeHistoryService.getUserLikeHistory(userId, 7, 20);
        List<ManuscriptBehaviorInfo> likeHistory = buildLikeHistoryFromVideoIds(likedVideoIds);
        behavior.setLikeHistory(likeHistory);

        return behavior;
    }

    /**
     * 根据视频ID列表构建浏览历史
     *
     * @param videoIds 视频ID列表
     * @return 稿件行为信息列表
     */
    private List<ManuscriptBehaviorInfo> buildWatchHistoryFromVideoIds(List<Integer> videoIds) {
        if (videoIds == null || videoIds.isEmpty()) {
            return Collections.emptyList();
        }

        List<ManuscriptBehaviorInfo> result = new ArrayList<>();
        Map<Integer, ManuscriptBehaviorInfo> manuscriptMap = new HashMap<>();

        for (Integer videoId : videoIds) {
            try {
                Video video = videoMapper.selectById(videoId);
                if (video != null && video.getManuscriptId() != null) {
                    ManuscriptBehaviorInfo info = manuscriptMap.get(video.getManuscriptId());
                    if (info == null) {
                        info = new ManuscriptBehaviorInfo();
                        info.setManuscriptId(video.getManuscriptId());
                        info.setCategoryId(video.getCategoryId());
                        info.setTags(videoTagService.getVideoTags(videoId));
                        manuscriptMap.put(video.getManuscriptId(), info);
                    }
                }
            } catch (Exception e) {
                log.warn("构建浏览历史失败，videoId: {}", videoId, e);
            }
        }

        result.addAll(manuscriptMap.values());
        return result;
    }

    /**
     * 根据视频ID列表构建点赞历史
     *
     * @param videoIds 视频ID列表
     * @return 稿件行为信息列表
     */
    private List<ManuscriptBehaviorInfo> buildLikeHistoryFromVideoIds(List<Integer> videoIds) {
        if (videoIds == null || videoIds.isEmpty()) {
            return Collections.emptyList();
        }

        List<ManuscriptBehaviorInfo> result = new ArrayList<>();
        Map<Integer, ManuscriptBehaviorInfo> manuscriptMap = new HashMap<>();

        for (Integer videoId : videoIds) {
            try {
                Video video = videoMapper.selectById(videoId);
                if (video != null && video.getManuscriptId() != null) {
                    ManuscriptBehaviorInfo info = manuscriptMap.get(video.getManuscriptId());
                    if (info == null) {
                        info = new ManuscriptBehaviorInfo();
                        info.setManuscriptId(video.getManuscriptId());
                        info.setCategoryId(video.getCategoryId());
                        info.setTags(videoTagService.getVideoTags(videoId));
                        manuscriptMap.put(video.getManuscriptId(), info);
                    }
                }
            } catch (Exception e) {
                log.warn("构建点赞历史失败，videoId: {}", videoId, e);
            }
        }

        result.addAll(manuscriptMap.values());
        return result;
    }

    /**
     * 计算稿件推荐权重
     *
     * @param manuscript 稿件
     * @param behavior   用户行为数据（可能为null）
     * @return 权重值
     */
    private double calculateWeight(Manuscript manuscript, UserBehavior behavior) {
        double weight = BASE_SCORE;

        // 兴趣分（仅对登录用户计算）
        if (behavior != null) {
            weight += calculateInterestScore(manuscript, behavior);
        }

        // 热度分
        weight += calculateHotScore(manuscript);

        // 随机分
        weight += calculateRandomScore();

        return weight;
    }

    /**
     * 计算兴趣分
     *
     * @param manuscript 候选稿件
     * @param behavior   用户行为数据
     * @return 兴趣分（0-50）
     */
    private double calculateInterestScore(Manuscript manuscript, UserBehavior behavior) {
        double score = 0.0;

        // 基于浏览历史的兴趣
        for (ManuscriptBehaviorInfo watched : behavior.getWatchHistory()) {
            score += calculateBehaviorMatchScore(manuscript, watched,
                    WATCH_CATEGORY_SCORE, WATCH_TAG_SCORE);
        }

        // 基于点赞历史的兴趣（权重更高）
        for (ManuscriptBehaviorInfo liked : behavior.getLikeHistory()) {
            score += calculateBehaviorMatchScore(manuscript, liked,
                    LIKE_CATEGORY_SCORE, LIKE_TAG_SCORE);
        }

        // 封顶
        return Math.min(score, MAX_INTEREST_SCORE);
    }

    /**
     * 计算行为匹配分数
     *
     * @param candidate         候选稿件
     * @param behaviorManuscript 历史行为稿件
     * @param categoryScore     分类匹配分数
     * @param tagScore          标签匹配分数
     * @return 匹配分数
     */
    private double calculateBehaviorMatchScore(Manuscript candidate,
                                                ManuscriptBehaviorInfo behaviorManuscript,
                                                double categoryScore,
                                                double tagScore) {
        double score = 0.0;

        // 分类匹配
        if (candidate.getCategoryId() != null &&
                candidate.getCategoryId().equals(behaviorManuscript.getCategoryId())) {
            score += categoryScore;
        }

        // 标签匹配
        if (candidate.getTags() != null && behaviorManuscript.getTags() != null) {
            Set<String> candidateTags = new HashSet<>(candidate.getTags());
            Set<String> behaviorTags = new HashSet<>(behaviorManuscript.getTags());
            candidateTags.retainAll(behaviorTags);
            score += candidateTags.size() * tagScore;
        }

        return score;
    }

    /**
     * 计算热度分
     *
     * @param manuscript 稿件
     * @return 热度分（0-20）
     */
    private double calculateHotScore(Manuscript manuscript) {
        int viewCount = manuscript.getViewCount() != null ? manuscript.getViewCount() : 0;
        return Math.min(viewCount / HOT_SCORE_DIVISOR, MAX_HOT_SCORE);
    }

    /**
     * 计算随机分
     *
     * @return 随机分（0-30）
     */
    private double calculateRandomScore() {
        return Math.random() * MAX_RANDOM_SCORE;
    }

    /**
     * 加权随机排序
     * 使用权重进行随机抽样，权重越高被选中的概率越大
     *
     * @param weightedList 带权重的稿件列表
     * @param size         需要返回的数量
     * @return 排序后的稿件列表
     */
    private List<Manuscript> weightedRandomSort(List<WeightedManuscript> weightedList, int size) {
        if (weightedList.isEmpty()) {
            return Collections.emptyList();
        }

        // 创建可修改的副本
        List<WeightedManuscript> remaining = new ArrayList<>(weightedList);
        List<Manuscript> result = new ArrayList<>();
        Random random = new Random();

        // 需要返回的数量不能超过总数
        int targetSize = Math.min(size, remaining.size());

        while (result.size() < targetSize && !remaining.isEmpty()) {
            // 计算总权重
            double totalWeight = remaining.stream()
                    .mapToDouble(WeightedManuscript::getWeight)
                    .sum();

            if (totalWeight <= 0) {
                // 如果总权重为0，直接随机选择
                int index = random.nextInt(remaining.size());
                result.add(remaining.remove(index).getManuscript());
                continue;
            }

            // 加权随机选择
            double randomValue = random.nextDouble() * totalWeight;
            double cumulativeWeight = 0.0;

            for (int i = 0; i < remaining.size(); i++) {
                cumulativeWeight += remaining.get(i).getWeight();
                if (cumulativeWeight >= randomValue) {
                    result.add(remaining.remove(i).getManuscript());
                    break;
                }
            }
        }

        return result;
    }

    /**
     * 用户行为数据类
     */
    @Data
    public static class UserBehavior {
        /**
         * 浏览历史列表
         */
        private List<ManuscriptBehaviorInfo> watchHistory;

        /**
         * 点赞历史列表
         */
        private List<ManuscriptBehaviorInfo> likeHistory;
    }

    /**
     * 稿件行为信息类
     * 用于存储用户行为关联的稿件信息
     */
    @Data
    public static class ManuscriptBehaviorInfo {
        /**
         * 稿件ID
         */
        private Integer manuscriptId;

        /**
         * 分类ID
         */
        private Integer categoryId;

        /**
         * 标签列表
         */
        private List<String> tags;
    }

    /**
     * 带权重的稿件类
     */
    @Data
    public static class WeightedManuscript {
        /**
         * 稿件
         */
        private Manuscript manuscript;

        /**
         * 权重值
         */
        private double weight;

        public WeightedManuscript(Manuscript manuscript, double weight) {
            this.manuscript = manuscript;
            this.weight = weight;
        }
    }
}
