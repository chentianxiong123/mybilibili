package com.mybilibili.web.controller;

import com.mybilibili.common.dto.UserPrivacySettingsDTO;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserPrivacySettingsVO;
import com.mybilibili.web.service.UserPrivacyService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/user/privacy")
@Tag(name = "用户隐私设置接口", description = "用户隐私设置相关操作")
@SecurityRequirement(name = "JWT")
public class UserPrivacyController {

    @Autowired
    private UserPrivacyService userPrivacyService;

    @GetMapping("/settings")
    @Operation(summary = "获取隐私设置", description = "获取当前用户的隐私设置")
    public Result<UserPrivacySettingsVO> getPrivacySettings(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            UserPrivacySettingsVO vo = userPrivacyService.getPrivacySettings(userId);
            return Result.success("获取成功", vo);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/settings")
    @Operation(summary = "更新隐私设置", description = "更新当前用户的隐私设置")
    public Result<Void> updatePrivacySettings(@RequestBody UserPrivacySettingsDTO dto, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            userPrivacyService.updatePrivacySettings(userId, dto);
            return Result.success("更新成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/tags")
    @Operation(summary = "获取个人标签", description = "获取当前用户的个人标签列表")
    public Result<List<String>> getUserTags(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            List<String> tags = userPrivacyService.getUserTags(userId);
            return Result.success("获取成功", tags);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/tags")
    @Operation(summary = "添加个人标签", description = "添加个人标签")
    public Result<Void> addUserTag(@RequestParam String tagName, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            userPrivacyService.addUserTag(userId, tagName);
            return Result.success("添加成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/tags")
    @Operation(summary = "删除个人标签", description = "删除个人标签")
    public Result<Void> removeUserTag(@RequestParam String tagName, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            userPrivacyService.removeUserTag(userId, tagName);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
