package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.LevelService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/level")
@Tag(name = "等级相关接口", description = "用户等级和经验值相关操作")
public class LevelController {

    @Autowired
    private LevelService levelService;

    @GetMapping("/info")
    @Operation(summary = "获取用户等级信息", description = "获取当前用户的等级和经验值信息")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getLevelInfo(HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            int level = levelService.getUserLevel(currentUserId);
            int experience = levelService.getUserExperience(currentUserId);
            
            Map<String, Object> data = new HashMap<>();
            data.put("level", level);
            data.put("experience", experience);
            
            return Result.success("获取成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/experience/add")
    @Operation(summary = "添加经验值", description = "为用户添加经验值")
    @SecurityRequirement(name = "JWT")
    public Result<?> addExperience(@RequestParam int experience, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            levelService.addExperience(currentUserId, experience);
            return Result.success("添加成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}