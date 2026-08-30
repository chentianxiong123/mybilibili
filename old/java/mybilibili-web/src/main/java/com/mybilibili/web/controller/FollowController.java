package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.web.service.FollowService;
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
@RequestMapping("/")
@Tag(name = "关注相关接口", description = "用户关注、取消关注、获取关注列表和粉丝列表")
public class FollowController {

    @Autowired
    private FollowService followService;

    @PostMapping("/follow/{id}")
    @Operation(summary = "关注用户", description = "关注指定用户，如果已关注则取消关注")
    @SecurityRequirement(name = "JWT")
    public Result<?> followUser(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            System.out.println("【调试】关注/取消关注请求 - 当前用户: " + currentUserId + ", 目标用户: " + id);
            // 检查是否已关注
            boolean isFollowing = followService.isFollowing(currentUserId, id);
            System.out.println("【调试】当前关注状态: " + isFollowing);
            if (isFollowing) {
                // 已关注，取消关注
                System.out.println("【调试】执行取消关注操作");
                boolean result = followService.unfollowUser(currentUserId, id);
                System.out.println("【调试】取消关注结果: " + result);
                return Result.success("取消关注成功");
            } else {
                // 未关注，添加关注
                System.out.println("【调试】执行添加关注操作");
                boolean result = followService.followUser(currentUserId, id);
                System.out.println("【调试】添加关注结果: " + result);
                return Result.success("关注成功");
            }
        } catch (Exception e) {
            System.out.println("【调试】关注操作异常: " + e.getMessage());
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/follow/{id}")
    @Operation(summary = "取消关注", description = "取消关注指定用户")
    @SecurityRequirement(name = "JWT")
    public Result<?> unfollowUser(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            boolean result = followService.unfollowUser(currentUserId, id);
            if (result) {
                return Result.success("取消关注成功");
            } else {
                return Result.error("还没有关注该用户");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{id}/following")
    @Operation(summary = "获取关注列表", description = "获取指定用户的关注列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<UserVO>> getFollowingList(@PathVariable Integer id) {
        try {
            List<UserVO> followingList = followService.getFollowingList(id);
            return Result.success("获取成功", followingList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{id}/followers")
    @Operation(summary = "获取粉丝列表", description = "获取指定用户的粉丝列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<UserVO>> getFollowerList(@PathVariable Integer id) {
        try {
            List<UserVO> followerList = followService.getFollowerList(id);
            return Result.success("获取成功", followerList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/follow/check/{id}")
    @Operation(summary = "检查关注状态", description = "检查当前用户是否已关注指定用户")
    @SecurityRequirement(name = "JWT")
    public Result<?> checkFollowStatus(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            boolean isFollowing = followService.isFollowing(currentUserId, id);
            Map<String, Boolean> result = new HashMap<>();
            result.put("isFollowing", isFollowing);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/follow/following/simple")
    @Operation(summary = "获取当前用户关注列表（简要）", description = "获取当前登录用户的关注列表，用于动态页筛选")
    @SecurityRequirement(name = "JWT")
    public Result<List<UserVO>> getCurrentUserFollowingList(HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            List<UserVO> followingList = followService.getFollowingList(currentUserId);
            return Result.success("获取成功", followingList);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}