package com.mybilibili.search.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoRecommendVO;
import com.mybilibili.search.service.VideoRecommendService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/recommend")
@Tag(name = "视频推荐", description = "视频推荐相关接口")
public class VideoRecommendController {

    @Autowired
    private VideoRecommendService videoRecommendService;

    private static final int DEFAULT_SIZE = 10;
    private static final int MAX_SIZE = 50;

    @GetMapping("/related/{videoId}")
    @Operation(summary = "相关视频推荐", description = "基于当前视频获取相似内容推荐")
    public Result<List<VideoRecommendVO>> getRelatedVideos(
            @Parameter(description = "视频ID") @PathVariable("videoId") Integer videoId,
            @Parameter(description = "推荐数量") @RequestParam(value = "size", defaultValue = "10") int size) {

        try {
            if (videoId == null || videoId <= 0) {
                return Result.error("视频ID不能为空");
            }

            size = Math.max(1, Math.min(size, MAX_SIZE));

            List<VideoRecommendVO> videos = videoRecommendService.getRelatedVideos(videoId, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error("获取相关视频失败: " + e.getMessage());
        }
    }

    @GetMapping("/hot")
    @Operation(summary = "热门视频推荐", description = "获取热门视频，支持按分类过滤")
    public Result<List<VideoRecommendVO>> getHotVideos(
            @Parameter(description = "分类ID") @RequestParam(value = "categoryId", required = false) Integer categoryId,
            @Parameter(description = "推荐数量") @RequestParam(value = "size", defaultValue = "10") int size) {

        try {
            size = Math.max(1, Math.min(size, MAX_SIZE));

            List<VideoRecommendVO> videos = videoRecommendService.getHotVideos(categoryId, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error("获取热门视频失败: " + e.getMessage());
        }
    }

    @GetMapping("/for-you")
    @Operation(summary = "个性化推荐", description = "基于用户浏览历史获取个性化推荐（需要登录）")
    public Result<List<VideoRecommendVO>> getRecommendedVideosForUser(
            @Parameter(description = "推荐数量") @RequestParam(value = "size", defaultValue = "10") int size,
            HttpServletRequest request) {

        try {
            Integer userId = getUserIdFromRequest(request);

            if (userId == null) {
                return Result.error("请先登录");
            }

            size = Math.max(1, Math.min(size, MAX_SIZE));

            List<VideoRecommendVO> videos = videoRecommendService.getRecommendedVideosForUser(userId, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error("获取个性化推荐失败: " + e.getMessage());
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