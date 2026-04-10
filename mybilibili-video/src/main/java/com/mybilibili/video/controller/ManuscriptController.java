package com.mybilibili.video.controller;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.ManuscriptService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.servlet.http.HttpServletRequest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/manuscript")
@Tag(name = "稿件接口", description = "稿件查询、管理接口")
public class ManuscriptController {

    @Autowired
    private ManuscriptService manuscriptService;

    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    @Operation(summary = "上传稿件", description = "上传新稿件，支持单视频和多视频分P")
    public Result<ManuscriptVO> uploadManuscript(
            @RequestParam("title") String title,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam("cover") MultipartFile cover,
            @RequestParam("categoryId") Integer categoryId,
            @RequestParam(value = "tags", required = false) List<String> tags,
            @RequestParam(value = "videos[0].file", required = false) MultipartFile video0,
            @RequestParam(value = "videos[0].title", required = false) String videoTitle0,
            @RequestParam(value = "videos[0].videoOrder", required = false) Integer videoOrder0,
            @RequestParam(value = "videos[0].durationSeconds", required = false) Integer videoDuration0,
            @RequestParam(value = "videos[1].file", required = false) MultipartFile video1,
            @RequestParam(value = "videos[1].title", required = false) String videoTitle1,
            @RequestParam(value = "videos[1].videoOrder", required = false) Integer videoOrder1,
            @RequestParam(value = "videos[1].durationSeconds", required = false) Integer videoDuration1,
            @RequestParam(value = "videos[2].file", required = false) MultipartFile video2,
            @RequestParam(value = "videos[2].title", required = false) String videoTitle2,
            @RequestParam(value = "videos[2].videoOrder", required = false) Integer videoOrder2,
            @RequestParam(value = "videos[2].durationSeconds", required = false) Integer videoDuration2,
            @RequestParam(value = "videos[3].file", required = false) MultipartFile video3,
            @RequestParam(value = "videos[3].title", required = false) String videoTitle3,
            @RequestParam(value = "videos[3].videoOrder", required = false) Integer videoOrder3,
            @RequestParam(value = "videos[3].durationSeconds", required = false) Integer videoDuration3,
            @RequestParam(value = "videos[4].file", required = false) MultipartFile video4,
            @RequestParam(value = "videos[4].title", required = false) String videoTitle4,
            @RequestParam(value = "videos[4].videoOrder", required = false) Integer videoOrder4,
            @RequestParam(value = "videos[4].durationSeconds", required = false) Integer videoDuration4,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }

            ManuscriptUploadDTO dto = new ManuscriptUploadDTO();
            dto.setTitle(title);
            dto.setDescription(description);
            dto.setCover(cover);
            dto.setCategoryId(categoryId);
            dto.setTags(tags);

