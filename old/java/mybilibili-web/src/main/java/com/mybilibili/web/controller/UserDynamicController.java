package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.DynamicVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.UserDynamic;
import com.mybilibili.web.service.LikeService;
import com.mybilibili.web.service.UserDynamicService;
import com.mybilibili.web.service.FollowService;
import com.mybilibili.web.utils.UploadFilePathUtils;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.servlet.http.HttpServletRequest;
import java.io.File;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/dynamic")
@Tag(name = "用户动态相关接口", description = "发布动态、获取动态列表、点赞和分享动态")
public class UserDynamicController {

    @Autowired
    private UserDynamicService userDynamicService;

    @Autowired
    private LikeService likeService;

    @Autowired
    private FollowService followService;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    private static final String TARGET_TYPE_DYNAMIC = "DYNAMIC";

    @PostMapping("/publish")
    @Operation(summary = "发布动态", description = "发布用户动态，支持文字、图片和引用视频")
    @SecurityRequirement(name = "JWT")
    public Result<?> publishDynamic(
            @RequestParam String content,
            @RequestParam(required = false) MultipartFile[] images,
            @RequestParam(required = false) Integer refVideoId,
            @RequestParam(required = false) Integer refManuscriptId,
            HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            
            String imageUrl = null;
            Integer dynamicType = 0;
            
            if (images != null && images.length > 0) {
                uploadFilePathUtils.createImagesDirectory();
                StringBuilder urlBuilder = new StringBuilder();
                for (int i = 0; i < images.length && i < 9; i++) {
                    MultipartFile file = images[i];
                    if (!uploadFilePathUtils.isValidImageType(file.getContentType())) {
                        continue;
                    }
                    String fileName = uploadFilePathUtils.generateImageFileName(file.getOriginalFilename());
                    String filePath = uploadFilePathUtils.getImagePath(fileName);
                    file.transferTo(new File(filePath));
                    if (urlBuilder.length() > 0) {
                        urlBuilder.append(",");
                    }
                    urlBuilder.append(uploadFilePathUtils.getImageUrl(fileName));
                }
                imageUrl = urlBuilder.toString();
                dynamicType = 1;
            }
            
            if (refVideoId != null) {
                dynamicType = 2;
            }
            
            return userDynamicService.publishDynamic(currentUserId, content, imageUrl, dynamicType, refVideoId, refManuscriptId);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/list")
    @Operation(summary = "获取全部动态列表", description = "获取全部动态列表（分页）")
    public Result<List<DynamicVO>> getAllDynamicList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size) {
        return userDynamicService.getAllDynamicList(page, size);
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
            Result<List<DynamicVO>> result = userDynamicService.getFollowingDynamicList(currentUserId, page, size, userId);
            
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
            return userDynamicService.getUserDynamicList(id, page, limit, currentUserId);
        } catch (Exception e) {
            // 如果获取当前用户失败，仍然返回动态列表，但isLiked都为false
            return userDynamicService.getUserDynamicList(id, page, limit, null);
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
            return userDynamicService.getDynamicById(id, currentUserId);
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
            return userDynamicService.deleteDynamic(currentUserId, id);
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
            return userDynamicService.likeDynamic(currentUserId, id);
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
            return userDynamicService.unlikeDynamic(currentUserId, id);
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
            boolean isLiked = likeService.isLiked(currentUserId, TARGET_TYPE_DYNAMIC, id);
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
            return userDynamicService.shareDynamic(currentUserId, id);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

}
