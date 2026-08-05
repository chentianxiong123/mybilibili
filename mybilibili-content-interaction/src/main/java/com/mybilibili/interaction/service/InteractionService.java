package com.mybilibili.interaction.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.danmaku.feign.VideoClient;
import com.mybilibili.interaction.mapper.UserInteractionMapper;
import com.mybilibili.mq.UserEventMQProducer;
import com.mybilibili.mq.UserNotificationEvent;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

@Service
public class InteractionService {

    @Autowired
    private UserInteractionMapper interactionMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoClient videoClient;

    @Autowired
    private UserEventMQProducer userEventMQProducer;

    public Result<Void> like(Integer userId, String targetType, Integer targetId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, targetType)
               .eq(UserInteraction::getTargetId, targetId)
               .eq(UserInteraction::getInteractionType, "LIKE");
        UserInteraction exist = interactionMapper.selectOne(wrapper);
        if (exist != null) {
            return Result.success();
        }
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType(targetType);
        interaction.setTargetId(targetId);
        interaction.setInteractionType("LIKE");
        interactionMapper.insert(interaction);
        redisTemplate.opsForValue().increment("like:" + targetType + ":" + targetId);

        if ("VIDEO".equals(targetType)) {
            try {
                Result<VideoVO> videoResult = videoClient.getVideoById(targetId);
                if (videoResult != null && videoResult.getCode() == 200 && videoResult.getData() != null) {
                    VideoVO video = videoResult.getData();
                    if (video.getUploader() != null) {
                        Integer ownerId = video.getUploader().getId();
                        if (ownerId != null && !ownerId.equals(userId)) {
                            userEventMQProducer.sendNotificationEvent(
                                    UserNotificationEvent.videoLike(userId, ownerId, targetId, video.getTitle()));
                        }
                    }
                }
            } catch (Exception e) {
            }
        }

        return Result.success();
    }

    public Result<Void> unlike(Integer userId, String targetType, Integer targetId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, targetType)
               .eq(UserInteraction::getTargetId, targetId)
               .eq(UserInteraction::getInteractionType, "LIKE");
        interactionMapper.delete(wrapper);
        redisTemplate.opsForValue().decrement("like:" + targetType + ":" + targetId);
        return Result.success();
    }

    public Result<Boolean> checkLike(Integer userId, String targetType, Integer targetId) {
        if (userId == null) {
            return Result.success(false);
        }
        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, targetType)
               .eq(UserInteraction::getTargetId, targetId)
               .eq(UserInteraction::getInteractionType, "LIKE");
        UserInteraction exist = interactionMapper.selectOne(wrapper);
        return Result.success(exist != null);
    }

    public Result<Void> coin(Integer userId, Integer manuscriptId, Integer amount) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType("MANUSCRIPT");
        interaction.setTargetId(manuscriptId);
        interaction.setInteractionType("COIN");
        interactionMapper.insert(interaction);
        redisTemplate.opsForValue().increment("coin:MANUSCRIPT:" + manuscriptId, amount);
        return Result.success();
    }

    public Result<Void> collect(Integer userId, Integer manuscriptId, Integer folderId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType("MANUSCRIPT");
        interaction.setTargetId(manuscriptId);
        interaction.setInteractionType("COLLECT");
        interactionMapper.insert(interaction);
        redisTemplate.opsForValue().increment("collect:MANUSCRIPT:" + manuscriptId);
        return Result.success();
    }

    public Result<Void> uncollect(Integer userId, Integer manuscriptId, Integer folderId) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getUserId, userId)
               .eq(UserInteraction::getTargetType, "MANUSCRIPT")
               .eq(UserInteraction::getTargetId, manuscriptId)
               .eq(UserInteraction::getInteractionType, "COLLECT");
        interactionMapper.delete(wrapper);
        redisTemplate.opsForValue().decrement("collect:MANUSCRIPT:" + manuscriptId);
        return Result.success();
    }

    public Result<Void> share(Integer userId, Integer manuscriptId) {
        if (userId == null) {
            return Result.success();
        }
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType("MANUSCRIPT");
        interaction.setTargetId(manuscriptId);
        interaction.setInteractionType("SHARE");
        interactionMapper.insert(interaction);
        redisTemplate.opsForValue().increment("share:MANUSCRIPT:" + manuscriptId);
        return Result.success();
    }

    public int getLikeCount(String targetType, Integer targetId) {
        String key = "like:" + targetType + ":" + targetId;
        String countStr = redisTemplate.opsForValue().get(key);
        if (countStr != null) {
            return Integer.parseInt(countStr);
        }
        LambdaQueryWrapper<UserInteraction> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(UserInteraction::getTargetType, targetType)
               .eq(UserInteraction::getTargetId, targetId)
               .eq(UserInteraction::getInteractionType, "LIKE");
        long count = interactionMapper.selectCount(wrapper);
        redisTemplate.opsForValue().set(key, String.valueOf(count));
        return (int) count;
    }
}
