package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.UserDynamic;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface UserDynamicMapper extends BaseMapper<UserDynamic> {

    @Select("SELECT * FROM user_dynamics WHERE user_id = #{userId} AND status = 0 ORDER BY created_at DESC LIMIT #{limit} OFFSET #{offset}")
    List<UserDynamic> getByUserId(@Param("userId") Integer userId, @Param("offset") int offset, @Param("limit") int limit);

    @Select("SELECT * FROM user_dynamics WHERE user_id IN (${userIds}) AND status = 0 ORDER BY created_at DESC LIMIT #{limit} OFFSET #{offset}")
    List<UserDynamic> getByUserIds(@Param("userIds") String userIds, @Param("offset") int offset, @Param("limit") int limit);

    @Select("SELECT * FROM user_dynamics WHERE status = 0 ORDER BY created_at DESC LIMIT #{limit} OFFSET #{offset}")
    List<UserDynamic> getAllDynamics(@Param("offset") int offset, @Param("limit") int limit);

    @Select("SELECT * FROM user_dynamics WHERE id = #{id}")
    UserDynamic getById(@Param("id") Integer id);

    @Update("UPDATE user_dynamics SET like_count = #{likeCount} WHERE id = #{id}")
    void updateLikeCount(@Param("id") Integer id, @Param("likeCount") int likeCount);

    @Update("UPDATE user_dynamics SET share_count = #{shareCount} WHERE id = #{id}")
    void updateShareCount(@Param("id") Integer id, @Param("shareCount") int shareCount);

    @Update("UPDATE user_dynamics SET comment_count = #{commentCount} WHERE id = #{id}")
    void updateCommentCount(@Param("id") Integer id, @Param("commentCount") int commentCount);

    @Delete("DELETE FROM user_dynamics WHERE id = #{id}")
    void deleteById(@Param("id") Integer id);
}
