package com.mybilibili.user.service;

import com.mybilibili.common.vo.UserVO;
import java.util.List;

public interface FollowService {
    boolean follow(Integer userId, Integer targetUserId);
    boolean unfollow(Integer userId, Integer targetUserId);
    boolean isFollowing(Integer userId, Integer targetUserId);
    List<UserVO> getFollowingList(Integer userId);
    List<UserVO> getFollowerList(Integer userId);
}