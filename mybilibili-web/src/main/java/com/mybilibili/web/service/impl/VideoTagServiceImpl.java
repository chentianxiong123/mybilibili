package com.mybilibili.web.service.impl;

import com.mybilibili.web.service.VideoTagService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import java.util.stream.Collectors;

@Slf4j
@Service
public class VideoTagServiceImpl implements VideoTagService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String KEY_VIDEO_TAGS = "video:tags:%d";
    private static final String KEY_TAG_VIDEOS = "tag:videos:%s";

    private static final long EXPIRE_DAYS = 365;

    @Override
    public boolean addTag(Integer videoId, String tagName) {
        if (videoId == null || tagName == null || tagName.trim().isEmpty()) {
            return false;
        }

        String videoTagsKey = String.format(KEY_VIDEO_TAGS, videoId);
        String tagVideosKey = String.format(KEY_TAG_VIDEOS, tagName.trim());

        redisTemplate.opsForSet().add(videoTagsKey, tagName.trim());
        redisTemplate.opsForSet().add(tagVideosKey, String.valueOf(videoId));

        redisTemplate.expire(videoTagsKey, EXPIRE_DAYS, TimeUnit.DAYS);
        redisTemplate.expire(tagVideosKey, EXPIRE_DAYS, TimeUnit.DAYS);

        log.info("视频 {} 添加标签 {}", videoId, tagName);
        return true;
    }

    @Override
    public boolean addTags(Integer videoId, List<String> tagNames) {
        if (videoId == null || tagNames == null || tagNames.isEmpty()) {
            return false;
        }

        for (String tagName : tagNames) {
            addTag(videoId, tagName);
        }
        return true;
    }

    @Override
    public boolean removeTag(Integer videoId, String tagName) {
        if (videoId == null || tagName == null || tagName.trim().isEmpty()) {
            return false;
        }

        String videoTagsKey = String.format(KEY_VIDEO_TAGS, videoId);
        String tagVideosKey = String.format(KEY_TAG_VIDEOS, tagName.trim());

        redisTemplate.opsForSet().remove(videoTagsKey, tagName.trim());
        redisTemplate.opsForSet().remove(tagVideosKey, String.valueOf(videoId));

        log.info("视频 {} 移除标签 {}", videoId, tagName);
        return true;
    }

    @Override
    public List<String> getVideoTags(Integer videoId) {
        if (videoId == null) {
            return new ArrayList<>();
        }

        String videoTagsKey = String.format(KEY_VIDEO_TAGS, videoId);
        Set<String> tags = redisTemplate.opsForSet().members(videoTagsKey);

        if (tags == null || tags.isEmpty()) {
            return new ArrayList<>();
        }

        return new ArrayList<>(tags);
    }

    @Override
    public List<Object> getVideosTags(List<Integer> videoIds) {
        if (videoIds == null || videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        List<Object> result = new ArrayList<>();
        for (Integer videoId : videoIds) {
            result.add(getVideoTags(videoId));
        }
        return result;
    }

    @Override
    public List<Integer> getVideosByTag(String tagName) {
        if (tagName == null || tagName.trim().isEmpty()) {
            return new ArrayList<>();
        }

        String tagVideosKey = String.format(KEY_TAG_VIDEOS, tagName.trim());
        Set<String> videoIds = redisTemplate.opsForSet().members(tagVideosKey);

        if (videoIds == null || videoIds.isEmpty()) {
            return new ArrayList<>();
        }

        return videoIds.stream().map(Integer::parseInt).collect(Collectors.toList());
    }

    @Override
    public boolean clearVideoTags(Integer videoId) {
        if (videoId == null) {
            return false;
        }

        String videoTagsKey = String.format(KEY_VIDEO_TAGS, videoId);

        Set<String> existingTags = redisTemplate.opsForSet().members(videoTagsKey);
        if (existingTags != null && !existingTags.isEmpty()) {
            for (String tagName : existingTags) {
                String tagVideosKey = String.format(KEY_TAG_VIDEOS, tagName);
                redisTemplate.opsForSet().remove(tagVideosKey, String.valueOf(videoId));
            }
        }

        redisTemplate.delete(videoTagsKey);
        log.info("清除视频 {} 的所有标签", videoId);
        return true;
    }

    @Override
    public void syncTags(Integer videoId, List<String> tagNames) {
        if (videoId == null) {
            return;
        }

        clearVideoTags(videoId);
        if (tagNames != null && !tagNames.isEmpty()) {
            addTags(videoId, tagNames);
        }
        log.info("同步视频 {} 的标签到Redis: {}", videoId, tagNames);
    }
}
