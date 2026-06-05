package com.mybilibili.message.controller;

import com.mybilibili.common.dto.MessageSettingDTO;
import com.mybilibili.common.dto.SendMessageDTO;
import com.mybilibili.common.vo.*;
import jakarta.validation.Valid;
import com.mybilibili.message.service.ConversationService;
import com.mybilibili.message.service.MessageService;
import com.mybilibili.message.service.MessageSettingService;
import com.mybilibili.message.service.UserLookupPort;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/message")
public class MessageController {

    @Autowired
    private MessageService messageService;

    @Autowired
    private ConversationService conversationService;

    @Autowired
    private MessageSettingService messageSettingService;

    @Autowired
    private UserLookupPort userLookupPort;

    @GetMapping("/conversations")
    public Result<List<ConversationVO>> getConversations(@RequestHeader("X-User-Id") Integer userId) {
        List<ConversationVO> conversations = conversationService.getConversationsByUserId(userId);
        return Result.success("获取成功", conversations);
    }

    @GetMapping("/conversations/{id}")
    public Result<ConversationVO> getConversationDetail(
            @PathVariable Long id,
            @RequestHeader("X-User-Id") Integer userId) {
        ConversationVO conversation = conversationService.getConversationById(id);
        if (conversation == null || !conversation.getUserId().equals(userId)) {
            return Result.error("会话不存在或无权访问");
        }
        return Result.success("获取成功", conversation);
    }

    @DeleteMapping("/conversations/{id}")
    public Result<?> deleteConversation(
            @PathVariable Long id,
            @RequestHeader("X-User-Id") Integer userId) {
        ConversationVO conversation = conversationService.getConversationById(id);
        if (conversation == null || !conversation.getUserId().equals(userId)) {
            return Result.error("会话不存在或无权操作");
        }
        conversationService.deleteConversation(id);
        return Result.success("删除成功", null);
    }

    @GetMapping("/conversations/{conversationId}/messages")
    public Result<List<MessageVO>> getConversationMessages(
            @PathVariable Long conversationId,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestHeader("X-User-Id") Integer userId) {
        ConversationVO conversation = conversationService.getConversationById(conversationId);
        if (conversation == null || !conversation.getUserId().equals(userId)) {
            return Result.error("会话不存在或无权访问");
        }
        List<MessageVO> messages = messageService.getMessagesBetweenUsers(userId, conversation.getTargetUserId(), page, size);
        return Result.success("获取成功", messages);
    }

    @PostMapping("/send")
    public Result<MessageVO> sendMessage(
            @Valid @RequestBody SendMessageDTO dto,
            @RequestHeader("X-User-Id") Integer senderId) {
        if (senderId.equals(dto.getReceiverId())) {
            return Result.error("不能给自己发送消息");
        }

        if (dto.getMessageType() != null && dto.getMessageType() != 1) {
            return Result.error("非法消息类型，私聊消息只允许messageType=1");
        }
        dto.setMessageType(1);

        Result<UserVO> userResult = userLookupPort.getUserById(dto.getReceiverId());
        if (userResult.getData() == null) {
            return Result.error("接收者不存在");
        }

        MessageVO message = messageService.sendMessage(senderId, dto);
        return Result.success("发送成功", message);
    }

    @GetMapping("/unread/counts")
    public Result<Map<String, Integer>> getUnreadCounts(@RequestHeader("X-User-Id") Integer userId) {
        Map<String, Integer> counts = messageService.getUnreadCounts(userId);
        return Result.success("获取成功", counts);
    }

    @GetMapping("/replies")
    public Result<List<ReplyMessageVO>> getReplies(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestHeader("X-User-Id") Integer userId) {
        List<ReplyMessageVO> replies = messageService.getReplies(userId, page, size);
        return Result.success("获取成功", replies);
    }

    @GetMapping("/at")
    public Result<List<AtMessageVO>> getAtList(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestHeader("X-User-Id") Integer userId) {
        List<AtMessageVO> atList = messageService.getAtList(userId, page, size);
        return Result.success("获取成功", atList);
    }

    @GetMapping("/likes")
    public Result<List<LikeMessageVO>> getLikes(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestHeader("X-User-Id") Integer userId) {
        List<LikeMessageVO> likes = messageService.getLikes(userId, page, size);
        return Result.success("获取成功", likes);
    }

