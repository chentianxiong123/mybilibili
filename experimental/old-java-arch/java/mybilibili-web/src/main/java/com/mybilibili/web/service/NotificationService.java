package com.mybilibili.web.service;

import com.mybilibili.common.dto.NotificationDTO;
import com.mybilibili.common.entity.NotificationSetting;
import com.mybilibili.common.vo.NotificationVO;

import java.util.List;

public interface NotificationService {
    // 发送通知
    void sendNotification(NotificationDTO notificationDTO);

    // 获取用户通知列表
    List<NotificationVO> getNotifications(Integer userId, Integer page, Integer size);

    // 获取未读通知数量
    int getUnreadCount(Integer userId);

    // 标记通知为已读
    void markAsRead(Integer id, Integer userId);

    // 批量标记通知为已读
    void batchMarkAsRead(List<Integer> ids, Integer userId);

    // 删除通知
    void deleteNotification(Integer id, Integer userId);

    // 获取通知设置
    NotificationSetting getNotificationSetting(Integer userId);

    // 更新通知设置
    void updateNotificationSetting(NotificationSetting setting);
}
