package com.mybilibili.web.service;

import com.mybilibili.common.vo.UserVO;

import java.util.List;

public interface FollowService {
    // 关注用户
    boolean followUser(Integer followerId, Integer followedId);

    // 取消关注
    boolean unfollowUser(Integer followerId, Integer followedId);

    // 检查是否已关注
    boolean isFollowing(Integer followerId, Integer followedId);

    // 获取用户的关注列表
    List<UserVO> getFollowingList(Integer userId);

    // 获取用户的粉丝列表
    List<UserVO> getFollowerList(Integer userId);

    // 获取用户的粉丝数
    int getFollowerCount(Integer userId);

    // 获取用户的关注数
    int getFollowingCount(Integer userId);
}