            List<ManuscriptUploadDTO.VideoItemDTO> videos = new ArrayList<>();
            if (video0 != null && !video0.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video0);
                videoItem.setTitle(videoTitle0);
                videoItem.setVideoOrder(videoOrder0 != null ? videoOrder0 : 0);
                videoItem.setDurationSeconds(videoDuration0);
                videos.add(videoItem);
            }
            if (video1 != null && !video1.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video1);
                videoItem.setTitle(videoTitle1);
                videoItem.setVideoOrder(videoOrder1 != null ? videoOrder1 : 1);
                videoItem.setDurationSeconds(videoDuration1);
                videos.add(videoItem);
            }
            if (video2 != null && !video2.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video2);
                videoItem.setTitle(videoTitle2);
                videoItem.setVideoOrder(videoOrder2 != null ? videoOrder2 : 2);
                videoItem.setDurationSeconds(videoDuration2);
                videos.add(videoItem);
            }
            if (video3 != null && !video3.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video3);
                videoItem.setTitle(videoTitle3);
                videoItem.setVideoOrder(videoOrder3 != null ? videoOrder3 : 3);
                videoItem.setDurationSeconds(videoDuration3);
                videos.add(videoItem);
            }
            if (video4 != null && !video4.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video4);
                videoItem.setTitle(videoTitle4);
                videoItem.setVideoOrder(videoOrder4 != null ? videoOrder4 : 4);
                videoItem.setDurationSeconds(videoDuration4);
                videos.add(videoItem);
            }
            dto.setVideos(videos);

            ManuscriptVO manuscriptVO = manuscriptService.uploadManuscript(dto, userId);
            return Result.success("上传成功", manuscriptVO);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取稿件详情", description = "根据稿件ID获取稿件详情，包含视频列表")
    public Result<ManuscriptVO> getManuscriptById(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer currentUserId = JwtUtils.getUserIdFromRequest(request);
            ManuscriptVO manuscriptVO = manuscriptService.getManuscriptWithVideos(id, currentUserId);
            if (manuscriptVO == null) {
                return Result.error("稿件不存在");
            }
            return Result.success("获取成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}")
    @Operation(summary = "获取用户稿件列表", description = "根据用户ID获取稿件列表，支持状态筛选")
    public Result<List<ManuscriptVO>> getManuscriptsByUserId(
            @PathVariable Integer userId,
            @Parameter(description = "状态筛选") @RequestParam(required = false) String status) {
        try {
            Integer statusCode = convertStatusParam(status);
            List<ManuscriptVO> manuscripts = manuscriptService.getManuscriptsByUserIdWithPaging(userId, statusCode, 1, 100).stream().limit(100).collect(Collectors.toList());
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}/list")
    @Operation(summary = "获取用户稿件列表（分页）", description = "根据用户ID获取稿件列表，支持分页和状态筛选")
    public Result<Map<String, Object>> getManuscriptsByUserIdWithPaging(
            @PathVariable Integer userId,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size,
            @Parameter(description = "状态筛选") @RequestParam(required = false) String status) {
        try {
            Integer statusCode = convertStatusParam(status);
            List<ManuscriptVO> manuscripts = manuscriptService.getManuscriptsByUserIdWithPaging(userId, statusCode, page, size);
            Integer total = manuscriptService.countManuscriptsByUserIdAndStatus(userId, statusCode);

            Map<String, Object> result = new HashMap<>();
            result.put("list", manuscripts);
            result.put("total", total);
            result.put("page", page);
            result.put("size", size);
            result.put("totalPages", (int) Math.ceil((double) total / size));

            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}/stats")
    @Operation(summary = "获取用户稿件统计", description = "获取用户各状态稿件的数量统计")
    public Result<Map<String, Integer>> getManuscriptStatsByUserId(@PathVariable Integer userId) {
        try {
            Map<String, Integer> stats = manuscriptService.getManuscriptStatsByUserId(userId);
            return Result.success("获取成功", stats);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/me/list")
    @Operation(summary = "获取当前用户稿件列表（分页）", description = "获取当前登录用户的稿件列表，支持分页和状态筛选")
    public Result<Map<String, Object>> getMyManuscriptsWithPaging(
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size,
            @Parameter(description = "状态筛选") @RequestParam(required = false) String status,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            System.out.println("=== 稿件列表调试 ===");
            System.out.println("X-User-Id header: " + request.getHeader("X-User-Id"));
            System.out.println("Authorization header: " + request.getHeader("Authorization"));
            System.out.println("解析出的用户ID: " + userId);
            
            if (userId == null) {
                return Result.error("用户未登录");
            }
            Integer statusCode = convertStatusParam(status);
            System.out.println("状态参数: " + status + ", 转换后: " + statusCode);
            
            List<ManuscriptVO> manuscripts = manuscriptService.getManuscriptsByUserIdWithPaging(userId, statusCode, page, size);
            Integer total = manuscriptService.countManuscriptsByUserIdAndStatus(userId, statusCode);
            
            System.out.println("查询到的稿件数量: " + manuscripts.size());
            System.out.println("总数量: " + total);

            Map<String, Object> result = new HashMap<>();
            result.put("list", manuscripts);
            result.put("total", total);
            result.put("page", page);
            result.put("size", size);
            result.put("totalPages", (int) Math.ceil((double) total / size));

            return Result.success("获取成功", result);
        } catch (Exception e) {
            System.out.println("获取稿件列表异常: " + e.getMessage());
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/me/stats")
    @Operation(summary = "获取当前用户稿件统计", description = "获取当前登录用户各状态稿件的数量统计")
    public Result<Map<String, Integer>> getMyManuscriptStats(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            System.out.println("=== 稿件统计调试 ===");
            System.out.println("解析出的用户ID: " + userId);
            
            if (userId == null) {
                return Result.error("用户未登录");
            }
            Map<String, Integer> stats = manuscriptService.getManuscriptStatsByUserId(userId);
            System.out.println("统计数据: " + stats);
            return Result.success("获取成功", stats);
        } catch (Exception e) {
            System.out.println("获取稿件统计异常: " + e.getMessage());
            e.printStackTrace();
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新稿件", description = "更新稿件信息")
    public Result<ManuscriptVO> updateManuscript(
            @PathVariable Integer id,
            @RequestBody Manuscript manuscript,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            ManuscriptVO manuscriptVO = manuscriptService.updateManuscript(id, manuscript);
            return Result.success("更新成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除稿件", description = "删除自己的稿件")
    public Result<?> deleteManuscript(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            manuscriptService.deleteManuscript(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/recommended")
    @Operation(summary = "获取推荐稿件列表", description = "获取已上架的推荐稿件列表")
    public Result<List<ManuscriptVO>> getRecommendedManuscripts(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            List<ManuscriptVO> manuscripts = manuscriptService.getRecommendedManuscripts(userId);
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/hot")
    @Operation(summary = "获取热门稿件列表", description = "获取已上架的热门稿件列表，按播放量排序")
    public Result<List<ManuscriptVO>> getHotManuscripts(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            List<ManuscriptVO> manuscripts = manuscriptService.getHotManuscripts(userId);
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "获取分类稿件列表", description = "根据分类ID获取稿件列表")
    public Result<Map<String, Object>> getManuscriptsByCategoryId(
            @PathVariable Integer categoryId,
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "20") Integer size) {
        try {
            Map<String, Object> result = manuscriptService.getManuscriptsByCategoryId(categoryId, page, size);
            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/fix-durations")
    @Operation(summary = "修复所有稿件时长", description = "计算并更新所有稿件的总时长")
    public Result<?> fixAllManuscriptDurations() {
        try {
            int fixedCount = manuscriptService.fixAllManuscriptDurations();
            return Result.success("修复成功，共修复 " + fixedCount + " 个稿件的时长", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    private Integer convertStatusParam(String status) {
        if (status == null || status.isEmpty()) {
            return null;
        }
        // 如果是数字字符串，直接返回数字
        try {
            return Integer.parseInt(status);
        } catch (NumberFormatException e) {
            // 不是数字，按字符串处理
        }
        switch (status.toLowerCase()) {
            case "processing": return Manuscript.STATUS_PENDING_REVIEW;  // 进行中 -> 待审核(0)
            case "published": return Manuscript.STATUS_PUBLISHED;  // 已通过 -> 已发布(3)
            case "rejected": return Manuscript.STATUS_REJECTED;  // 未通过 -> 已拒绝(4)
            case "ready": return Manuscript.STATUS_READY_TO_PUBLISH;  // 待发布 -> 待发布(2)
            default: return null;
        }
    }
}