package com.mybilibili.user.service;

import com.mybilibili.user.entity.LoginLog;

import java.util.List;

public interface LoginLogService {
    void saveLog(Integer userId, String ip, String userAgent, String location, Integer status);
    List<LoginLog> getUserLogs(Integer userId, Integer page, Integer size);
    Integer countUserLogs(Integer userId);
    List<LoginLog> getLogsByCondition(String ip, Integer userId, Integer status, String startTime, String endTime, Integer page, Integer size);
    Integer countByCondition(String ip, Integer userId, Integer status, String startTime, String endTime);
}