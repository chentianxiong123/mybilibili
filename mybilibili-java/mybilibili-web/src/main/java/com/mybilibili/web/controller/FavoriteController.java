package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.FavoriteFolder;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.web.service.VideoInteractionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import javax.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/favorite")
@Tag(name = "收藏夹管理")
public class FavoriteController {

    @Autowired
    private VideoInteractionService videoInteractionService;

    @GetMapping("/folders")
    @Operation(summary = "获取收藏夹列表", description = "获取当前用户的收藏夹列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<com.mybilibili.common.entity.FavoriteFolder>> getFavoriteFolders(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<com.mybilibili.common.entity.FavoriteFolder> folders = videoInteractionService.getFavoriteFolders(userId);
            return Result.success("获取成功", folders);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/folders")
    @Operation(summary = "创建收藏夹", description = "创建新的收藏夹")
    @SecurityRequirement(name = "JWT")
    public Result<com.mybilibili.common.entity.FavoriteFolder> createFavoriteFolder(@RequestBody Map<String, String> requestBody, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            String name = requestBody.get("name");
            if (name == null || name.trim().isEmpty()) {
                return Result.error("收藏夹名称不能为空");
            }
            com.mybilibili.common.entity.FavoriteFolder folder = videoInteractionService.createFavoriteFolder(userId, name);
            return Result.success("创建成功", folder);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/folders/{folderId}")
    @Operation(summary = "更新收藏夹", description = "更新收藏夹名称")
    @SecurityRequirement(name = "JWT")
    public Result<com.mybilibili.common.entity.FavoriteFolder> updateFavoriteFolder(@PathVariable Integer folderId, @RequestBody Map<String, String> requestBody, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            String name = requestBody.get("name");
            if (name == null || name.trim().isEmpty()) {
                return Result.error("收藏夹名称不能为空");
            }
            com.mybilibili.common.entity.FavoriteFolder folder = videoInteractionService.updateFavoriteFolder(userId, folderId, name);
            return Result.success("更新成功", folder);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/folders/{folderId}")
    @Operation(summary = "删除收藏夹", description = "删除指定的收藏夹")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteFavoriteFolder(@PathVariable Integer folderId, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean result = videoInteractionService.deleteFavoriteFolder(userId, folderId);
            if (result) {
                return Result.success("删除成功");
            } else {
                return Result.error("删除失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/folders/{folderId}/manuscripts")
    @Operation(summary = "获取收藏夹内稿件", description = "获取指定收藏夹的稿件列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<com.mybilibili.common.vo.VideoVO>> getFavoriteFolderManuscripts(@PathVariable Integer folderId,
                                                                                       @RequestParam(defaultValue = "1") Integer page,
                                                                                       @RequestParam(defaultValue = "12") Integer size,
                                                                                       HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<com.mybilibili.common.vo.VideoVO> videos = videoInteractionService.getFavoriteFolderVideos(userId, folderId, page, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/folders/{folderId}/manuscripts/{manuscriptId}")
    @Operation(summary = "从收藏夹移除稿件", description = "从指定收藏夹中移除稿件")
    @SecurityRequirement(name = "JWT")
    public Result<?> removeManuscriptFromFavoriteFolder(@PathVariable Integer folderId, @PathVariable Integer manuscriptId, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean result = videoInteractionService.removeVideoFromFavoriteFolder(userId, manuscriptId, folderId);
            if (result) {
                return Result.success("移除成功");
            } else {
                return Result.error("移除失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
