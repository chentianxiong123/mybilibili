package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.common.enums.InteractionType;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.web.mapper.UserInteractionMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.FollowService;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class FollowServiceImpl implements FollowService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Autowired
    private UserMapper userMapper;

    private static final String TARGET_TYPE_USER = "USER";

    @Override
    public boolean followUser(Integer followerId, Integer followedId) {
        // 检查是否关注自己
        if (followerId.equals(followedId)) {
            throw new RuntimeException("不能关注自己");
        }

        // 检查关注者用户是否存在
        User follower = userMapper.findById(followerId);
        if (follower == null) {
            throw new RuntimeException("当前用户不存在");
        }

        // 检查被关注用户是否存在
        User followedUser = userMapper.findById(followedId);
        if (followedUser == null) {
            throw new RuntimeException("被关注的用户不存在");
        }

        // 检查是否已关注
        UserInteraction existing = userInteractionMapper.findByUserAndTarget(
                followerId, TARGET_TYPE_USER, followedId, InteractionType.FOLLOW.getCode());
        if (existing != null) {
            return false; // 已经关注过了
        }

        // 添加关注记录
        UserInteraction interaction = new UserInteraction();
        interaction.setUserId(followerId);
        interaction.setTargetType(TARGET_TYPE_USER);
        interaction.setTargetId(followedId);
        interaction.setInteractionType(InteractionType.FOLLOW.getCode());
        userInteractionMapper.insert(interaction);

        // 更新用户的关注数和粉丝数
        follower.setFollowingCount(follower.getFollowingCount() + 1);
        userMapper.update(follower);

        followedUser.setFollowerCount(followedUser.getFollowerCount() + 1);
        userMapper.update(followedUser);

        return true;
    }

    @Override
    public boolean unfollowUser(Integer followerId, Integer followedId) {
        // 检查是否已关注
        UserInteraction existing = userInteractionMapper.findByUserAndTarget(
                followerId, TARGET_TYPE_USER, followedId, InteractionType.FOLLOW.getCode());
        if (existing == null) {
            return false; // 还没有关注
        }

        // 删除关注记录
        userInteractionMapper.delete(followerId, TARGET_TYPE_USER, followedId, InteractionType.FOLLOW.getCode());

        // 更新用户的关注数和粉丝数
        User follower = userMapper.findById(followerId);
        User followed = userMapper.findById(followedId);
        if (follower != null && followed != null) {
            if (follower.getFollowingCount() > 0) {
                follower.setFollowingCount(follower.getFollowingCount() - 1);
            }
            if (followed.getFollowerCount() > 0) {
                followed.setFollowerCount(followed.getFollowerCount() - 1);
            }
            userMapper.update(follower);
            userMapper.update(followed);
        }

        return true;
    }

    @Override
    public boolean isFollowing(Integer followerId, Integer followedId) {
        if (followerId == null || followedId == null) {
            return false;
        }
        UserInteraction interaction = userInteractionMapper.findByUserAndTarget(
                followerId, TARGET_TYPE_USER, followedId, InteractionType.FOLLOW.getCode());
        return interaction != null;
    }

    @Override
    public List<UserVO> getFollowingList(Integer userId) {
        List<UserInteraction> interactions = userInteractionMapper.findByUserAndInteractionType(
                userId, TARGET_TYPE_USER, InteractionType.FOLLOW.getCode());
        List<UserVO> userVOs = new ArrayList<>();

        for (UserInteraction interaction : interactions) {
            User user = userMapper.findById(interaction.getTargetId());
            if (user != null) {
                UserVO userVO = new UserVO();
                BeanUtils.copyProperties(user, userVO);
                userVOs.add(userVO);
            }
        }

        return userVOs;
    }

    @Override
    public List<UserVO> getFollowerList(Integer userId) {
        // 查询所有关注该用户的记录
        // 这里需要一个反向查询，获取所有targetId=userId且interactionType=FOLLOW的记录
        // 由于当前Mapper设计是按userId查询，我们需要查询所有用户的关注记录然后过滤
        // 暂时简化处理，实际生产环境应该添加专门的查询方法
        List<User> allUsers = userMapper.findAll();
        List<UserVO> followers = new ArrayList<>();

        for (User user : allUsers) {
            if (isFollowing(user.getId(), userId)) {
                UserVO userVO = new UserVO();
                BeanUtils.copyProperties(user, userVO);
                followers.add(userVO);
            }
        }

        return followers;
    }

    @Override
    public int getFollowerCount(Integer userId) {
        // 获取粉丝数 - 查询所有用户中关注了userId的数量
        List<User> allUsers = userMapper.findAll();
        int count = 0;
        for (User user : allUsers) {
            if (isFollowing(user.getId(), userId)) {
                count++;
            }
        }
        return count;
    }

    @Override
    public int getFollowingCount(Integer userId) {
        // 获取关注数 - 查询userId的关注列表数量
        List<UserInteraction> interactions = userInteractionMapper.findByUserAndInteractionType(
                userId, TARGET_TYPE_USER, InteractionType.FOLLOW.getCode());
        return interactions.size();
    }
}
