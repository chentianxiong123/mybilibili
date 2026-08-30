package com.mybilibili.interaction.controller;

import com.mybilibili.common.entity.FavoriteFolder;
import com.mybilibili.common.vo.InteractionResponse;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.interaction.feign.VideoClient;
import com.mybilibili.interaction.service.InteractionService;
import com.mybilibili.interaction.service.VideoInteractionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/manuscript")
@Tag(name = "互动接口", description = "稿件互动相关操作，包括点赞、投币、收藏、分享等")
public class InteractionController {

    @Autowired
    private InteractionService interactionService;

    @Autowired
    private VideoInteractionService videoInteractionService;

    @Autowired
    private VideoClient videoClient;

    @PostMapping("/{id}/like")
    @Operation(summary = "点赞稿件", description = "对指定稿件进行点赞操作")
    public Result<InteractionResponse> likeVideo(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            boolean result = videoInteractionService.likeVideo(userId, id);
            int count = interactionService.getLikeCount("MANUSCRIPT", id);
            InteractionResponse response = new InteractionResponse(true, count, result ? "like" : "already_liked");
            return Result.success(result ? "点赞成功" : "已经点赞过该稿件", response);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}/like")
    @Operation(summary = "取消点赞稿件", description = "取消对指定稿件的点赞操作")
    public Result<InteractionResponse> unlikeVideo(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            boolean result = videoInteractionService.unlikeVideo(userId, id);
            int count = interactionService.getLikeCount("MANUSCRIPT", id);
            InteractionResponse response = new InteractionResponse(false, count, result ? "unlike" : "not_liked");
            return Result.success(result ? "取消点赞成功" : "还没有点赞该稿件", response);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/coin")
    @Operation(summary = "投币稿件", description = "对指定稿件进行投币操作")
    public Result<?> coinVideo(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "投币数量") @RequestParam Integer coinCount,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            boolean result = videoInteractionService.coinVideo(userId, id, coinCount);
            if (result) {
                return Result.success("投币成功");
            } else {
                return Result.error("投币失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/collect")
    @Operation(summary = "收藏稿件", description = "对指定稿件进行收藏操作")
    public Result<InteractionResponse> collectVideo(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            boolean result = videoInteractionService.collectVideo(userId, id);
            InteractionResponse response = new InteractionResponse(true, 0, result ? "collect" : "already_collected");
            return Result.success(result ? "收藏成功" : "已经收藏过该稿件", response);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}/collect")
    @Operation(summary = "取消收藏稿件", description = "取消对指定稿件的收藏操作")
    public Result<InteractionResponse> uncollectVideo(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            boolean result = videoInteractionService.uncollectVideo(userId, id);
            InteractionResponse response = new InteractionResponse(false, 0, result ? "uncollect" : "not_collected");
            return Result.success(result ? "取消收藏成功" : "还没有收藏该稿件", response);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/share")
    @Operation(summary = "分享稿件", description = "对指定稿件进行分享操作")
    public Result<?> shareVideo(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "分享渠道") @RequestParam(required = false) String channel,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        System.out.println("[SHARE] controller called, manuscriptId=" + id + ", userId=" + userId + ", channel=" + channel);
        try {
            videoInteractionService.shareVideo(userId, id, channel, null);
            System.out.println("[SHARE] shareVideo completed successfully");
            return Result.success("分享成功");
        } catch (Exception e) {
            System.out.println("[SHARE] shareVideo failed: " + e.getMessage());
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/share/statistics")
    @Operation(summary = "获取分享统计", description = "获取稿件的分享统计数据")
    public Result<Map<String, Object>> getShareStatistics(@Parameter(description = "稿件ID") @PathVariable Integer id) {
        try {
            Map<String, Object> statistics = videoInteractionService.getShareStatistics(id);
            return Result.success("获取成功", statistics);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/status")
    @Operation(summary = "获取互动状态", description = "获取用户对指定稿件的互动状态")
    public Result<VideoInteractionService.VideoInteractionStatus> getInteractionStatus(
            @Parameter(description = "稿件ID") @PathVariable Integer id,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            VideoInteractionService.VideoInteractionStatus status = videoInteractionService.getInteractionStatus(userId, id);
            return Result.success("获取成功", status);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/likes")
    @Operation(summary = "获取点赞列表", description = "获取用户点赞的稿件列表")
    public Result<List<VideoVO>> getLikedVideos(@Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<VideoVO> videos = videoInteractionService.getLikedVideos(userId);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/collections")
    @Operation(summary = "获取收藏列表", description = "获取用户收藏的稿件列表")
    public Result<List<VideoVO>> getCollectedVideos(@Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<VideoVO> videos = videoInteractionService.getCollectedVideos(userId);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/favorite/folders")
    @Operation(summary = "获取收藏夹列表", description = "获取用户的收藏夹列表")
    public Result<List<FavoriteFolder>> getFavoriteFolders(@Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<FavoriteFolder> folders = videoInteractionService.getFavoriteFolders(userId);
            return Result.success("获取成功", folders);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/favorite/folders")
    @Operation(summary = "创建收藏夹", description = "创建新的收藏夹")
    public Result<FavoriteFolder> createFavoriteFolder(@Parameter(description = "收藏夹名称") @RequestBody Map<String, String> requestBody,
                                                       @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            String name = requestBody.get("name");
            if (name == null || name.trim().isEmpty()) {
                return Result.error("收藏夹名称不能为空");
            }
            FavoriteFolder folder = videoInteractionService.createFavoriteFolder(userId, name);
            return Result.success("创建成功", folder);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/favorite/folders/{folderId}")
    @Operation(summary = "更新收藏夹", description = "更新收藏夹名称")
    public Result<FavoriteFolder> updateFavoriteFolder(@Parameter(description = "收藏夹ID") @PathVariable Integer folderId,
                                                       @Parameter(description = "收藏夹名称") @RequestBody Map<String, String> requestBody,
                                                       @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            String name = requestBody.get("name");
            if (name == null || name.trim().isEmpty()) {
                return Result.error("收藏夹名称不能为空");
            }
            FavoriteFolder folder = videoInteractionService.updateFavoriteFolder(userId, folderId, name);
            return Result.success("更新成功", folder);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/favorite/folders/{folderId}")
    @Operation(summary = "删除收藏夹", description = "删除指定的收藏夹")
    public Result<?> deleteFavoriteFolder(@Parameter(description = "收藏夹ID") @PathVariable Integer folderId,
                                          @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
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

    @PostMapping("/{id}/favorite")
    @Operation(summary = "添加稿件到收藏夹", description = "将稿件添加到指定的收藏夹")
    public Result<?> addVideoToFavoriteFolders(@Parameter(description = "稿件ID") @PathVariable Integer id,
                                              @Parameter(description = "收藏夹ID列表") @RequestBody Map<String, List<Integer>> requestBody,
                                              @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<Integer> folderIds = requestBody.get("folderIds");
            if (folderIds == null || folderIds.isEmpty()) {
                return Result.error("请选择收藏夹");
            }
            boolean result = videoInteractionService.addVideoToFavoriteFolders(userId, id, folderIds);
            if (result) {
                return Result.success("添加成功");
            } else {
                return Result.error("添加失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/favorite/folders/{folderId}/videos/{videoId}")
    @Operation(summary = "从收藏夹移除稿件", description = "从收藏夹中移除指定稿件")
    public Result<?> removeVideoFromFavoriteFolder(@Parameter(description = "收藏夹ID") @PathVariable Integer folderId,
                                                   @Parameter(description = "稿件ID") @PathVariable Integer videoId,
                                                   @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            boolean result = videoInteractionService.removeVideoFromFavoriteFolder(userId, videoId, folderId);
            if (result) {
                return Result.success("移除成功");
            } else {
                return Result.error("移除失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/favorite/folders/{folderId}/videos")
    @Operation(summary = "获取收藏夹稿件列表", description = "获取指定收藏夹中的稿件列表")
    public Result<List<VideoVO>> getFavoriteFolderVideos(@Parameter(description = "收藏夹ID") @PathVariable Integer folderId,
                                                         @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
                                                         @Parameter(description = "每页数量") @RequestParam(defaultValue = "12") Integer size,
                                                         @Parameter(description = "排序: desc-最新收藏, asc-最早收藏") @RequestParam(defaultValue = "desc") String sortOrder,
                                                         @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<VideoVO> videos = videoInteractionService.getFavoriteFolderVideos(userId, folderId, page, size, sortOrder);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/favorite/folders")
    @Operation(summary = "获取稿件的收藏夹", description = "获取指定稿件所在的收藏夹列表")
    public Result<List<FavoriteFolder>> getVideoFavoriteFolders(@Parameter(description = "稿件ID") @PathVariable Integer id,
                                                                @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<FavoriteFolder> folders = videoInteractionService.getVideoFavoriteFolders(userId, id);
            return Result.success("获取成功", folders);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}/favorite/folders")
    @Operation(summary = "更新稿件的收藏夹", description = "更新稿件的收藏夹列表")
    public Result<?> updateVideoFavoriteFolders(@Parameter(description = "稿件ID") @PathVariable Integer id,
                                                @Parameter(description = "收藏夹ID列表") @RequestBody Map<String, List<Integer>> requestBody,
                                                @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            List<Integer> folderIds = requestBody.get("folderIds");
            boolean result = videoInteractionService.updateVideoFavoriteFolders(userId, id, folderIds);
            if (result) {
                return Result.success("更新成功");
            } else {
                return Result.error("更新失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
