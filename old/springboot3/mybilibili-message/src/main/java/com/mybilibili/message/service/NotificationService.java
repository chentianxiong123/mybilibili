package com.mybilibili.message.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.vo.Result;
import com.mybilibili.message.entity.Notification;
import com.mybilibili.message.mapper.NotificationMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class NotificationService {

    @Autowired
    private NotificationMapper notificationMapper;

    public Result<Page<Notification>> getNotifications(Integer userId, Integer page, Integer size) {
        if (userId == null) {
            throw new BusinessException("请先登录");
        }
        Page<Notification> pageParam = new Page<>(page, size);
        LambdaQueryWrapper<Notification> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(Notification::getUserId, userId)
               .eq(Notification::getStatus, 0)
               .orderByDesc(Notification::getCreatedAt);
        Page<Notification> result = notificationMapper.selectPage(pageParam, wrapper);
        return Result.success(result);
    }

    public Result<Void> create(Notification notification) {
        notification.setIsRead(0);
        notification.setStatus(0);
        notificationMapper.insert(notification);
        return Result.success();
    }

    public Result<Void> markAsRead(Long id) {
        Notification notification = new Notification();
        notification.setId(id);
        notification.setIsRead(1);
        notificationMapper.updateById(notification);
        return Result.success();
    }

    public Result<Integer> getUnreadCount(Integer userId) {
        if (userId == null) {
            return Result.success(0);
        }
        LambdaQueryWrapper<Notification> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(Notification::getUserId, userId)
               .eq(Notification::getIsRead, 0)
               .eq(Notification::getStatus, 0);
        Long count = notificationMapper.selectCount(wrapper);
        return Result.success(count.intValue());
    }
}
