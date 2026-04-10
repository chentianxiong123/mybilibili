package com.mybilibili.message.service.impl;

import com.mybilibili.common.dto.SendMessageDTO;
import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.entity.Message;
import com.mybilibili.common.vo.*;
import com.mybilibili.message.mapper.ConversationMapper;
import com.mybilibili.message.mapper.MessageMapper;
import com.mybilibili.message.service.ConversationService;
import com.mybilibili.message.service.MessageService;
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
        Message message = new Message();
        message.setSenderId(senderId);
        message.setReceiverId(receiverId);
        message.setContent("赞了你的视频《" + videoTitle + "》");
        message.setMessageType(4);
        message.setTargetId(videoId);
        message.setIsRead(false);
        messageMapper.insert(message);
    }

    @Override
    public void sendSystemNotification(Integer userId, String title, String content) {
        Message message = new Message();
        message.setSenderId(1);
        message.setReceiverId(userId);
        message.setContent(content);
        message.setMessageType(5);
        message.setIsRead(false);
        messageMapper.insert(message);
    }

    @Override
    public void sendReplyNotification(Integer senderId, Integer receiverId, String content, Integer messageType, Integer targetId, Integer commentId) {
        Message message = new Message();
        message.setSenderId(senderId);
        message.setReceiverId(receiverId);
        message.setContent(content);
        message.setMessageType(messageType);
        message.setTargetId(targetId);
        message.setCommentId(commentId);
        message.setIsRead(false);
        messageMapper.insert(message);
    }
}