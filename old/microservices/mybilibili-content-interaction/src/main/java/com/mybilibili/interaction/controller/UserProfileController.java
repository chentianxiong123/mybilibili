package com.mybilibili.interaction.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.interaction.entity.UserProfile;
import com.mybilibili.interaction.service.UserProfileService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/profile")
public class UserProfileController {

    @Autowired
    private UserProfileService userProfileService;

    @GetMapping("/{userId}")
    public Result<UserProfile> getProfile(@PathVariable Integer userId) {
        UserProfile profile = userProfileService.getOrCreateProfile(userId);
        return Result.success("获取成功", profile);
    }

    @PostMapping("/init/{userId}")
    public Result<?> initProfile(@PathVariable Integer userId, @RequestBody Map<String, Object> body) {
        @SuppressWarnings("unchecked")
        java.util.List<String> tags = (java.util.List<String>) body.get("tags");
        userProfileService.initProfileWithTags(userId, tags);
        return Result.success("初始化成功", null);
    }

    @PostMapping("/record/watch")
    public Result<?> recordWatch(
            @RequestHeader("X-User-Id") Integer userId,
            @RequestBody Map<String, Object> body) {
        Integer categoryId = body.get("categoryId") != null ? ((Number) body.get("categoryId")).intValue() : null;
        @SuppressWarnings("unchecked")
        java.util.List<String> tags = (java.util.List<String>) body.get("tags");
        Integer duration = body.get("durationSeconds") != null ? ((Number) body.get("durationSeconds")).intValue() : null;
        userProfileService.recordWatch(userId, categoryId, tags, duration);
        return Result.success("记录成功", null);
    }

    @PostMapping("/record/like")
    public Result<?> recordLike(
            @RequestHeader("X-User-Id") Integer userId,
            @RequestBody Map<String, Object> body) {
        Integer categoryId = body.get("categoryId") != null ? ((Number) body.get("categoryId")).intValue() : null;
        @SuppressWarnings("unchecked")
        java.util.List<String> tags = (java.util.List<String>) body.get("tags");
        userProfileService.recordLike(userId, categoryId, tags);
        return Result.success("记录成功", null);
    }

    @PostMapping("/record/collect")
    public Result<?> recordCollect(
            @RequestHeader("X-User-Id") Integer userId,
            @RequestBody Map<String, Object> body) {
        Integer categoryId = body.get("categoryId") != null ? ((Number) body.get("categoryId")).intValue() : null;
        @SuppressWarnings("unchecked")
        java.util.List<String> tags = (java.util.List<String>) body.get("tags");
        userProfileService.recordCollect(userId, categoryId, tags);
        return Result.success("记录成功", null);
    }
}
