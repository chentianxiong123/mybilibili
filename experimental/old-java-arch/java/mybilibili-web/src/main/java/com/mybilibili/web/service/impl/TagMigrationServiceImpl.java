package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.Tag;
import com.mybilibili.web.mapper.TagMapper;
import com.mybilibili.web.mapper.VideoMapper;
import com.mybilibili.web.service.TagMigrationService;
import com.mybilibili.web.service.VideoTagService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Set;
import java.util.stream.Collectors;

@Slf4j
@Service
public class TagMigrationServiceImpl implements TagMigrationService {

    @Autowired
    private TagMapper tagMapper;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoTagService videoTagService;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String KEY_VIDEO_TAGS = "video:tags:%d";
    private static final String KEY_TAG_VIDEOS = "tag:videos:%s";

    @Override
    public boolean migrateVideoTags(Integer videoId) {
        if (videoId == null) {
            return false;
        }

        try {
            // 从 MySQL 获取标签
            List<Tag> tags = tagMapper.selectByVideoId(videoId);
            List<String> tagNames = tags.stream()
                    .map(Tag::getName)
                    .collect(Collectors.toList());

            // 同步到 Redis
            videoTagService.syncTags(videoId, tagNames);

            log.info("视频 {} 的标签迁移完成，共 {} 个标签", videoId, tagNames.size());
            return true;
        } catch (Exception e) {
            log.error("视频 {} 的标签迁移失败", videoId, e);
            return false;
        }
    }

    @Override
    public int migrateVideoTagsBatch(List<Integer> videoIds) {
        if (videoIds == null || videoIds.isEmpty()) {
            return 0;
        }

        int successCount = 0;
        for (Integer videoId : videoIds) {
            if (migrateVideoTags(videoId)) {
                successCount++;
            }
        }

        log.info("批量迁移完成，成功 {}/{} 个视频", successCount, videoIds.size());
        return successCount;
    }

    @Override
    public int migrateAllVideoTags() {
        try {
            // 获取所有视频ID
            List<Integer> allVideoIds = videoMapper.selectAllVideoIds();

            if (allVideoIds == null || allVideoIds.isEmpty()) {
                log.warn("没有找到任何视频");
                return 0;
            }

            log.info("开始迁移所有视频的标签，共 {} 个视频", allVideoIds.size());
            return migrateVideoTagsBatch(allVideoIds);
        } catch (Exception e) {
            log.error("获取视频列表失败", e);
            return 0;
        }
    }

    @Override
    public TagMigrationResult verifyMigration(Integer videoId) {
        TagMigrationResult result = new TagMigrationResult();
        result.setVideoId(videoId);

        try {
            // 从 MySQL 获取标签
            List<Tag> mysqlTagList = tagMapper.selectByVideoId(videoId);
            List<String> mysqlTags = mysqlTagList.stream()
                    .map(Tag::getName)
                    .sorted()
                    .collect(Collectors.toList());
            result.setMysqlTags(mysqlTags);

            // 从 Redis 获取标签
            List<String> redisTags = videoTagService.getVideoTags(videoId);
            redisTags = redisTags.stream().sorted().collect(Collectors.toList());
            result.setRedisTags(redisTags);

            // 比较
            boolean matched = mysqlTags.equals(redisTags);
            result.setMatched(matched);

            if (matched) {
                result.setMessage("验证通过，MySQL 和 Redis 标签一致");
            } else {
                result.setMessage(String.format(
                        "验证失败，MySQL: %s, Redis: %s",
                        mysqlTags, redisTags));
            }
        } catch (Exception e) {
            result.setMatched(false);
            result.setMessage("验证异常: " + e.getMessage());
            log.error("验证视频 {} 的标签迁移失败", videoId, e);
        }

        return result;
    }

    @Override
    public boolean clearAllRedisTags() {
        try {
            // 获取所有视频标签的 key
            Set<String> videoTagKeys = redisTemplate.keys("video:tags:*");
            if (videoTagKeys != null && !videoTagKeys.isEmpty()) {
                redisTemplate.delete(videoTagKeys);
                log.info("清空了 {} 个视频标签 key", videoTagKeys.size());
            }

            // 获取所有标签索引的 key
            Set<String> tagVideoKeys = redisTemplate.keys("tag:videos:*");
            if (tagVideoKeys != null && !tagVideoKeys.isEmpty()) {
                redisTemplate.delete(tagVideoKeys);
                log.info("清空了 {} 个标签索引 key", tagVideoKeys.size());
            }

            log.info("Redis 标签数据清空完成");
            return true;
        } catch (Exception e) {
            log.error("清空 Redis 标签数据失败", e);
            return false;
        }
    }
}
