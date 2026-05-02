package com.mybilibili.user.controller;

import com.mybilibili.common.dto.UserPrivacySettingsDTO;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.UserPrivacySettingsVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.user.service.UserPrivacySettingsService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/user/privacy")
@Tag(name = "用户隐私设置接口", description = "管理用户隐私设置")
public class UserPrivacySettingsController {

    @Autowired
    private UserPrivacySettingsService privacySettingsService;

    @GetMapping("/settings")
    @Operation(summary = "获取隐私设置", description = "获取当前用户的隐私设置")
    @SecurityRequirement(name = "JWT")
    public Result<UserPrivacySettingsVO> getPrivacySettings(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return privacySettingsService.getPrivacySettings(userId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/settings")
    @Operation(summary = "更新隐私设置", description = "更新当前用户的隐私设置")
    @SecurityRequirement(name = "JWT")
    public Result<?> updatePrivacySettings(
            @RequestBody UserPrivacySettingsDTO dto,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return privacySettingsService.updatePrivacySettings(userId, dto);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/tags")
    @Operation(summary = "获取个人标签", description = "获取当前用户的个人标签列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<String>> getUserTags(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return privacySettingsService.getUserTags(userId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/tags")
    @Operation(summary = "添加个人标签", description = "为当前用户添加一个个人标签")
    @SecurityRequirement(name = "JWT")
    public Result<?> addUserTag(
            @RequestParam String tagName,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return privacySettingsService.addUserTag(userId, tagName);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/tags")
    @Operation(summary = "删除个人标签", description = "删除当前用户的一个个人标签")
    @SecurityRequirement(name = "JWT")
    public Result<?> removeUserTag(
            @RequestParam String tagName,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            return privacySettingsService.removeUserTag(userId, tagName);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
