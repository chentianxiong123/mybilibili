package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.User;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

@Mapper
public interface UserMapper {
    // 根据用户名查询用户
    User findByUsername(@Param("username") String username);

    // 插入新用户
    int insert(User user);

    // 根据ID查询用户
    User findById(@Param("id") Integer id);
    
    // 更新用户信息
    int update(User user);
    
    // 更新用户点赞数
    int updateLikedCount(@Param("id") Integer id, @Param("count") int count);
    
    // 更新用户硬币数
    int updateCoinCount(@Param("id") Integer id, @Param("count") int count);
    
    // 更新用户经验值
    int updateExperience(@Param("id") Integer id, @Param("count") int count);

    // 统计用户关注数量
    int countFollowing(@Param("userId") Integer userId);

    // 统计用户粉丝数量
    int countFollowers(@Param("userId") Integer userId);

    // 统计用户动态数量
    int countDynamics(@Param("userId") Integer userId);

    // 查询所有用户
    java.util.List<User> findAll();
}