package com.mybilibili.message.service.impl;

import com.mybilibili.common.dto.SendMessageDTO;
import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.entity.Message;
import com.mybilibili.common.entity.MessageSetting;
import com.mybilibili.common.vo.*;
import com.mybilibili.message.mapper.ConversationMapper;
import com.mybilibili.message.mapper.MessageMapper;
import com.mybilibili.message.service.ConversationService;
import com.mybilibili.message.service.MessageService;
import com.mybilibili.message.service.MessageSettingService;
import com.mybilibili.message.websocket.NotificationPushService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class MessageServiceImpl implements MessageService {

    @Autowired
    private MessageMapper messageMapper;

    @Autowired
    private ConversationMapper conversationMapper;

    @Autowired
    private ConversationService conversationService;

    @Autowired
    private MessageSettingService messageSettingService;

    @Autowired
    private NotificationPushService notificationPushService;

    private void pushNotificationAndCounts(Integer receiverId, String notificationType, String content) {
        try {
            notificationPushService.pushToUser(receiverId, notificationType, notificationType, content);
            Map<String, Integer> counts = getUnreadCounts(receiverId);
            notificationPushService.pushUnreadCount(receiverId, counts);
        } catch (Exception ignored) {
        }
    }

    private boolean isNotificationEnabled(Integer receiverId, Integer messageType) {
        try {
            MessageSetting setting = messageSettingService.getOrCreateSettings(receiverId);
            if (setting == null) return true;
            if (messageType == 2) return Boolean.TRUE.equals(setting.getReplyNotify());
            if (messageType == 3) return Boolean.TRUE.equals(setting.getAtNotify());
            if (messageType == 4 || messageType == 6) return Boolean.TRUE.equals(setting.getLikeNotify());
            if (messageType == 5) return Boolean.TRUE.equals(setting.getSystemNotify());
            return true;
        } catch (Exception e) {
            return true;
        }
    }

    @Override
    public void sendMessage(Message message) {
        message.setIsRead(false);
        messageMapper.insert(message);
    }

    @Override
    @Transactional
    public MessageVO sendMessage(Integer senderId, SendMessageDTO dto) {
        Conversation senderConversation = conversationService.getOrCreateConversation(senderId, dto.getReceiverId());
        Conversation receiverConversation = conversationService.getOrCreateConversation(dto.getReceiverId(), senderId);

        Message message = new Message();
        message.setSenderId(senderId);
        message.setReceiverId(dto.getReceiverId());
        message.setContent(dto.getContent());
        message.setMessageType(dto.getMessageType());
        message.setTargetId(dto.getTargetId());
        message.setMediaUrl(dto.getMediaUrl());
        message.setCommentId(dto.getCommentId());
        message.setConversationId(senderConversation.getId());
        message.setIsRead(false);
        messageMapper.insert(message);

        senderConversation.setLastMessageId(message.getId());
        senderConversation.setLastMessageContent(dto.getContent());
        senderConversation.setLastMessageTime(new Date());
        conversationMapper.update(senderConversation);

        receiverConversation.setLastMessageId(message.getId());
        receiverConversation.setLastMessageContent(dto.getContent());
        receiverConversation.setLastMessageTime(new Date());
        receiverConversation.setUnreadCount(receiverConversation.getUnreadCount() + 1);
        conversationMapper.update(receiverConversation);

        Map<String, Object> extra = new HashMap<>();
        extra.put("messageId", message.getId());
        extra.put("senderId", senderId);
        extra.put("receiverId", dto.getReceiverId());
        extra.put("conversationId", senderConversation.getId());
        extra.put("messageType", dto.getMessageType());
        notificationPushService.pushToUser(dto.getReceiverId(), "private", "新私信", dto.getContent(), extra);
        Map<String, Integer> counts = getUnreadCounts(dto.getReceiverId());
        notificationPushService.pushUnreadCount(dto.getReceiverId(), counts);

        return getMessageById(message.getId());
    }

    @Override
    public MessageVO getMessageById(Long id) {
        return messageMapper.selectVOById(id);
    }

    @Override
    public List<MessageVO> getMessagesByUserId(Integer userId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectByReceiverId(userId, offset, size);
    }

    @Override
    public List<MessageVO> getMessagesByConversationId(Long conversationId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectByConversationId(conversationId, offset, size);
    }

    @Override
    public List<MessageVO> getMessagesBetweenUsers(Integer userId, Integer targetUserId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectBetweenUsers(userId, targetUserId, offset, size);
    }

    @Override
    public void markMessageAsRead(Long id) {
        messageMapper.updateReadStatus(id, true);
    }

    @Override
    @Transactional
    public void batchMarkMessagesAsRead(List<Long> ids) {
        if (!ids.isEmpty()) {
            MessageVO firstMessage = messageMapper.selectVOById(ids.get(0));
            if (firstMessage != null) {
                Integer receiverId = firstMessage.getReceiverId();
                Integer senderId = firstMessage.getSenderId();

                messageMapper.batchUpdateReadStatus(ids, true);

                Conversation conversation = conversationMapper.selectByUserAndTarget(receiverId, senderId);
                if (conversation != null) {
                    int unreadCount = Math.max(0, conversation.getUnreadCount() - ids.size());
                    conversation.setUnreadCount(unreadCount);
                    conversationMapper.update(conversation);
                }
            }
        }
    }

    @Override
    public void deleteMessage(Long id) {
        messageMapper.deleteById(id);
    }

    @Override
    public void clearMessagesByUserId(Integer userId) {
        messageMapper.deleteByReceiverId(userId);
    }

    @Override
    public int getUnreadMessageCount(Integer userId) {
        return messageMapper.countUnreadByReceiverId(userId);
    }

    @Override
    public List<ReplyMessageVO> getReplies(Integer userId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectReplies(userId, offset, size);
    }

    @Override
    public List<AtMessageVO> getAtList(Integer userId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectAtList(userId, offset, size);
    }

    @Override
    public List<LikeMessageVO> getLikes(Integer userId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectLikes(userId, offset, size);
    }

    @Override
    public List<SystemNotificationMessageVO> getSystemNotifications(Integer userId, Integer page, Integer size) {
        int offset = (page - 1) * size;
        return messageMapper.selectSystemNotifications(userId, offset, size);
    }

    @Override
    public Map<String, Integer> getUnreadCounts(Integer userId) {
        Map<String, Integer> counts = new HashMap<>();
        counts.put("reply", messageMapper.countUnreadReplies(userId));
        counts.put("at", messageMapper.countUnreadAt(userId));
        counts.put("like", messageMapper.countUnreadLikes(userId));
        counts.put("system", messageMapper.countUnreadSystem(userId));
        counts.put("private", messageMapper.countUnreadByReceiverId(userId));
        counts.put("total", counts.values().stream().mapToInt(Integer::intValue).sum());
        return counts;
    }

    @Override
    public void sendLikeNotification(Integer senderId, Integer receiverId, Integer videoId, String videoTitle) {
        if (!isNotificationEnabled(receiverId, 4)) return;
        Message message = new Message();
        message.setSenderId(senderId);
        message.setReceiverId(receiverId);
        message.setContent("赞了你的视频《" + videoTitle + "》");
        message.setMessageType(4);
        message.setTargetId(videoId);
        message.setIsRead(false);
        messageMapper.insert(message);
        pushNotificationAndCounts(receiverId, "like", message.getContent());
    }

    @Override
    public void sendSystemNotification(Integer userId, String title, String content) {
        if (!isNotificationEnabled(userId, 5)) return;
        Message message = new Message();
        message.setSenderId(1);
        message.setReceiverId(userId);
        message.setContent(content);
        message.setMessageType(5);
        message.setIsRead(false);
        messageMapper.insert(message);
        pushNotificationAndCounts(userId, "system", content);
    }

    @Override
    public void sendReplyNotification(Integer senderId, Integer receiverId, String content, Integer messageType, Integer targetId, Integer commentId) {
        if (!isNotificationEnabled(receiverId, messageType)) return;
        Message message = new Message();
        message.setSenderId(senderId);
        message.setReceiverId(receiverId);
        message.setContent(content);
        message.setMessageType(messageType);
        message.setTargetId(targetId);
        message.setCommentId(commentId);
        message.setIsRead(false);
        messageMapper.insert(message);
        pushNotificationAndCounts(receiverId, "reply", content);
    }

    @Override
    public void sendCommentLikeNotification(Integer senderId, Integer receiverId, Integer commentId, String commentContent) {
        if (senderId.equals(receiverId)) {
            return;
        }
        if (!isNotificationEnabled(receiverId, 6)) return;
        Message message = new Message();
        message.setSenderId(senderId);
        message.setReceiverId(receiverId);
        String preview = commentContent != null && commentContent.length() > 30
                ? commentContent.substring(0, 30) + "..."
                : commentContent;
        message.setContent("赞了你的评论\"" + preview + "\"");
        message.setMessageType(6);
        message.setCommentId(commentId);
        message.setIsRead(false);
        messageMapper.insert(message);
        pushNotificationAndCounts(receiverId, "like", message.getContent());
    }

    @Override
    public void sendSystemNotificationToAll(String content) {
        List<Integer> userIds = messageMapper.selectUserIdsWithSystemNotifyEnabled();
        if (userIds == null || userIds.isEmpty()) return;
        for (Integer userId : userIds) {
            try {
                Message message = new Message();
                message.setSenderId(1);
                message.setReceiverId(userId);
                message.setContent(content);
                message.setMessageType(5);
                message.setIsRead(false);
                messageMapper.insert(message);
                pushNotificationAndCounts(userId, "system", content);
            } catch (Exception e) {
                org.slf4j.LoggerFactory.getLogger(getClass()).warn("向用户 {} 发送系统通知失败: {}", userId, e.getMessage());
            }
        }
    }
}