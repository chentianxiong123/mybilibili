package com.mybilibili.web.controller;

import com.mybilibili.common.dto.CreatorSettingsDTO;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.CreatorSettingsVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.CreatorSettingsService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;

@RestController
@RequestMapping("/creator")
@Tag(name = "创作者设置接口", description = "创作者相关设置操作")
public class CreatorSettingsController {

    @Autowired
    private CreatorSettingsService creatorSettingsService;

    @GetMapping("/settings")
    @Operation(summary = "获取创作设置", description = "获取当前用户的创作者设置")
    @SecurityRequirement(name = "JWT")
    public Result<CreatorSettingsVO> getSettings(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            CreatorSettingsVO settings = creatorSettingsService.getSettingsByUserId(userId);
            if (settings == null) {
                creatorSettingsService.createDefaultSettings(userId);
                settings = creatorSettingsService.getSettingsByUserId(userId);
            }
            return Result.success("获取成功", settings);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/settings")
    @Operation(summary = "更新创作设置", description = "更新当前用户的创作者设置")
    @SecurityRequirement(name = "JWT")
    public Result<CreatorSettingsVO> updateSettings(HttpServletRequest request, @RequestBody CreatorSettingsDTO dto) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            creatorSettingsService.updateSettings(userId, dto);
            CreatorSettingsVO settings = creatorSettingsService.getSettingsByUserId(userId);
            return Result.success("更新成功", settings);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
