package com.mybilibili.web.service.impl;

import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.web.service.VideoService;
import com.mybilibili.web.service.WatchHistoryService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Slf4j
@Service
public class WatchHistoryServiceImpl implements WatchHistoryService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoService videoService;

    /**
     * 记录浏览历史的进度阈值（10%）
     */
    private static final double RECORD_THRESHOLD = 0.1;

    /**
     * 用户浏览历史列表 ZSet key 前缀
     */
    private static final String KEY_WATCH_HISTORY_LIST = "watch_history:user:%s";

    /**
     * 浏览详情 Hash key 前缀
     */
    private static final String KEY_WATCH_HISTORY_DETAIL = "watch_history:detail:%s:%s";

    /**
     * Redis 过期时间：30天
     */
    private static final long EXPIRE_DAYS = 30;

    @Override
    public void recordWatchHistory(Integer userId, Integer videoId, Integer progressSeconds, Integer videoDuration) {
        if (userId == null || videoId == null || progressSeconds == null || videoDuration == null) {
            return;
        }

        // 当进度超过视频时长的10%时记录
        if (progressSeconds < videoDuration * RECORD_THRESHOLD) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);
        String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoId);
        long currentTimeMillis = System.currentTimeMillis();

        // 计算观看百分比 (0-100)
        int watchPercentage = videoDuration > 0 ? (int) ((progressSeconds * 100.0) / videoDuration) : 0;
        // 限制最大为100%
        watchPercentage = Math.min(watchPercentage, 100);

        // 将视频ID添加到用户的浏览历史 ZSet 中，score 为时间戳
        redisTemplate.opsForZSet().add(listKey, String.valueOf(videoId), currentTimeMillis);

        // 设置浏览详情到 Hash 中
        redisTemplate.opsForHash().put(detailKey, "progressSeconds", String.valueOf(progressSeconds));
        redisTemplate.opsForHash().put(detailKey, "videoDuration", String.valueOf(videoDuration));
        redisTemplate.opsForHash().put(detailKey, "watchPercentage", String.valueOf(watchPercentage));
        redisTemplate.opsForHash().put(detailKey, "watchedAt", String.valueOf(currentTimeMillis));

        // 设置过期时间
        redisTemplate.expire(listKey, EXPIRE_DAYS, TimeUnit.DAYS);
        redisTemplate.expire(detailKey, EXPIRE_DAYS, TimeUnit.DAYS);

        log.debug("记录用户 {} 的视频 {} 浏览历史，进度：{}秒/{}秒 ({}%)", 
                userId, videoId, progressSeconds, videoDuration, watchPercentage);
    }

    @Override
    public List<WatchHistoryVO> getWatchHistoryList(Integer userId, Integer page, Integer size) {
        if (userId == null) {
            return new ArrayList<>();
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);

        // 计算分页参数
        int start = (page - 1) * size;
        int end = start + size - 1;

        // 从 ZSet 中获取分页的视频ID列表（按时间倒序）
        Set<String> videoIds = redisTemplate.opsForZSet().reverseRange(listKey, start, end);

        if (videoIds == null || videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<WatchHistoryVO> historyVOs = new ArrayList<>();
        int index = start + 1;

        for (String videoIdStr : videoIds) {
            Integer videoId = Integer.valueOf(videoIdStr);
            String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoId);

            // 获取浏览详情
            String progressSecondsStr = (String) redisTemplate.opsForHash().get(detailKey, "progressSeconds");
            String videoDurationStr = (String) redisTemplate.opsForHash().get(detailKey, "videoDuration");
            String watchPercentageStr = (String) redisTemplate.opsForHash().get(detailKey, "watchPercentage");
            String watchedAtStr = (String) redisTemplate.opsForHash().get(detailKey, "watchedAt");

            WatchHistoryVO vo = new WatchHistoryVO();
            vo.setId(index++);
            vo.setVideoId(videoId);

            if (progressSecondsStr != null) {
                vo.setProgressSeconds(Integer.valueOf(progressSecondsStr));
            }

            if (videoDurationStr != null) {
                vo.setVideoDuration(Integer.valueOf(videoDurationStr));
            }

            if (watchPercentageStr != null) {
                vo.setWatchPercentage(Integer.valueOf(watchPercentageStr));
            }

            if (watchedAtStr != null) {
                vo.setWatchedAt(new Date(Long.parseLong(watchedAtStr)));
            }

            // 获取视频信息
            VideoVO videoVO = videoService.getVideoById(videoId);
            vo.setVideo(videoVO);

            historyVOs.add(vo);
        }

        return historyVOs;
    }

    @Override
    public void clearWatchHistory(Integer userId) {
        if (userId == null) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);

        // 获取该用户所有的视频ID
        Set<String> videoIds = redisTemplate.opsForZSet().range(listKey, 0, -1);

        if (videoIds != null && !videoIds.isEmpty()) {
            // 删除所有浏览详情
            for (String videoIdStr : videoIds) {
                String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoIdStr);
                redisTemplate.delete(detailKey);
            }
        }

        // 删除浏览历史列表
        redisTemplate.delete(listKey);

        log.info("清空用户 {} 的浏览历史", userId);
    }

    @Override
    public void deleteWatchHistory(Integer id, Integer userId) {
        // id 在 Redis 实现中对应的是分页中的序号，这里需要根据 videoId 来删除
        // 由于接口定义限制，这里通过获取列表后根据索引删除
        if (userId == null || id == null) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);

        // 获取该用户所有的视频ID（按时间倒序）
        Set<String> videoIds = redisTemplate.opsForZSet().reverseRange(listKey, 0, -1);

        if (videoIds == null || videoIds.isEmpty()) {
            return;
        }

        // 将 Set 转换为 List 以便根据索引访问
        List<String> videoIdList = new ArrayList<>(videoIds);

        // id 是从 1 开始的序号
        int index = id - 1;
        if (index < 0 || index >= videoIdList.size()) {
            log.warn("删除浏览历史失败，无效的索引：{}，用户：{}", id, userId);
            return;
        }

        String videoIdStr = videoIdList.get(index);
        String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoIdStr);

        // 从 ZSet 中移除
        redisTemplate.opsForZSet().remove(listKey, videoIdStr);

        // 删除详情
        redisTemplate.delete(detailKey);

        log.info("删除用户 {} 的视频 {} 浏览历史", userId, videoIdStr);
    }

    @Override
    public void updateWatchProgress(Integer userId, Integer videoId, Integer progressSeconds) {
        if (userId == null || videoId == null || progressSeconds == null) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);
        String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoId);

        // 检查是否存在该浏览记录
        Double score = redisTemplate.opsForZSet().score(listKey, String.valueOf(videoId));
        if (score == null) {
            log.debug("更新进度失败，用户 {} 没有视频 {} 的浏览记录", userId, videoId);
            return;
        }

        // 更新进度
        redisTemplate.opsForHash().put(detailKey, "progressSeconds", String.valueOf(progressSeconds));

        // 刷新过期时间
        redisTemplate.expire(listKey, EXPIRE_DAYS, TimeUnit.DAYS);
        redisTemplate.expire(detailKey, EXPIRE_DAYS, TimeUnit.DAYS);

        log.debug("更新用户 {} 的视频 {} 观看进度为 {} 秒", userId, videoId, progressSeconds);
    }

    @Override
    public List<Integer> getRecentWatchVideoIds(Integer userId, int days, int limit) {
        if (userId == null) {
            return new ArrayList<>();
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);

        // 计算时间范围（毫秒）
        long minScore = 0;
        if (days > 0) {
            minScore = System.currentTimeMillis() - (long) days * 24 * 60 * 60 * 1000;
        }

        Set<String> videoIds;
        if (days > 0) {
            // 按分数范围查询
            videoIds = redisTemplate.opsForZSet().rangeByScore(listKey, minScore, Double.MAX_VALUE);
        } else {
            // 查询最近的N条
            videoIds = redisTemplate.opsForZSet().reverseRange(listKey, 0, limit - 1);
        }

        if (videoIds == null || videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<String> result = new ArrayList<>(videoIds);
        if (limit > 0 && result.size() > limit) {
            result = result.subList(0, limit);
        }

        return result.stream()
                .map(Integer::parseInt)
                .collect(Collectors.toList());
    }
}
