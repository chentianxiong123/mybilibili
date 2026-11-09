package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.User;
import com.mybilibili.common.entity.UserInteraction;
import com.mybilibili.common.enums.InteractionType;
import com.mybilibili.common.vo.FanVO;
import com.mybilibili.common.vo.FansStatsVO;
import com.mybilibili.web.mapper.UserInteractionMapper;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.CreatorFansService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.text.SimpleDateFormat;
import java.util.ArrayList;
import java.util.List;

/**
 * 创作者粉丝管理服务实现
 */
@Service
public class CreatorFansServiceImpl implements CreatorFansService {

    @Autowired
    private UserInteractionMapper userInteractionMapper;

    @Autowired
    private UserMapper userMapper;

    private static final String TARGET_TYPE_USER = "USER";

    private static final SimpleDateFormat DATE_FORMAT = new SimpleDateFormat("yyyy-MM-dd HH:mm:ss");

    @Override
    public List<FanVO> getFansList(Integer userId, Integer page, Integer size, Boolean mutual) {
        int offset = (page - 1) * size;
        List<UserInteraction> interactions;
        List<FanVO> fanVOs = new ArrayList<>();

        if (Boolean.TRUE.equals(mutual)) {
            // 查询互关粉丝
            interactions = userInteractionMapper.findMutualFollowers(
                    userId, TARGET_TYPE_USER, InteractionType.FOLLOW.getCode(), offset, size);
        } else {
            // 查询所有粉丝
            interactions = userInteractionMapper.findFollowersByTargetId(
                    userId, TARGET_TYPE_USER, InteractionType.FOLLOW.getCode(), offset, size);
        }

        for (UserInteraction interaction : interactions) {
            User fan = userMapper.findById(interaction.getUserId());
            if (fan != null) {
                FanVO fanVO = new FanVO();
                fanVO.setId(fan.getId());
                fanVO.setUsername(fan.getUsername());
                fanVO.setNickname(fan.getNickname());
                fanVO.setAvatar(fan.getAvatar());
                fanVO.setLevel(fan.getLevel());
                fanVO.setSignature(fan.getSignature());

                // 检查是否互关（当前用户是否关注了该粉丝）
                boolean isMutual = checkMutualFollow(userId, fan.getId());
                fanVO.setIsMutual(isMutual);

                // 设置关注时间
                if (interaction.getCreatedAt() != null) {
                    fanVO.setFollowedAt(DATE_FORMAT.format(interaction.getCreatedAt()));
                }

                fanVOs.add(fanVO);
            }
        }

        return fanVOs;
    }

    @Override
    public FansStatsVO getFansStats(Integer userId) {
        FansStatsVO statsVO = new FansStatsVO();

        // 获取总粉丝数
        int totalFans = userInteractionMapper.countFollowersByTargetId(
                userId, TARGET_TYPE_USER, InteractionType.FOLLOW.getCode());
        statsVO.setTotalFans(totalFans);

        // 获取互关粉丝数
        int mutualFans = userInteractionMapper.countMutualFollowers(
                userId, TARGET_TYPE_USER, InteractionType.FOLLOW.getCode());
        statsVO.setMutualFans(mutualFans);

        // 近7天新增粉丝数和近30天新增粉丝数
        // 由于 user_interactions 表有 created_at 字段，可以统计
        // 这里暂时使用简化实现，实际应该添加按时间范围统计的方法
        statsVO.setNewFansWeek(0);
        statsVO.setNewFansMonth(0);

        return statsVO;
    }

    /**
     * 检查是否互相关注
     *
     * @param userId 当前用户ID
     * @param fanId  粉丝ID
     * @return 是否互关
     */
    private boolean checkMutualFollow(Integer userId, Integer fanId) {
        // 检查当前用户是否关注了该粉丝
        UserInteraction interaction = userInteractionMapper.findByUserAndTarget(
                userId, TARGET_TYPE_USER, fanId, InteractionType.FOLLOW.getCode());
        return interaction != null;
    }
}
