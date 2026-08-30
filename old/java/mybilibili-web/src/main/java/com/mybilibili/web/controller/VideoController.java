package com.mybilibili.web.controller;

import com.mybilibili.common.dto.VideoUploadDTO;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.VideoVO;
import com.mybilibili.web.service.VideoService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.servlet.http.HttpServletRequest;
import java.util.List;

@RestController
@RequestMapping("/video")
@Tag(name = "视频相关接口", description = "视频上传、获取视频列表等操作")
public class VideoController {

    @Autowired
    private VideoService videoService;

    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    @Operation(summary = "上传视频", description = "上传新视频")
    @SecurityRequirement(name = "JWT")
    public Result<VideoVO> uploadVideo(
            @RequestParam("title") String title,
            @RequestParam("description") String description,
            @RequestParam("cover") MultipartFile cover,
            @RequestParam("video") MultipartFile video,
            @RequestParam("categoryId") Integer categoryId,
            @RequestParam(value = "tags", required = false) List<String> tags,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            
            // 创建VideoUploadDTO对象
            VideoUploadDTO videoUploadDTO = new VideoUploadDTO();
            videoUploadDTO.setTitle(title);
            videoUploadDTO.setDescription(description);
            videoUploadDTO.setCover(cover);
            videoUploadDTO.setVideo(video);
            videoUploadDTO.setCategoryId(categoryId);
            videoUploadDTO.setTags(tags);
            
            VideoVO videoVO = videoService.uploadVideo(videoUploadDTO, userId);
            return Result.success("上传成功", videoVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取视频详情", description = "根据视频ID获取视频详情")
    public Result<VideoVO> getVideoById(@PathVariable Integer id, HttpServletRequest request) {
        try {
            // 尝试从JWT中获取当前用户ID
            Integer currentUserId = null;
            try {
                String authHeader = request.getHeader("Authorization");
                if (authHeader != null && authHeader.startsWith("Bearer ")) {
                    // 去除Bearer前缀
                    String token = authHeader.substring(7);
                    currentUserId = JwtUtils.getUserIdFromToken(token);
                }
            } catch (Exception e) {
                // 未登录或token无效，currentUserId保持为null
                System.out.println("获取当前用户ID失败: " + e.getMessage());
            }
            System.out.println("获取视频详情，视频ID: " + id + ", 当前用户ID: " + currentUserId);
            VideoVO videoVO = videoService.getVideoById(id, currentUserId);
            if (videoVO == null) {
                return Result.error("视频不存在");
            }
            // 更新观看次数
            videoService.updateViewCount(id);
            return Result.success("获取成功", videoVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/recommended")
    @Operation(summary = "获取推荐视频", description = "获取推荐视频列表")
    public Result<List<VideoVO>> getRecommendedVideos() {
        try {
            List<VideoVO> videos = videoService.getRecommendedVideos();
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/category/{id}")
    @Operation(summary = "获取分类视频", description = "根据分类ID获取视频列表")
    public Result<List<VideoVO>> getVideosByCategoryId(@PathVariable Integer id) {
        try {
            List<VideoVO> videos = videoService.getVideosByCategoryId(id);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/hot")
    @Operation(summary = "获取热门视频", description = "获取热门视频列表")
    public Result<List<VideoVO>> getHotVideos() {
        try {
            List<VideoVO> videos = videoService.getHotVideos();
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{id}")
    @Operation(summary = "获取用户视频", description = "根据用户ID获取视频列表，支持排序")
    public Result<List<VideoVO>> getVideosByUserId(
            @PathVariable Integer id,
            @RequestParam(value = "sort", required = false, defaultValue = "latest") String sort) {
        try {
            System.out.println("【调试】接收到获取用户视频请求，用户ID: " + id + ", 排序方式: " + sort);
            List<VideoVO> videos = videoService.getVideosByUserId(id, sort);
            System.out.println("【调试】返回视频数量: " + videos.size());
            if (!videos.isEmpty()) {
                VideoVO first = videos.get(0);
                System.out.println("【调试】第一个视频 - ID: " + first.getId() + 
                    ", 标题: " + first.getTitle() + 
                    ", 播放数: " + first.getViewCount() + 
                    ", 收藏数: " + first.getCollectCount());
            }
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            System.err.println("【调试】获取用户视频失败: " + e.getMessage());
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/list")
    @Operation(summary = "获取视频列表", description = "获取视频列表，支持分页")
    public Result<List<VideoVO>> getVideoList(
            @RequestParam(value = "page", defaultValue = "1") Integer page,
            @RequestParam(value = "size", defaultValue = "10") Integer size) {
        try {
            List<VideoVO> videos = videoService.getVideoList(page, size);
            return Result.success("获取成功", videos);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping(value = "/{id}", consumes = "multipart/form-data")
    @Operation(summary = "编辑视频", description = "编辑视频信息")
    @SecurityRequirement(name = "JWT")
    public Result<VideoVO> updateVideo(
            @PathVariable Integer id,
            @RequestParam(value = "title", required = false) String title,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam(value = "cover", required = false) MultipartFile cover,
            @RequestParam(value = "video", required = false) MultipartFile video,
            @RequestParam(value = "categoryId", required = false) Integer categoryId,
            @RequestParam(value = "tags", required = false) List<String> tags,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            
            // 创建VideoUploadDTO对象
            VideoUploadDTO videoUploadDTO = new VideoUploadDTO();
            videoUploadDTO.setTitle(title);
            videoUploadDTO.setDescription(description);
            videoUploadDTO.setCover(cover);
            videoUploadDTO.setVideo(video);
            videoUploadDTO.setCategoryId(categoryId);
            videoUploadDTO.setTags(tags);
            
            VideoVO videoVO = videoService.updateVideo(id, videoUploadDTO, userId);
            return Result.success("更新成功", videoVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除视频", description = "删除自己的视频")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteVideo(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            
            videoService.deleteVideo(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/manuscript/{manuscriptId}")
    @Operation(summary = "根据稿件ID获取视频", description = "获取稿件的第p个视频，用于稿件详情页")
    public Result<VideoVO> getVideoByManuscriptId(
            @PathVariable Integer manuscriptId,
            @RequestParam(value = "p", defaultValue = "1") Integer p,
            HttpServletRequest request) {
        try {
            // 1. 获取当前用户ID
            Integer currentUserId = null;
            try {
                String authHeader = request.getHeader("Authorization");
                if (authHeader != null && authHeader.startsWith("Bearer ")) {
                    String token = authHeader.substring(7);
                    currentUserId = JwtUtils.getUserIdFromToken(token);
                }
            } catch (Exception e) {
                // 未登录或token无效
            }

            // 2. 调用Service获取视频
            VideoVO videoVO = videoService.getVideoByManuscriptId(manuscriptId, p, currentUserId);
            if (videoVO == null) {
                return Result.error("稿件不存在或没有视频");
            }

            return Result.success("获取成功", videoVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}