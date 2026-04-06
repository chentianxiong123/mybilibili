package com.mybilibili.web.controller;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.ManuscriptService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.servlet.http.HttpServletRequest;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/manuscript")
@Tag(name = "稿件接口", description = "稿件上传、查询、更新、删除等操作")
public class ManuscriptController {

    @Autowired
    private ManuscriptService manuscriptService;

    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    @Operation(summary = "上传稿件", description = "上传新稿件，支持单视频和多视频分P")
    @SecurityRequirement(name = "JWT")
    public Result<ManuscriptVO> uploadManuscript(
            @RequestParam("title") String title,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam("cover") MultipartFile cover,
            @RequestParam("categoryId") Integer categoryId,
            @RequestParam(value = "tags", required = false) List<String> tags,
            @RequestParam(value = "videos[0].video", required = false) MultipartFile video0,
            @RequestParam(value = "videos[0].title", required = false) String videoTitle0,
            @RequestParam(value = "videos[0].videoOrder", required = false) Integer videoOrder0,
            @RequestParam(value = "videos[1].video", required = false) MultipartFile video1,
            @RequestParam(value = "videos[1].title", required = false) String videoTitle1,
            @RequestParam(value = "videos[1].videoOrder", required = false) Integer videoOrder1,
            @RequestParam(value = "videos[2].video", required = false) MultipartFile video2,
            @RequestParam(value = "videos[2].title", required = false) String videoTitle2,
            @RequestParam(value = "videos[2].videoOrder", required = false) Integer videoOrder2,
            @RequestParam(value = "videos[3].video", required = false) MultipartFile video3,
            @RequestParam(value = "videos[3].title", required = false) String videoTitle3,
            @RequestParam(value = "videos[3].videoOrder", required = false) Integer videoOrder3,
            @RequestParam(value = "videos[4].video", required = false) MultipartFile video4,
            @RequestParam(value = "videos[4].title", required = false) String videoTitle4,
            @RequestParam(value = "videos[4].videoOrder", required = false) Integer videoOrder4,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            // 构建DTO
            ManuscriptUploadDTO dto = new ManuscriptUploadDTO();
            dto.setTitle(title);
            dto.setDescription(description);
            dto.setCover(cover);
            dto.setCategoryId(categoryId);
            dto.setTags(tags);

            // 构建视频列表
            List<ManuscriptUploadDTO.VideoItemDTO> videos = new ArrayList<>();
            if (video0 != null && !video0.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video0);
                videoItem.setTitle(videoTitle0);
                videoItem.setVideoOrder(videoOrder0 != null ? videoOrder0 : 0);
                videos.add(videoItem);
            }
            if (video1 != null && !video1.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video1);
                videoItem.setTitle(videoTitle1);
                videoItem.setVideoOrder(videoOrder1 != null ? videoOrder1 : 1);
                videos.add(videoItem);
            }
            if (video2 != null && !video2.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video2);
                videoItem.setTitle(videoTitle2);
                videoItem.setVideoOrder(videoOrder2 != null ? videoOrder2 : 2);
                videos.add(videoItem);
            }
            if (video3 != null && !video3.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video3);
                videoItem.setTitle(videoTitle3);
                videoItem.setVideoOrder(videoOrder3 != null ? videoOrder3 : 3);
                videos.add(videoItem);
            }
            if (video4 != null && !video4.isEmpty()) {
                ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
                videoItem.setVideo(video4);
                videoItem.setTitle(videoTitle4);
                videoItem.setVideoOrder(videoOrder4 != null ? videoOrder4 : 4);
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
    public Result<ManuscriptVO> getManuscriptById(@PathVariable Integer id, HttpServletRequest request) {
        try {
            // 使用getManuscriptWithVideos获取包含视频列表的稿件详情
            ManuscriptVO manuscriptVO = manuscriptService.getManuscriptWithVideos(id);
            if (manuscriptVO == null) {
                return Result.error("稿件不存在");
            }
            return Result.success("获取成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}")
    @Operation(summary = "获取用户稿件列表", description = "根据用户ID获取稿件列表")
    public Result<List<ManuscriptVO>> getManuscriptsByUserId(@PathVariable Integer userId) {
        try {
            List<ManuscriptVO> manuscripts = manuscriptService.getManuscriptsByUserId(userId);
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}/list")
    @Operation(summary = "获取用户稿件列表（分页）", description = "根据用户ID获取稿件列表，支持分页和状态筛选")
    public Result<Map<String, Object>> getManuscriptsByUserIdWithPaging(
            @PathVariable Integer userId,
            @Parameter(description = "页码，从1开始") @RequestParam(value = "page", defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(value = "size", defaultValue = "10") Integer size,
            @Parameter(description = "状态筛选：draft(草稿/待审核), processing(处理中), approved(已通过/待上架), rejected(审核拒绝)") 
            @RequestParam(value = "status", required = false) String status) {
        try {
            // 转换状态参数
            Integer statusCode = convertStatusParam(status);
            
            // 分页查询稿件列表
            List<ManuscriptVO> manuscripts = manuscriptService.getManuscriptsByUserIdWithPaging(userId, statusCode, page, size);
            
            // 查询总数
            Integer total = manuscriptService.countManuscriptsByUserIdAndStatus(userId, statusCode);
            
            // 构建返回结果
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

    /**
     * 转换状态参数
     * @param status 状态字符串
     * @return 状态码
     */
    private Integer convertStatusParam(String status) {
        if (status == null || status.isEmpty()) {
            return null;
        }
        switch (status.toLowerCase()) {
            case "draft":
                return Manuscript.STATUS_PENDING_REVIEW;
            case "processing":
                return Manuscript.STATUS_PROCESSING;
            case "approved":
                return Manuscript.STATUS_READY_TO_PUBLISH;
            case "rejected":
                return Manuscript.STATUS_REJECTED;
            default:
                return null;
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新稿件", description = "更新稿件信息")
    @SecurityRequirement(name = "JWT")
    public Result<ManuscriptVO> updateManuscript(
            @PathVariable Integer id,
            @RequestBody Manuscript manuscript,
            HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            ManuscriptVO manuscriptVO = manuscriptService.updateManuscript(id, manuscript);
            return Result.success("更新成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除稿件", description = "删除自己的稿件")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteManuscript(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            manuscriptService.deleteManuscript(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/recommended")
    @Operation(summary = "获取推荐稿件列表", description = "获取已上架的推荐稿件列表，用于首页展示。支持个性化推荐（已登录用户）")
    public Result<List<ManuscriptVO>> getRecommendedManuscripts(HttpServletRequest request) {
        try {
            // 尝试从请求中获取用户ID（如果用户已登录）
            Integer userId = null;
            try {
                String authHeader = request.getHeader("Authorization");
                if (authHeader != null && !authHeader.isEmpty()) {
                    userId = JwtUtils.getUserIdFromToken(authHeader);
                }
            } catch (Exception e) {
                // 用户未登录或token无效，使用匿名推荐
                log.debug("用户未登录或token无效，使用匿名推荐");
            }

            List<ManuscriptVO> manuscripts = manuscriptService.getRecommendedManuscripts(userId);
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            log.error("获取推荐稿件失败", e);
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "获取分类稿件列表", description = "根据分类ID获取已上架的稿件列表")
    public Result<List<ManuscriptVO>> getManuscriptsByCategoryId(@PathVariable Integer categoryId) {
        try {
            List<ManuscriptVO> manuscripts = manuscriptService.getManuscriptsByCategoryId(categoryId);
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            log.error("获取分类稿件失败", e);
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/recalculate-duration")
    @Operation(summary = "重新计算稿件时长", description = "重新计算并更新指定稿件的总时长，用于修复时长数据")
    @SecurityRequirement(name = "JWT")
    public Result<ManuscriptVO> recalculateDuration(@PathVariable Integer id, HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            // 检查权限（只有稿件所有者或管理员可以操作）
            ManuscriptVO existing = manuscriptService.getManuscriptById(id);
            if (existing == null) {
                return Result.error("稿件不存在");
            }
            if (!existing.getUserId().equals(userId)) {
                return Result.error("没有权限操作此稿件");
            }

            ManuscriptVO manuscriptVO = manuscriptService.recalculateDuration(id);
            return Result.success("时长已重新计算", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/fix-all-durations")
    @Operation(summary = "批量修复所有稿件时长", description = "修复系统中所有稿件的时长数据，需要管理员权限")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> fixAllDurations(HttpServletRequest request) {
        try {
            // 从JWT中获取用户ID
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));

            // TODO: 检查管理员权限（这里简化处理，实际需要检查用户角色）
            // 暂时只允许特定用户ID访问，或根据实际权限系统调整

            int fixedCount = manuscriptService.fixAllManuscriptDurations();

            Map<String, Object> result = new HashMap<>();
            result.put("fixedCount", fixedCount);
            result.put("message", "共修复 " + fixedCount + " 个稿件的时长");

            return Result.success("修复完成", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
