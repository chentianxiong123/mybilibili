package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.web.service.impl.RandomRecommendServiceImpl;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface UserInteractionMapper {

    int insert(UserInteraction interaction);

    int delete(@Param("userId") Integer userId,
               @Param("targetType") String targetType,
               @Param("targetId") Integer targetId,
               @Param("interactionType") String interactionType);

    UserInteraction findByUserAndTarget(@Param("userId") Integer userId,
                                        @Param("targetType") String targetType,
                                        @Param("targetId") Integer targetId,
                                        @Param("interactionType") String interactionType);

    List<UserInteraction> findByUserAndTargets(@Param("userId") Integer userId,
                                               @Param("targetType") String targetType,
                                               @Param("targetIds") List<Integer> targetIds,
                                               @Param("interactionType") String interactionType);

    int countByTarget(@Param("targetType") String targetType,
                      @Param("targetId") Integer targetId,
                      @Param("interactionType") String interactionType);

    List<UserInteraction> countByTargets(@Param("targetType") String targetType,
                                         @Param("targetIds") List<Integer> targetIds,
                                         @Param("interactionType") String interactionType);

    List<UserInteraction> findByUserAndInteractionType(@Param("userId") Integer userId,
                                                       @Param("targetType") String targetType,
                                                       @Param("interactionType") String interactionType);

    /**
     * 查询粉丝列表（关注了目标用户的所有用户）
     */
    List<UserInteraction> findFollowersByTargetId(@Param("targetId") Integer targetId,
                                                   @Param("targetType") String targetType,
                                                   @Param("interactionType") String interactionType,
                                                   @Param("offset") Integer offset,
                                                   @Param("size") Integer size);

    /**
     * 统计粉丝数量
     */
    int countFollowersByTargetId(@Param("targetId") Integer targetId,
                                  @Param("targetType") String targetType,
                                  @Param("interactionType") String interactionType);

    /**
     * 查询互关粉丝列表（双方互相关注）
     */
    List<UserInteraction> findMutualFollowers(@Param("userId") Integer userId,
                                               @Param("targetType") String targetType,
                                               @Param("interactionType") String interactionType,
                                               @Param("offset") Integer offset,
                                               @Param("size") Integer size);

    /**
     * 统计互关粉丝数量
     */
    int countMutualFollowers(@Param("userId") Integer userId,
                              @Param("targetType") String targetType,
                              @Param("interactionType") String interactionType);

    /**
     * 查询用户最近点赞历史（带稿件详情）
     * 用于推荐算法
     *
     * @param userId       用户ID
     * @param days         最近几天
     * @param limit        限制数量
     * @return 稿件行为信息列表
     */
    List<RandomRecommendServiceImpl.ManuscriptBehaviorInfo> selectRecentLikeHistoryWithDetails(
            @Param("userId") Integer userId,
            @Param("days") int days,
            @Param("limit") int limit);
}
