package com.mybilibili.interaction.service.impl;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.common.vo.WatchHistoryVO;
import com.mybilibili.interaction.feign.VideoClient;
import com.mybilibili.interaction.service.WatchHistoryService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Slf4j
@Service
public class WatchHistoryServiceImpl implements WatchHistoryService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoClient videoClient;

    private static final String KEY_WATCH_HISTORY_LIST = "watch_history:user:%s";
    private static final String KEY_WATCH_HISTORY_DETAIL = "watch_history:detail:%s:%s";
    private static final long EXPIRE_DAYS = 30;

    @Override
    public List<WatchHistoryVO> getWatchHistoryList(Integer userId, Integer page, Integer size) {
        if (userId == null) {
            return new ArrayList<>();
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);
        int start = (page - 1) * size;
        int end = start + size - 1;

        Set<String> videoIds = redisTemplate.opsForZSet().reverseRange(listKey, start, end);

        if (videoIds == null || videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<WatchHistoryVO> historyVOs = new ArrayList<>();
        int index = start + 1;

        for (String videoIdStr : videoIds) {
            Integer videoId = Integer.valueOf(videoIdStr);
            String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoId);

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

            try {
                Result<VideoVO> result = videoClient.getVideoById(videoId);
                if (result != null && result.getCode() == 200 && result.getData() != null) {
                    vo.setVideo(result.getData());
                }
            } catch (Exception e) {
                log.warn("获取视频信息失败: {}", videoId);
            }

            historyVOs.add(vo);
        }

        return historyVOs;
    }

    @Override
    public void recordWatchHistory(Integer userId, Integer videoId, Integer progressSeconds, Integer videoDuration) {
        if (userId == null || videoId == null) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);
        String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoId);
        long currentTimeMillis = System.currentTimeMillis();

        int watchPercentage = videoDuration > 0 ? (int) ((progressSeconds * 100.0) / videoDuration) : 0;
        watchPercentage = Math.min(watchPercentage, 100);

        redisTemplate.opsForZSet().add(listKey, String.valueOf(videoId), currentTimeMillis);

        redisTemplate.opsForHash().put(detailKey, "progressSeconds", String.valueOf(progressSeconds));
        redisTemplate.opsForHash().put(detailKey, "videoDuration", String.valueOf(videoDuration));
        redisTemplate.opsForHash().put(detailKey, "watchPercentage", String.valueOf(watchPercentage));
        redisTemplate.opsForHash().put(detailKey, "watchedAt", String.valueOf(currentTimeMillis));

        redisTemplate.expire(listKey, EXPIRE_DAYS, TimeUnit.DAYS);
        redisTemplate.expire(detailKey, EXPIRE_DAYS, TimeUnit.DAYS);

        log.debug("记录用户 {} 的视频 {} 浏览历史", userId, videoId);
    }

    @Override
    public void clearWatchHistory(Integer userId) {
        if (userId == null) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);
        Set<String> videoIds = redisTemplate.opsForZSet().range(listKey, 0, -1);

        if (videoIds != null && !videoIds.isEmpty()) {
            for (String videoIdStr : videoIds) {
                String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoIdStr);
                redisTemplate.delete(detailKey);
            }
        }

        redisTemplate.delete(listKey);
        log.info("清空用户 {} 的浏览历史", userId);
    }

    @Override
    public void deleteWatchHistory(Integer id, Integer userId) {
        if (userId == null || id == null) {
            return;
        }

        String listKey = String.format(KEY_WATCH_HISTORY_LIST, userId);
        Set<String> videoIds = redisTemplate.opsForZSet().reverseRange(listKey, 0, -1);

        if (videoIds == null || videoIds.isEmpty()) {
            return;
        }

        List<String> videoIdList = new ArrayList<>(videoIds);
        int index = id - 1;
        if (index < 0 || index >= videoIdList.size()) {
            return;
        }

        String videoIdStr = videoIdList.get(index);
        String detailKey = String.format(KEY_WATCH_HISTORY_DETAIL, userId, videoIdStr);

        redisTemplate.opsForZSet().remove(listKey, videoIdStr);
        redisTemplate.delete(detailKey);

        log.info("删除用户 {} 的视频 {} 浏览历史", userId, videoIdStr);
    }
}
