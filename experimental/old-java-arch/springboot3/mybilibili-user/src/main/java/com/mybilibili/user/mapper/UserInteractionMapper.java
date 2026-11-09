package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.UserInteraction;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface UserInteractionMapper extends BaseMapper<UserInteraction> {

    @Select("SELECT * FROM user_interactions WHERE user_id = #{userId} AND target_id = #{targetId} AND target_type = #{targetType} AND interaction_type = #{interactionType} LIMIT 1")
    UserInteraction selectInteraction(@Param("userId") Integer userId, @Param("targetId") Integer targetId, @Param("targetType") String targetType, @Param("interactionType") String interactionType);

    @Insert("INSERT INTO user_interactions (user_id, target_type, target_id, interaction_type, created_at) VALUES (#{userId}, #{targetType}, #{targetId}, #{interactionType}, NOW())")
    int insertInteraction(@Param("userId") Integer userId, @Param("targetId") Integer targetId, @Param("targetType") String targetType, @Param("interactionType") String interactionType);

    @Delete("DELETE FROM user_interactions WHERE user_id = #{userId} AND target_id = #{targetId} AND target_type = #{targetType} AND interaction_type = #{interactionType}")
    int deleteInteraction(@Param("userId") Integer userId, @Param("targetId") Integer targetId, @Param("targetType") String targetType, @Param("interactionType") String interactionType);

    @Select("SELECT COUNT(*) > 0 FROM user_interactions WHERE user_id = #{userId} AND target_id = #{targetId} AND target_type = 'user' AND interaction_type = 'follow'")
    boolean isFollowing(@Param("userId") Integer userId, @Param("targetId") Integer targetId);

    @Select("SELECT target_id FROM user_interactions WHERE user_id = #{userId} AND target_type = 'user' AND interaction_type = 'follow'")
    List<Integer> selectFollowingIds(@Param("userId") Integer userId);

    @Select("SELECT user_id FROM user_interactions WHERE target_id = #{userId} AND target_type = 'user' AND interaction_type = 'follow'")
    List<Integer> selectFollowerIds(@Param("userId") Integer userId);
}