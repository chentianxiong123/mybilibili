package com.mybilibili.interaction.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

@Mapper
public interface UserMapper extends BaseMapper<User> {

    @Select("SELECT * FROM users WHERE id = #{id}")
    User findById(@Param("id") Integer id);

    @Update("UPDATE users SET liked_count = liked_count + #{count} WHERE id = #{id}")
    int updateLikedCount(@Param("id") Integer id, @Param("count") Integer count);

    @Update("UPDATE users SET coin_count = coin_count + #{count} WHERE id = #{id}")
    int updateCoinCount(@Param("id") Integer id, @Param("count") Integer count);
}
