package com.mybilibili.interaction.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.utils.FileUploadUtils;
import com.mybilibili.common.vo.DynamicVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.interaction.service.DynamicService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.servlet.http.HttpServletRequest;
import java.util.*;

@RestController
@RequestMapping("/dynamic")
@Tag(name = "用户动态相关接口", description = "发布动态、获取动态列表、点赞和分享动态")
public class DynamicController {

    @Autowired
    private DynamicService dynamicService;

    @Autowired
    private FileUploadUtils fileUploadUtils;

    @GetMapping("/list")
    @Operation(summary = "获取全部动态列表", description = "获取全部动态列表（分页）")
    @SecurityRequirement(name = "JWT")
    public Result<List<DynamicVO>> getAllDynamicList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.getAllDynamicList(page, size, currentUserId);
        } catch (Exception e) {
            return dynamicService.getAllDynamicList(page, size, null);
        }
    }

    @GetMapping("/following")
    @Operation(summary = "获取关注用户动态列表", description = "获取当前用户关注的所有用户的动态列表")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getFollowingDynamicList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) Integer userId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            Result<List<DynamicVO>> result = dynamicService.getFollowingDynamicList(currentUserId, page, size, userId);

            Map<String, Object> response = new HashMap<>();
            response.put("list", result.getData());
            response.put("page", page);
            response.put("size", size);

            return Result.success(result.getMessage(), response);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{id}")
    @Operation(summary = "获取用户动态列表", description = "获取指定用户的动态列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<DynamicVO>> getUserDynamicList(
            @PathVariable Integer id,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer limit,
            HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.getUserDynamicList(id, page, limit, currentUserId);
        } catch (Exception e) {
            return dynamicService.getUserDynamicList(id, page, limit, null);
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取动态详情", description = "获取单条动态详情")
    @SecurityRequirement(name = "JWT")
    public Result<DynamicVO> getDynamicById(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.getDynamicById(id, currentUserId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/publish")
    @Operation(summary = "发布动态", description = "发布新的动态")
    @SecurityRequirement(name = "JWT")
    public Result<DynamicVO> publishDynamic(
            @RequestParam(required = false) String content,
            @RequestParam(required = false) Integer refVideoId,
            @RequestParam(required = false) List<MultipartFile> images,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);

            List<String> imageUrls = new ArrayList<>();
            if (images != null && !images.isEmpty()) {
                for (MultipartFile image : images) {
                    if (image != null && !image.isEmpty()) {
                        String url = fileUploadUtils.uploadCover(image);
                        imageUrls.add(url);
                    }
                }
            }

            return dynamicService.publishDynamic(userId, content, refVideoId, imageUrls);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除动态", description = "删除自己的动态")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteDynamic(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.deleteDynamic(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/like/{id}")
    @Operation(summary = "点赞动态", description = "点赞指定的动态")
    @SecurityRequirement(name = "JWT")
    public Result<?> likeDynamic(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.likeDynamic(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/like/{id}")
    @Operation(summary = "取消点赞动态", description = "取消点赞指定的动态")
    @SecurityRequirement(name = "JWT")
    public Result<?> unlikeDynamic(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.unlikeDynamic(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/like/status/{id}")
    @Operation(summary = "检查点赞状态", description = "检查当前用户是否已点赞该动态")
    @SecurityRequirement(name = "JWT")
    public Result<Boolean> checkLikeStatus(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            boolean isLiked = dynamicService.isLiked(currentUserId, "DYNAMIC", id);
            return Result.success("获取成功", isLiked);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/share/{id}")
    @Operation(summary = "分享动态", description = "分享指定的动态")
    @SecurityRequirement(name = "JWT")
    public Result<?> shareDynamic(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            return dynamicService.shareDynamic(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/comment/count/{id}")
    @Operation(summary = "更新评论数量", description = "增加或减少评论数量")
    public Result<?> incrementCommentCount(@PathVariable Integer id, @RequestParam Integer delta) {
        try {
            return dynamicService.incrementCommentCount(id, delta);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
