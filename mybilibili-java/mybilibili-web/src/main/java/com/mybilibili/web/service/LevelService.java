package com.mybilibili.web.service;

public interface LevelService {
    // 计算用户等级
    int calculateLevel(int experience);
    
    // 添加用户经验值
    void addExperience(Integer userId, int experience);
    
    // 获取用户等级信息
    int getUserLevel(Integer userId);
    
    // 获取用户经验值
    int getUserExperience(Integer userId);
}