    @GetMapping("/system")
    public Result<List<SystemNotificationMessageVO>> getSystemNotifications(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            @RequestHeader("X-User-Id") Integer userId) {
        List<SystemNotificationMessageVO> notifications = messageService.getSystemNotifications(userId, page, size);
        return Result.success("获取成功", notifications);
    }

    @PutMapping("/{id}/read")
    public Result<?> markMessageAsRead(
            @PathVariable Long id,
            @RequestHeader("X-User-Id") Integer userId) {
        MessageVO message = messageService.getMessageById(id);
        if (message == null || !message.getReceiverId().equals(userId)) {
            return Result.error("消息不存在或无权操作");
        }
        messageService.markMessageAsRead(id);
        return Result.success("标记成功", null);
    }

    @PutMapping("/batch/read")
    public Result<?> batchMarkMessagesAsRead(
            @RequestBody Map<String, List<Long>> requestBody,
            @RequestHeader("X-User-Id") Integer userId) {
        List<Long> ids = requestBody.get("ids");
        if (ids == null || ids.isEmpty()) {
            return Result.error("消息ID列表不能为空");
        }
        for (Long id : ids) {
            MessageVO message = messageService.getMessageById(id);
            if (message == null || !message.getReceiverId().equals(userId)) {
                return Result.error("消息不存在或无权操作");
            }
        }
        messageService.batchMarkMessagesAsRead(ids);
        return Result.success("批量标记成功", null);
    }

    @DeleteMapping("/{id}")
    public Result<?> deleteMessage(
            @PathVariable Long id,
            @RequestHeader("X-User-Id") Integer userId) {
        MessageVO message = messageService.getMessageById(id);
        if (message == null || !message.getReceiverId().equals(userId)) {
            return Result.error("消息不存在或无权操作");
        }
        messageService.deleteMessage(id);
        return Result.success("删除成功", null);
    }

    @GetMapping("/settings")
    public Result<MessageSettingVO> getMessageSettings(@RequestHeader("X-User-Id") Integer userId) {
        MessageSettingVO settings = messageSettingService.getSettingsByUserId(userId);
        if (settings == null) {
            messageSettingService.createDefaultSettings(userId);
            settings = messageSettingService.getSettingsByUserId(userId);
        }
        return Result.success("获取成功", settings);
    }

    @PutMapping("/settings")
    public Result<?> updateMessageSettings(
            @RequestBody MessageSettingDTO dto,
            @RequestHeader("X-User-Id") Integer userId) {
        messageSettingService.updateSettings(userId, dto);
        return Result.success("更新成功", null);
    }

    @PostMapping("/notify/like")
    public Result<?> sendLikeNotification(
            @RequestParam Integer senderId,
            @RequestParam Integer receiverId,
            @RequestParam Integer videoId,
            @RequestParam String videoTitle) {
        messageService.sendLikeNotification(senderId, receiverId, videoId, videoTitle);
        return Result.success("发送成功", null);
    }

    @PostMapping("/notify/system")
    public Result<?> sendSystemNotification(
            @RequestParam Integer userId,
            @RequestParam String title,
            @RequestParam String content) {
        messageService.sendSystemNotification(userId, title, content);
        return Result.success("发送成功", null);
    }

    @PostMapping("/internal/reply-notify")
    public Result<?> sendReplyNotification(
            @RequestParam Integer senderId,
            @RequestParam Integer receiverId,
            @RequestParam String content,
            @RequestParam Integer messageType,
            @RequestParam(required = false) Integer targetId,
            @RequestParam(required = false) Integer commentId) {
        messageService.sendReplyNotification(senderId, receiverId, content, messageType, targetId, commentId);
        return Result.success("发送成功", null);
    }

    @PostMapping("/internal/comment-like-notify")
    public Result<?> sendCommentLikeNotification(
            @RequestParam Integer senderId,
            @RequestParam Integer receiverId,
            @RequestParam Integer commentId,
            @RequestParam String commentContent) {
        messageService.sendCommentLikeNotification(senderId, receiverId, commentId, commentContent);
        return Result.success("发送成功", null);
    }

    @PostMapping("/admin/system/broadcast")
    public Result<?> broadcastSystemNotification(@RequestBody Map<String, String> body) {
        String content = body.get("content");
        if (content == null || content.trim().isEmpty()) {
            return Result.error("通知内容不能为空");
        }
        messageService.sendSystemNotificationToAll(content.trim());
        return Result.success("全站系统通知发送成功", null);
    }
}
