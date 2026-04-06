package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.InteractionResponse;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.FavoriteFolder;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.web.service.LikeService;
import com.mybilibili.web.service.VideoInteractionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/manuscript")
@Tag(name = "稿件互动")
public class VideoInteractionController {

    @Autowired
    private VideoInteractionService videoInteractionService;

    @Autowired
    private LikeService likeService;

    private static final String TARGET_TYPE_MANUSCRIPT = "MANUSCRIPT";

    @PostMapping("/{id}/like")
    @Operation(summary = "点赞视频", description = "点赞指定视频")
    @SecurityRequirement(name = "JWT")
    public Result<InteractionResponse> likeVideo(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean result = videoInteractionService.likeVideo(userId, id);
            if (result) {
                int count = likeService.getLikeCount(TARGET_TYPE_MANUSCRIPT, id);
                InteractionResponse response = new InteractionResponse(true, count, "like");
                return Result.success("点赞成功", response);
            } else {
                // 已经点赞过，返回当前状态
                int count = likeService.getLikeCount(TARGET_TYPE_MANUSCRIPT, id);
                InteractionResponse response = new InteractionResponse(true, count, "like");
                return Result.success("已经点赞过该视频", response);
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}/like")
    @Operation(summary = "取消点赞", description = "取消点赞指定视频")
    @SecurityRequirement(name = "JWT")
    public Result<InteractionResponse> unlikeVideo(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean result = videoInteractionService.unlikeVideo(userId, id);
            if (result) {
                int count = likeService.getLikeCount(TARGET_TYPE_MANUSCRIPT, id);
                InteractionResponse response = new InteractionResponse(false, count, "unlike");
                return Result.success("取消点赞成功", response);
            } else {
                // 还没有点赞
                int count = likeService.getLikeCount(TARGET_TYPE_MANUSCRIPT, id);
                InteractionResponse response = new InteractionResponse(false, count, "unlike");
                return Result.success("还没有点赞该视频", response);
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/coin")
    @Operation(summary = "投币支持", description = "向指定视频投币")
    @SecurityRequirement(name = "JWT")
    public Result<?> coinVideo(@PathVariable Integer id, @RequestParam Integer coinCount, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
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
    @Operation(summary = "收藏视频", description = "收藏指定视频")
    @SecurityRequirement(name = "JWT")
    public Result<InteractionResponse> collectVideo(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean result = videoInteractionService.collectVideo(userId, id);
            if (result) {
                InteractionResponse response = new InteractionResponse(true, 0, "collect");
                return Result.success("收藏成功", response);
            } else {
                InteractionResponse response = new InteractionResponse(true, 0, "collect");
                return Result.success("已经收藏过该视频", response);
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}/collect")
    @Operation(summary = "取消收藏", description = "取消收藏指定视频")
    @SecurityRequirement(name = "JWT")
    public Result<InteractionResponse> uncollectVideo(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean result = videoInteractionService.uncollectVideo(userId, id);
            if (result) {
                InteractionResponse response = new InteractionResponse(false, 0, "uncollect");
                return Result.success("取消收藏成功", response);
            } else {
                InteractionResponse response = new InteractionResponse(false, 0, "uncollect");
                return Result.success("还没有收藏该视频", response);
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/share")
    @Operation(summary = "分享视频", description = "分享指定视频")
    @SecurityRequirement(name = "JWT")
    public Result<?> shareVideo(
            @PathVariable Integer id,
            @RequestParam(required = false) String channel,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            String ipAddress = request.getRemoteAddr();
            videoInteractionService.shareVideo(userId, id, channel, ipAddress);
            return Result.success("分享成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/share/statistics")
    @Operation(summary = "获取分享统计", description = "获取指定视频的分享统计信息")
    public Result<?> getShareStatistics(@PathVariable Integer id) {
        try {
            java.util.Map<String, Object> statistics = videoInteractionService.getShareStatistics(id);
            return Result.success("获取成功", statistics);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/danmaku")
    @Operation(summary = "发送弹幕", description = "向指定视频发送弹幕")
    @SecurityRequirement(name = "JWT")
    public Result<?> sendDanmaku(
            @PathVariable Integer id,
            @RequestParam String content,
            @RequestParam String time,
            @RequestParam(required = false) String color,
            @RequestParam(required = false) Integer mode,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            videoInteractionService.sendDanmaku(userId, id, content, time, color, mode);
            return Result.success("发送弹幕成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/danmakus")
    @Operation(summary = "获取视频弹幕", description = "获取指定视频的弹幕列表")
    public Result<?> getDanmakus(@PathVariable Integer id) {
        try {
            List<?> danmakus = videoInteractionService.getDanmakus(id);
            return Result.success("获取成功", danmakus);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}/status")
    @Operation(summary = "获取互动状态", description = "获取当前用户对视频的互动状态")
    @SecurityRequirement(name = "JWT")
    public Result<?> getInteractionStatus(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            VideoInteractionService.VideoInteractionStatus status = videoInteractionService.getInteractionStatus(userId, id);
            return Result.success("获取成功", status);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/likes")
    @Operation(summary = "获取用户点赞视频", description = "获取当前用户点赞的视频列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<VideoVO>> getLikedVideos(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<VideoVO> videos = videoInteractionService.getLikedVideos(userId);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/collections")
    @Operation(summary = "获取用户收藏视频", description = "获取当前用户收藏的视频列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<VideoVO>> getCollectedVideos(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<VideoVO> videos = videoInteractionService.getCollectedVideos(userId);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
    
    @GetMapping("/user/{id}/collections")
    @Operation(summary = "获取指定用户收藏视频", description = "获取指定用户收藏的视频列表")
    @SecurityRequirement(name = "JWT")
    public Result<List<VideoVO>> getCollectedVideosByUserId(@PathVariable Integer id) {
        try {
            List<VideoVO> videos = videoInteractionService.getCollectedVideos(id);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
