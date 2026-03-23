package com.mybilibili.web.service.impl;

import com.mybilibili.common.entity.User;
import com.mybilibili.web.mapper.UserMapper;
import com.mybilibili.web.service.LevelService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class LevelServiceImpl implements LevelService {

    @Autowired
    private UserMapper userMapper;

    // 等级经验值对照表
    private static final int[] LEVEL_EXPERIENCE = {
        0,      // 0级
        100,    // 1级
        300,    // 2级
        600,    // 3级
        1000,   // 4级
        1500,   // 5级
        2100,   // 6级
        2800,   // 7级
        3600,   // 8级
        4500,   // 9级
        5500    // 10级
    };

    @Override
    public int calculateLevel(int experience) {
        for (int i = LEVEL_EXPERIENCE.length - 1; i >= 0; i--) {
            if (experience >= LEVEL_EXPERIENCE[i]) {
                return i;
            }
        }
        return 0;
    }

    @Override
    public void addExperience(Integer userId, int experience) {
        // 更新经验值
        userMapper.updateExperience(userId, experience);
        
        // 获取用户信息
        User user = userMapper.findById(userId);
        if (user != null) {
            // 计算新等级
            int newLevel = calculateLevel(user.getExperience() + experience);
            
            // 如果等级提升，更新等级
            if (newLevel > user.getLevel()) {
                user.setLevel(newLevel);
                userMapper.update(user);
            }
        }
    }

    @Override
    public int getUserLevel(Integer userId) {
        User user = userMapper.findById(userId);
        return user != null ? user.getLevel() : 0;
    }

    @Override
    public int getUserExperience(Integer userId) {
        User user = userMapper.findById(userId);
        return user != null ? user.getExperience() : 0;
    }
}