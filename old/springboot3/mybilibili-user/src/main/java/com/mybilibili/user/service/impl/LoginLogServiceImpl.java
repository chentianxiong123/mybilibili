package com.mybilibili.user.service.impl;

import com.mybilibili.user.entity.LoginLog;
import com.mybilibili.user.mapper.LoginLogMapper;
import com.mybilibili.user.service.LoginLogService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class LoginLogServiceImpl implements LoginLogService {

    @Autowired
    private LoginLogMapper loginLogMapper;

    @Override
    public void saveLog(Integer userId, String ip, String userAgent, String location, Integer status) {
        LoginLog log = new LoginLog();
        log.setUserId(userId);
        log.setIp(ip);
        log.setUserAgent(userAgent);
        log.setLocation(location);
        log.setStatus(status);
        loginLogMapper.insert(log);
    }

    @Override
    public List<LoginLog> getUserLogs(Integer userId, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        Integer offset = (page - 1) * size;
        return loginLogMapper.selectByUserId(userId, offset, size);
    }

    @Override
    public Integer countUserLogs(Integer userId) {
        return loginLogMapper.countByUserId(userId);
    }

    @Override
    public List<LoginLog> getLogsByCondition(String ip, Integer userId, Integer status, String startTime, String endTime, Integer page, Integer size) {
        if (page == null || page < 1) page = 1;
        if (size == null || size < 1) size = 10;
        Integer offset = (page - 1) * size;
        return loginLogMapper.selectByCondition(ip, userId, status, startTime, endTime, offset, size);
    }

    @Override
    public Integer countByCondition(String ip, Integer userId, Integer status, String startTime, String endTime) {
        return loginLogMapper.countByCondition(ip, userId, status, startTime, endTime);
    }
}