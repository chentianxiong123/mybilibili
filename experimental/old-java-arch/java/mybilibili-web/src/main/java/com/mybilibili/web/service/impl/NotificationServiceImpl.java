package com.mybilibili.web.service.impl;

import com.mybilibili.common.dto.NotificationDTO;
import com.mybilibili.common.entity.Notification;
import com.mybilibili.common.entity.NotificationSetting;
import com.mybilibili.common.vo.NotificationVO;
import com.mybilibili.web.mapper.NotificationMapper;
import com.mybilibili.web.service.NotificationService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
public class NotificationServiceImpl implements NotificationService {

    @Autowired
    private NotificationMapper notificationMapper;

    @Override
    public void sendNotification(NotificationDTO notificationDTO) {
        // 检查用户通知设置
        NotificationSetting setting = notificationMapper.selectSettingByUserId(notificationDTO.getUserId());
        if (setting == null) {
            // 默认开启所有通知
            setting = new NotificationSetting();
            setting.setUserId(notificationDTO.getUserId());
            setting.setLikeNotification(1);
            setting.setCommentNotification(1);
            setting.setFollowNotification(1);
            setting.setSystemNotification(1);
            setting.setVideoNotification(1);
            notificationMapper.insertOrUpdateSetting(setting);
        }

        // 根据通知类型检查是否需要发送
        boolean shouldSend = false;
        switch (notificationDTO.getType()) {
            case 1: // 互动通知
                // 这里可以根据具体互动类型进一步判断
                shouldSend = true;
                break;
            case 2: // 系统通知
                shouldSend = setting.getSystemNotification() == 1;
                break;
            case 3: // 私信通知
                shouldSend = true; // 私信通知默认开启
                break;
            case 4: // 视频通知
                shouldSend = setting.getVideoNotification() == 1;
                break;
        }

        if (shouldSend) {
            Notification notification = new Notification();
            notification.setUserId(notificationDTO.getUserId());
            notification.setType(notificationDTO.getType());
            notification.setTitle(notificationDTO.getTitle());
            notification.setContent(notificationDTO.getContent());
            notification.setRelatedId(notificationDTO.getRelatedId());
            notification.setIsRead(0); // 初始为未读
            notificationMapper.insert(notification);
        }
    }

    @Override
    public List<NotificationVO> getNotifications(Integer userId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        List<Notification> notifications = notificationMapper.selectByUserId(userId, offset, size);
        List<NotificationVO> notificationVOs = new ArrayList<>();

        for (Notification notification : notifications) {
            NotificationVO vo = new NotificationVO();
            vo.setId(notification.getId());
            vo.setType(notification.getType());
            vo.setTitle(notification.getTitle());
            vo.setContent(notification.getContent());
            vo.setRelatedId(notification.getRelatedId());
            vo.setIsRead(notification.getIsRead());
            vo.setCreatedAt(notification.getCreatedAt());
            vo.setReadAt(notification.getReadAt());
            vo.setTypeName(getTypeName(notification.getType()));
            notificationVOs.add(vo);
        }

        return notificationVOs;
    }

    @Override
    public int getUnreadCount(Integer userId) {
        return notificationMapper.countUnreadByUserId(userId);
    }

    @Override
    public void markAsRead(Integer id, Integer userId) {
        notificationMapper.updateReadStatus(id, userId);
    }

    @Override
    public void batchMarkAsRead(List<Integer> ids, Integer userId) {
        if (!ids.isEmpty()) {
            notificationMapper.batchUpdateReadStatus(ids, userId);
        }
    }

    @Override
    public void deleteNotification(Integer id, Integer userId) {
        notificationMapper.delete(id, userId);
    }

    @Override
    public NotificationSetting getNotificationSetting(Integer userId) {
        NotificationSetting setting = notificationMapper.selectSettingByUserId(userId);
        if (setting == null) {
            // 默认开启所有通知
            setting = new NotificationSetting();
            setting.setUserId(userId);
            setting.setLikeNotification(1);
            setting.setCommentNotification(1);
            setting.setFollowNotification(1);
            setting.setSystemNotification(1);
            setting.setVideoNotification(1);
            notificationMapper.insertOrUpdateSetting(setting);
        }
        return setting;
    }

    @Override
    public void updateNotificationSetting(NotificationSetting setting) {
        notificationMapper.insertOrUpdateSetting(setting);
    }

    private String getTypeName(Integer type) {
        switch (type) {
            case 1:
                return "互动通知";
            case 2:
                return "系统通知";
            case 3:
                return "私信通知";
            case 4:
                return "视频通知";
            default:
                return "未知通知";
        }
    }
}
