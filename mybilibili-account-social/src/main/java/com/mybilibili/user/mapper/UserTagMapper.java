package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.UserTag;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Delete;

import java.util.List;

@Mapper
public interface UserTagMapper extends BaseMapper<UserTag> {

    @Select("SELECT * FROM user_tags WHERE user_id = #{userId} ORDER BY created_at DESC")
    List<UserTag> selectByUserId(Integer userId);

    @Select("SELECT * FROM user_tags WHERE user_id = #{userId} AND tag_name = #{tagName}")
    UserTag selectByUserIdAndTagName(Integer userId, String tagName);

    @Delete("DELETE FROM user_tags WHERE user_id = #{userId} AND tag_name = #{tagName}")
    int deleteByUserIdAndTagName(Integer userId, String tagName);
}
