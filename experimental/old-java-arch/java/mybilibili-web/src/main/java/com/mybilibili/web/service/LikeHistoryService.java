package com.mybilibili.web.service;

import java.util.List;

/**
 * 点赞历史服务接口（Redis版）
 */
public interface LikeHistoryService {

    /**
     * 记录用户点赞
     *
     * @param userId    用户ID
     * @param targetType 目标类型（VIDEO/DYNAMIC等）
     * @param targetId   目标ID
     * @return 是否成功
     */
    boolean like(Integer userId, String targetType, Integer targetId);

    /**
     * 取消点赞
     *
     * @param userId    用户ID
     * @param targetType 目标类型
     * @param targetId   目标ID
     * @return 是否成功
     */
    boolean unlike(Integer userId, String targetType, Integer targetId);

    /**
     * 检查是否已点赞
     *
     * @param userId    用户ID
     * @param targetType 目标类型
     * @param targetId   目标ID
     * @return 是否已点赞
     */
    boolean isLiked(Integer userId, String targetType, Integer targetId);

    /**
     * 批量检查是否已点赞
     *
     * @param userId     用户ID
     * @param targetType 目标类型
     * @param targetIds  目标ID列表
     * @return 点赞状态Map
     */
    List<Integer> getLikedTargetIds(Integer userId, String targetType, List<Integer> targetIds);

    /**
     * 获取用户点赞历史（最近N条）
     *
     * @param userId     用户ID
     * @param days       最近天数（0表示不限制）
     * @param limit      返回数量限制
     * @return 目标ID列表（按时间倒序）
     */
    List<Integer> getUserLikeHistory(Integer userId, int days, int limit);

    /**
     * 获取用户点赞的稿件ID列表
     *
     * @param userId  用户ID
     * @param days    最近天数（0表示不限制）
     * @param limit   返回数量限制
     * @return 稿件ID列表
     */
    List<Integer> getUserLikedManuscriptIds(Integer userId, int days, int limit);

    /**
     * 获取目标被点赞次数
     *
     * @param targetType 目标类型
     * @param targetId   目标ID
     * @return 点赞数
     */
    long getLikeCount(String targetType, Integer targetId);

    /**
     * 批量获取点赞数
     *
     * @param targetType 目标类型
     * @param targetIds  目标ID列表
     * @return 点赞数Map
     */
    List<Object> getLikeCounts(String targetType, List<Integer> targetIds);
}
