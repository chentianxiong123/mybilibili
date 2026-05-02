package com.mybilibili.message.service;

import com.mybilibili.common.dto.SendMessageDTO;
import com.mybilibili.common.entity.Message;
import com.mybilibili.common.vo.*;

import java.util.List;
import java.util.Map;

public interface MessageService {

    void sendMessage(Message message);

    MessageVO sendMessage(Integer senderId, SendMessageDTO dto);

    MessageVO getMessageById(Long id);

    List<MessageVO> getMessagesByUserId(Integer userId, Integer page, Integer size);

    List<MessageVO> getMessagesByConversationId(Long conversationId, Integer page, Integer size);

    List<MessageVO> getMessagesBetweenUsers(Integer userId, Integer targetUserId, Integer page, Integer size);

    void markMessageAsRead(Long id);

    void batchMarkMessagesAsRead(List<Long> ids);

    void deleteMessage(Long id);

    void clearMessagesByUserId(Integer userId);

    int getUnreadMessageCount(Integer userId);

    List<ReplyMessageVO> getReplies(Integer userId, Integer page, Integer size);

    List<AtMessageVO> getAtList(Integer userId, Integer page, Integer size);

    List<LikeMessageVO> getLikes(Integer userId, Integer page, Integer size);

    List<SystemNotificationMessageVO> getSystemNotifications(Integer userId, Integer page, Integer size);

    Map<String, Integer> getUnreadCounts(Integer userId);

    void sendLikeNotification(Integer senderId, Integer receiverId, Integer videoId, String videoTitle);

    void sendSystemNotification(Integer userId, String title, String content);

    void sendReplyNotification(Integer senderId, Integer receiverId, String content, Integer messageType, Integer targetId, Integer commentId);

    void sendCommentLikeNotification(Integer senderId, Integer receiverId, Integer commentId, String commentContent);

    void sendSystemNotificationToAll(String content);
}