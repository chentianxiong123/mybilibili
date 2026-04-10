package com.mybilibili.user.service.impl;

import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.user.mapper.UserInteractionMapper;
import com.mybilibili.user.mapper.UserMapper;
import com.mybilibili.user.service.FollowService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;

@Service
public class FollowServiceImpl implements FollowService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Autowired
    private UserMapper userMapper;

    @Override
    @Transactional
    public boolean follow(Integer userId, Integer targetUserId) {
        UserInteraction existing = userInteractionMapper.selectInteraction(
            userId, targetUserId, "user", "follow");
        if (existing != null) {
            return false;
        }

        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(userId);
        interaction.setTargetType("user");
        interaction.setTargetId(targetUserId);
        interaction.setInteractionType("follow");
        userInteractionMapper.insert(interaction);

        userMapper.incrementFollowingCount(userId);
        userMapper.incrementFollowerCount(targetUserId);

        return true;
    }

    @Override
    @Transactional
    public boolean unfollow(Integer userId, Integer targetUserId) {
        int deleted = userInteractionMapper.deleteInteraction(
            userId, targetUserId, "user", "follow");
        if (deleted == 0) {
            return false;
        }

        userMapper.decrementFollowingCount(userId);
        userMapper.decrementFollowerCount(targetUserId);

        return true;
    }

    @Override
    public boolean isFollowing(Integer userId, Integer targetUserId) {
        return userInteractionMapper.isFollowing(userId, targetUserId);
    }

    @Override
    public List<UserVO> getFollowingList(Integer userId) {
        List<Integer> followingIds = userInteractionMapper.selectFollowingIds(userId);
        List<UserVO> userVOs = new ArrayList<>();
        for (Integer followingId : followingIds) {
            User user = userMapper.selectById(followingId);
            if (user != null) {
                UserVO userVO = new UserVO();
                userVO.setId(user.getId());
                userVO.setUsername(user.getUsername());
                userVO.setNickname(user.getNickname());
                userVO.setAvatar(user.getAvatar());
                userVO.setLevel(user.getLevel());
                userVOs.add(userVO);
            }
        }
        return userVOs;
    }

    @Override
    public List<UserVO> getFollowerList(Integer userId) {
        List<Integer> followerIds = userInteractionMapper.selectFollowerIds(userId);
        List<UserVO> userVOs = new ArrayList<>();
        for (Integer followerId : followerIds) {
            User user = userMapper.selectById(followerId);
            if (user != null) {
                UserVO userVO = new UserVO();
                userVO.setId(user.getId());
                userVO.setUsername(user.getUsername());
                userVO.setNickname(user.getNickname());
                userVO.setAvatar(user.getAvatar());
                userVO.setLevel(user.getLevel());
                userVOs.add(userVO);
            }
        }
        return userVOs;
    }
}