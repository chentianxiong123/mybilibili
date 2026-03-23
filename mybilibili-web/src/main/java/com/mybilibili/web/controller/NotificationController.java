package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.NotificationVO;
import com.mybilibili.common.entity.NotificationSetting;
import com.mybilibili.web.service.NotificationService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/notification")
@Tag(name = "通知相关接口", description = "通知管理、设置等操作")
public class NotificationController {

    @Autowired
    private NotificationService notificationService;

    @GetMapping("/list")
    @Operation(summary = "获取通知列表", description = "获取用户的通知列表，支持分页")
    @SecurityRequirement(name = "JWT")
    public Result<List<NotificationVO>> getNotifications(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "10") Integer size,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<NotificationVO> notifications = notificationService.getNotifications(userId, page, size);
            return Result.success("获取成功", notifications);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/unread-count")
    @Operation(summary = "获取未读通知数量", description = "获取用户的未读通知数量")
    @SecurityRequirement(name = "JWT")
    public Result<Integer> getUnreadCount(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            int count = notificationService.getUnreadCount(userId);
            return Result.success("获取成功", count);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/read/{id}")
    @Operation(summary = "标记通知为已读", description = "标记单个通知为已读")
    @SecurityRequirement(name = "JWT")
    public Result<?> markAsRead(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            notificationService.markAsRead(id, userId);
            return Result.success("标记成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/read/batch")
    @Operation(summary = "批量标记通知为已读", description = "批量标记多个通知为已读")
    @SecurityRequirement(name = "JWT")
    public Result<?> batchMarkAsRead(@RequestBody List<Integer> ids, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            notificationService.batchMarkAsRead(ids, userId);
            return Result.success("标记成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除通知", description = "删除单个通知")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteNotification(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            notificationService.deleteNotification(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/setting")
    @Operation(summary = "获取通知设置", description = "获取用户的通知设置")
    @SecurityRequirement(name = "JWT")
    public Result<NotificationSetting> getNotificationSetting(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            NotificationSetting setting = notificationService.getNotificationSetting(userId);
            return Result.success("获取成功", setting);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/setting")
    @Operation(summary = "更新通知设置", description = "更新用户的通知设置")
    @SecurityRequirement(name = "JWT")
    public Result<?> updateNotificationSetting(@RequestBody NotificationSetting setting, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            setting.setUserId(userId);
            notificationService.updateNotificationSetting(setting);
            return Result.success("更新成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
