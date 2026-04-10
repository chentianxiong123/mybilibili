package com.mybilibili.user.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.user.service.FollowService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/follow")
@Tag(name = "关注接口", description = "用户关注/取消关注操作")
public class FollowController {

    @Autowired
    private FollowService followService;

    @PostMapping("/{userId}")
    @Operation(summary = "关注用户", description = "关注指定用户")
    public Result<Map<String, Object>> follow(
            @Parameter(description = "目标用户ID") @PathVariable Integer userId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = getUserIdFromRequest(request);
            if (currentUserId == null) {
                return Result.error("用户未登录");
            }
            if (currentUserId.equals(userId)) {
                return Result.error("不能关注自己");
            }
            boolean success = followService.follow(currentUserId, userId);
            Map<String, Object> data = new HashMap<>();
            data.put("following", true);
            if (success) {
                return Result.success("关注成功", data);
            } else {
                return Result.success("已经关注过该用户", data);
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{userId}")
    @Operation(summary = "取消关注用户", description = "取消关注指定用户")
    public Result<Map<String, Object>> unfollow(
            @Parameter(description = "目标用户ID") @PathVariable Integer userId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = getUserIdFromRequest(request);
            if (currentUserId == null) {
                return Result.error("用户未登录");
            }
            boolean success = followService.unfollow(currentUserId, userId);
            Map<String, Object> data = new HashMap<>();
            data.put("following", false);
            if (success) {
                return Result.success("取消关注成功", data);
            } else {
                return Result.success("还没有关注该用户", data);
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/check/{userId}")
    @Operation(summary = "检查关注状态", description = "检查当前用户是否关注了指定用户")
    public Result<Map<String, Object>> checkFollow(
            @Parameter(description = "目标用户ID") @PathVariable Integer userId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = getUserIdFromRequest(request);
            Map<String, Object> data = new HashMap<>();
            if (currentUserId == null) {
                data.put("following", false);
                return Result.success("未登录", data);
            }
            boolean isFollowing = followService.isFollowing(currentUserId, userId);
            data.put("following", isFollowing);
            return Result.success("查询成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{id}/following")
    @Operation(summary = "获取用户关注列表", description = "获取指定用户的关注列表")
    public Result<List<UserVO>> getFollowingList(
            @Parameter(description = "用户ID") @PathVariable("id") Integer userId) {
        try {
            List<UserVO> followingList = followService.getFollowingList(userId);
            return Result.success("获取成功", followingList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{id}/followers")
    @Operation(summary = "获取用户粉丝列表", description = "获取指定用户的粉丝列表")
    public Result<List<UserVO>> getFollowerList(
            @Parameter(description = "用户ID") @PathVariable("id") Integer userId) {
        try {
            List<UserVO> followerList = followService.getFollowerList(userId);
            return Result.success("获取成功", followerList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/me/followers")
    @Operation(summary = "获取当前用户粉丝列表", description = "获取当前登录用户的粉丝列表")
    public Result<List<UserVO>> getMyFollowers(HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            List<UserVO> followerList = followService.getFollowerList(userId);
            return Result.success("获取成功", followerList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/me/following")
    @Operation(summary = "获取当前用户关注列表", description = "获取当前登录用户的关注列表")
    public Result<List<UserVO>> getMyFollowing(HttpServletRequest request) {
        try {
            Integer userId = getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            List<UserVO> followingList = followService.getFollowingList(userId);
            return Result.success("获取成功", followingList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    private Integer getUserIdFromRequest(HttpServletRequest request) {
        String userIdStr = request.getHeader("X-User-Id");
        if (userIdStr != null && !userIdStr.isEmpty()) {
            try {
                return Integer.parseInt(userIdStr);
            } catch (NumberFormatException e) {
                return null;
            }
        }
        return null;
    }
}