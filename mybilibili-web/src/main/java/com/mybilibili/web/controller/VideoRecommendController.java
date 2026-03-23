package com.mybilibili.web.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoRecommendVO;
import com.mybilibili.web.service.VideoRecommendService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

/**
 * 视频推荐控制器
 */
@RestController
@RequestMapping("/recommend")
@Tag(name = "视频推荐", description = "视频推荐相关接口")
public class VideoRecommendController {

    @Autowired
    private VideoRecommendService videoRecommendService;

    /**
     * 默认推荐数量
     */
    private static final int DEFAULT_SIZE = 10;

    /**
     * 最大推荐数量
     */
    private static final int MAX_SIZE = 50;

    /**
     * 获取相关视频推荐
     *
     * @param videoId 当前视频ID
     * @param size    推荐数量
     * @return 相关视频列表
     */
    @GetMapping("/related/{videoId}")
    @Operation(summary = "相关视频推荐", description = "基于当前视频获取相似内容推荐")
    public Result<List<VideoRecommendVO>> getRelatedVideos(
            @Parameter(description = "视频ID") @PathVariable("videoId") Integer videoId,
            @Parameter(description = "推荐数量") @RequestParam(value = "size", defaultValue = "10") int size) {

        try {
            // 参数校验
            if (videoId == null || videoId <= 0) {
                return Result.error("视频ID不能为空");
            }

            // 限制size范围
            size = Math.max(1, Math.min(size, MAX_SIZE));

            List<VideoRecommendVO> videos = videoRecommendService.getRelatedVideos(videoId, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error("获取相关视频失败: " + e.getMessage());
        }
    }

    /**
     * 获取热门视频推荐
     *
     * @param categoryId 分类ID（可选）
     * @param size       推荐数量
     * @return 热门视频列表
     */
    @GetMapping("/hot")
    @Operation(summary = "热门视频推荐", description = "获取热门视频，支持按分类过滤")
    public Result<List<VideoRecommendVO>> getHotVideos(
            @Parameter(description = "分类ID") @RequestParam(value = "categoryId", required = false) Integer categoryId,
            @Parameter(description = "推荐数量") @RequestParam(value = "size", defaultValue = "10") int size) {

        try {
            // 限制size范围
            size = Math.max(1, Math.min(size, MAX_SIZE));

            List<VideoRecommendVO> videos = videoRecommendService.getHotVideos(categoryId, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error("获取热门视频失败: " + e.getMessage());
        }
    }

    /**
     * 获取个性化推荐视频（需要登录）
     *
     * @param size    推荐数量
     * @param request HTTP请求
     * @return 个性化推荐视频列表
     */
    @GetMapping("/for-you")
    @Operation(summary = "个性化推荐", description = "基于用户浏览历史获取个性化推荐（需要登录）")
    public Result<List<VideoRecommendVO>> getRecommendedVideosForUser(
            @Parameter(description = "推荐数量") @RequestParam(value = "size", defaultValue = "10") int size,
            HttpServletRequest request) {

        try {
            // 从请求属性中获取当前用户ID
            Integer userId = (Integer) request.getAttribute("userId");

            if (userId == null) {
                return Result.error("请先登录");
            }

            // 限制size范围
            size = Math.max(1, Math.min(size, MAX_SIZE));

            List<VideoRecommendVO> videos = videoRecommendService.getRecommendedVideosForUser(userId, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error("获取个性化推荐失败: " + e.getMessage());
        }
    }
}
