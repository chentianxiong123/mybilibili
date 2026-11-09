package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.Follow;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface FollowMapper {
    // 添加关注
    int insert(Follow follow);
    
    // 取消关注
    int delete(@Param("followerId") Integer followerId, @Param("followedId") Integer followedId);
    
    // 检查是否已关注
    Follow findByFollowerAndFollowed(@Param("followerId") Integer followerId, @Param("followedId") Integer followedId);
    
    // 获取用户的关注列表
    List<Follow> findByFollowerId(@Param("followerId") Integer followerId);
    
    // 获取用户的粉丝列表
    List<Follow> findByFollowedId(@Param("followedId") Integer followedId);
    
    // 获取用户关注的所有用户ID列表
    List<Integer> getFollowedUserIds(@Param("followerId") Integer followerId);

    // 获取用户的粉丝数量
    int countFollowers(@Param("followedId") Integer followedId);
}