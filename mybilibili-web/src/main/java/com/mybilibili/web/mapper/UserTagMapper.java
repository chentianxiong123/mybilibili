package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.UserTag;
import org.apache.ibatis.annotations.*;

import java.util.List;

@Mapper
public interface UserTagMapper {

    @Select("SELECT * FROM user_tags WHERE user_id = #{userId}")
    List<UserTag> findByUserId(Integer userId);

    @Insert("INSERT INTO user_tags (user_id, tag_name) VALUES (#{userId}, #{tagName})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(UserTag tag);

    @Delete("DELETE FROM user_tags WHERE user_id = #{userId} AND tag_name = #{tagName}")
    int deleteByUserIdAndTagName(@Param("userId") Integer userId, @Param("tagName") String tagName);

    @Delete("DELETE FROM user_tags WHERE user_id = #{userId}")
    int deleteByUserId(Integer userId);

    @Select("SELECT COUNT(*) FROM user_tags WHERE user_id = #{userId}")
    int countByUserId(Integer userId);
}
