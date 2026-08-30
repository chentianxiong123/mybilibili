package com.mybilibili.web.controller;

import com.mybilibili.common.dto.MessageSettingDTO;
import com.mybilibili.common.dto.SendMessageDTO;
import com.mybilibili.common.entity.Conversation;
import com.mybilibili.common.entity.Message;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.ConversationVO;
import com.mybilibili.common.vo.MessageSettingVO;
import com.mybilibili.common.vo.MessageVO;
import com.mybilibili.common.vo.ReplyMessageVO;
import com.mybilibili.common.vo.AtMessageVO;
import com.mybilibili.common.vo.LikeMessageVO;
import com.mybilibili.common.vo.SystemNotificationMessageVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.ConversationService;
import com.mybilibili.web.service.MessageService;
import com.mybilibili.web.service.MessageSettingService;
import com.mybilibili.web.service.UserService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/message")
@Tag(name = "消息相关接口", description = "消息的查询、发送、标记已读、删除等操作")
public class MessageController {

    @Autowired
    private MessageService messageService;

    @Autowired
    private ConversationService conversationService;

    @Autowired
    private MessageSettingService messageSettingService;

    @Autowired
    private UserService userService;

    // ==================== 会话相关接口 ====================

    @GetMapping("/conversations")
    @Operation(summary = "获取会话列表", description = "获取当前用户的所有会话列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<ConversationVO>> getConversations(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<ConversationVO> conversations = conversationService.getConversationsByUserId(userId);
            return Result.success("获取成功", conversations);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/conversations/{id}")
    @Operation(summary = "获取会话详情", description = "获取指定会话的详情")
    @SecurityRequirement(name = "JWT")
    public Result<ConversationVO> getConversationDetail(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            ConversationVO conversation = conversationService.getConversationById(id);
            if (conversation == null || !conversation.getUserId().equals(userId)) {
                return Result.error("会话不存在或无权访问");
            }
            return Result.success("获取成功", conversation);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/conversations/{id}")
    @Operation(summary = "删除会话", description = "删除指定会话")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteConversation(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            ConversationVO conversation = conversationService.getConversationById(id);
            if (conversation == null || !conversation.getUserId().equals(userId)) {
                return Result.error("会话不存在或无权操作");
            }
            conversationService.deleteConversation(id);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    // ==================== 消息相关接口 ====================

    @GetMapping("/conversations/{conversationId}/messages")
    @Operation(summary = "获取会话消息", description = "获取指定会话的消息列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<MessageVO>> getConversationMessages(
            @PathVariable Integer conversationId,
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            ConversationVO conversation = conversationService.getConversationById(conversationId);
            if (conversation == null || !conversation.getUserId().equals(userId)) {
                return Result.error("会话不存在或无权访问");
            }
            // 查询当前用户和对方用户之间的所有消息
            List<MessageVO> messages = messageService.getMessagesBetweenUsers(userId, conversation.getTargetUserId(), page, size);
            return Result.success("获取成功", messages);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/send")
    @Operation(summary = "发送消息", description = "向其他用户发送消息")
    @SecurityRequirement(name = "JWT")
    public Result<MessageVO> sendMessage(
            @RequestBody SendMessageDTO dto,
            HttpServletRequest request) {
        try {
            Integer senderId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            if (senderId.equals(dto.getReceiverId())) {
                return Result.error("不能给自己发送消息");
            }

            if (userService.getUserById(dto.getReceiverId()) == null) {
                return Result.error("接收者不存在");
            }

            MessageVO message = messageService.sendMessage(senderId, dto);
            return Result.success("发送成功", message);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/unread/counts")
    @Operation(summary = "获取各类未读消息数", description = "获取当前用户各类消息的未读数量")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Integer>> getUnreadCounts(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            Map<String, Integer> counts = messageService.getUnreadCounts(userId);
            return Result.success("获取成功", counts);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/replies")
    @Operation(summary = "获取回复我的消息", description = "获取回复我的消息列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<ReplyMessageVO>> getReplies(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<ReplyMessageVO> replies = messageService.getReplies(userId, page, size);
            return Result.success("获取成功", replies);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/at")
    @Operation(summary = "获取@我的消息", description = "获取@我的消息列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<AtMessageVO>> getAtList(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<AtMessageVO> atList = messageService.getAtList(userId, page, size);
            return Result.success("获取成功", atList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/likes")
    @Operation(summary = "获取收到的赞", description = "获取收到的赞消息列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<LikeMessageVO>> getLikes(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<LikeMessageVO> likes = messageService.getLikes(userId, page, size);
            return Result.success("获取成功", likes);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/system")
    @Operation(summary = "获取系统通知", description = "获取系统通知列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<SystemNotificationMessageVO>> getSystemNotifications(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "20") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<SystemNotificationMessageVO> notifications = messageService.getSystemNotifications(userId, page, size);
            return Result.success("获取成功", notifications);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}/read")
    @Operation(summary = "标记消息为已读", description = "将指定消息标记为已读")
    @SecurityRequirement(name = "JWT")
    public Result<?> markMessageAsRead(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            MessageVO message = messageService.getMessageById(id);
            if (message == null || !message.getReceiverId().equals(userId)) {
                return Result.error("消息不存在或无权操作");
            }
            messageService.markMessageAsRead(id);
            return Result.success("标记成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/batch/read")
    @Operation(summary = "批量标记消息为已读", description = "批量将消息标记为已读")
    @SecurityRequirement(name = "JWT")
    public Result<?> batchMarkMessagesAsRead(
            @RequestBody Map<String, List<Integer>> requestBody,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<Integer> ids = requestBody.get("ids");
            if (ids == null || ids.isEmpty()) {
                return Result.error("消息ID列表不能为空");
            }
            for (Integer id : ids) {
                MessageVO message = messageService.getMessageById(id);
                if (message == null || !message.getReceiverId().equals(userId)) {
                    return Result.error("消息不存在或无权操作");
                }
            }
            messageService.batchMarkMessagesAsRead(ids);
            return Result.success("批量标记成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除消息", description = "删除指定消息")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteMessage(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            MessageVO message = messageService.getMessageById(id);
            if (message == null || !message.getReceiverId().equals(userId)) {
                return Result.error("消息不存在或无权操作");
            }
            messageService.deleteMessage(id);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    // ==================== 消息设置相关接口 ====================

    @GetMapping("/settings")
    @Operation(summary = "获取消息设置", description = "获取当前用户的消息设置")
    @SecurityRequirement(name = "JWT")
    public Result<MessageSettingVO> getMessageSettings(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            MessageSettingVO settings = messageSettingService.getSettingsByUserId(userId);
            if (settings == null) {
                messageSettingService.createDefaultSettings(userId);
                settings = messageSettingService.getSettingsByUserId(userId);
            }
            return Result.success("获取成功", settings);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/settings")
    @Operation(summary = "更新消息设置", description = "更新当前用户的消息设置")
    @SecurityRequirement(name = "JWT")
    public Result<?> updateMessageSettings(
            @RequestBody MessageSettingDTO dto,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            messageSettingService.updateSettings(userId, dto);
            return Result.success("更新成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
