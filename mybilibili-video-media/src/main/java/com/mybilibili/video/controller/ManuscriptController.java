package com.mybilibili.video.controller;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.dto.ManuscriptUpdateDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.dto.CreateUploadSessionRequest;
import com.mybilibili.video.service.ManuscriptService;
import com.mybilibili.video.service.ManuscriptUploadSessionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;
import org.springframework.web.multipart.MultipartHttpServletRequest;

import jakarta.servlet.http.HttpServletRequest;
import java.util.Arrays;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/manuscript")
@Tag(name = "稿件接口", description = "稿件查询、管理接口")
public class ManuscriptController {

    private static final Pattern VIDEO_FILE_PATTERN = Pattern.compile("^videos\\[(\\d+)]\\.file$");

    @Autowired
    private ManuscriptService manuscriptService;

    @Autowired
    private ManuscriptUploadSessionService uploadSessionService;

    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    @Operation(summary = "上传稿件", description = "上传新稿件，支持单视频和多视频分P")
    public Result<ManuscriptVO> uploadManuscript(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            if (!(request instanceof MultipartHttpServletRequest multipartRequest)) {
                return Result.error("请求格式错误");
            }

            ManuscriptUploadDTO dto = new ManuscriptUploadDTO();
            dto.setTitle(multipartRequest.getParameter("title"));
            dto.setDescription(multipartRequest.getParameter("description"));
            dto.setCover(multipartRequest.getFile("cover"));
            dto.setCategoryId(parseInteger(multipartRequest.getParameter("categoryId")));
            dto.setTags(extractTags(multipartRequest));
            dto.setVideos(extractVideos(multipartRequest));

            ManuscriptVO manuscriptVO = manuscriptService.uploadManuscript(dto, userId);
            return Result.success("上传成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/upload-session")
    @Operation(summary = "创建分片上传会话", description = "为大文件分片上传创建会话")
    public Result<?> createUploadSession(@RequestBody CreateUploadSessionRequest body, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId == null) {
            return Result.error("用户未登录");
        }
        try {
            return Result.success("创建成功", uploadSessionService.createSession(userId, body));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/upload-session/{uploadId}")
    @Operation(summary = "获取分片上传会话状态", description = "查询已上传分片和待上传进度")
    public Result<?> getUploadSessionStatus(@PathVariable String uploadId, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId == null) {
            return Result.error("用户未登录");
        }
        try {
            return Result.success("获取成功", uploadSessionService.getSessionStatus(userId, uploadId));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping(value = "/upload-chunk", consumes = "multipart/form-data")
    @Operation(summary = "上传视频分片", description = "上传单个视频分片")
    public Result<?> uploadChunk(@RequestParam String uploadId,
                                 @RequestParam int partIndex,
                                 @RequestParam int chunkIndex,
                                 @RequestParam int totalChunks,
                                 @RequestParam("file") MultipartFile file,
                                 HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId == null) {
            return Result.error("用户未登录");
        }
        try {
            return Result.success("上传成功", uploadSessionService.uploadChunk(userId, uploadId, partIndex, chunkIndex, totalChunks, file));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping(value = "/upload-complete", consumes = "multipart/form-data")
    @Operation(summary = "完成分片上传", description = "合并分片并提交稿件")
    public Result<ManuscriptVO> completeUploadSession(@RequestParam String uploadId,
                                                      @RequestParam("cover") MultipartFile cover,
                                                      HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId == null) {
            return Result.error("用户未登录");
        }
        try {
            return Result.success("上传成功", uploadSessionService.completeSession(userId, uploadId, cover));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/upload-session/{uploadId}")
    @Operation(summary = "取消分片上传", description = "取消并清理分片上传会话")
    public Result<?> cancelUploadSession(@PathVariable String uploadId, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId == null) {
            return Result.error("用户未登录");
        }
        try {
            uploadSessionService.cancelSession(userId, uploadId);
            return Result.success("已取消");
        } catch (Exception e) {
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
            String viewerKey = currentUserId != null ? "u:" + currentUserId : "ip:" + getClientIp(request);
            manuscriptService.incrementViewCount(id, viewerKey);
            return Result.success("获取成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/internal/{id}")
    @Operation(summary = "内部获取稿件详情", description = "服务间调用，不增加播放量")
    public Result<ManuscriptVO> getManuscriptByIdInternal(@PathVariable Integer id) {
        try {
            ManuscriptVO manuscriptVO = manuscriptService.getManuscriptById(id);
            if (manuscriptVO == null) {
                return Result.error("稿件不存在");
            }
            return Result.success("获取成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/internal/{id}/take-down")
    @Operation(summary = "内部下架稿件", description = "服务间调用,下架违规稿件")
    public Result<?> takeDownManuscriptInternal(@PathVariable Integer id,
                                                 @RequestParam(required = false) String reason) {
        try {
            boolean success = manuscriptService.takeDownViolatingManuscript(id, reason);
            if (success) {
                return Result.success("下架成功");
            }
            return Result.error("稿件不存在");
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

    @GetMapping("/user/{userId}/search")
    @Operation(summary = "搜索用户稿件", description = "根据关键词搜索用户的稿件标题")
    public Result<List<ManuscriptVO>> searchUserManuscripts(
            @PathVariable Integer userId,
            @Parameter(description = "搜索关键词") @RequestParam String keyword,
            @Parameter(description = "排序方式") @RequestParam(required = false) String sort) {
        try {
            List<ManuscriptVO> results = manuscriptService.searchUserManuscripts(userId, keyword, sort);
            return Result.success("搜索成功", results);
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
            if (userId == null) {
                return Result.error("用户未登录");
            }
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

    @GetMapping("/me/stats")
    @Operation(summary = "获取当前用户稿件统计", description = "获取当前登录用户各状态稿件的数量统计")
    public Result<Map<String, Integer>> getMyManuscriptStats(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            Map<String, Integer> stats = manuscriptService.getManuscriptStatsByUserId(userId);
            return Result.success("获取成功", stats);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/me/{id}")
    @Operation(summary = "获取当前用户稿件详情", description = "获取当前登录用户自己的稿件详情，不增加播放量")
    public Result<ManuscriptVO> getMyManuscriptById(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            ManuscriptVO manuscriptVO = manuscriptService.getManuscriptWithVideos(id, userId);
            if (manuscriptVO == null) {
                return Result.error("稿件不存在");
            }
            if (!userId.equals(manuscriptVO.getUserId())) {
                return Result.error(403, "没有权限查看此稿件");
            }
            return Result.success("获取成功", manuscriptVO);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping(value = "/{id}", consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    @Operation(summary = "更新稿件", description = "更新稿件信息")
    public Result<ManuscriptVO> updateManuscript(
            @PathVariable Integer id,
            @Valid @ModelAttribute ManuscriptUpdateDTO dto,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            ManuscriptVO manuscriptVO = manuscriptService.updateManuscriptByOwner(id, userId, dto);
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

    @PostMapping("/{id}/unpublish")
    @Operation(summary = "下架稿件", description = "下架自己的稿件")
    public Result<?> unpublishManuscript(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            boolean success = manuscriptService.unpublishManuscriptByOwner(id, userId);
            if (success) {
                return Result.success("下架成功", null);
            } else {
                return Result.error("稿件不存在或无权操作");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/publish")
    @Operation(summary = "上架稿件", description = "上架自己的稿件")
    public Result<?> publishManuscript(@PathVariable Integer id, HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                return Result.error("用户未登录");
            }
            boolean success = manuscriptService.publishManuscriptByOwner(id, userId);
            if (success) {
                return Result.success("上架成功", null);
            } else {
                return Result.error("稿件不存在或无权操作");
            }
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

    @PutMapping("/{id}/comment-count")
    @Operation(summary = "更新稿件评论数", description = "内部接口：更新稿件的评论数量")
    public Result<?> updateCommentCount(
            @PathVariable Integer id,
            @RequestParam Integer count) {
        try {
            manuscriptService.updateCommentCount(id, count);
            return Result.success("更新成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/increment-comment")
    @Operation(summary = "增加评论数", description = "内部接口：增加稿件的评论数")
    public Result<?> incrementCommentCount(@PathVariable Integer id) {
        try {
            manuscriptService.incrementCommentCount(id);
            return Result.success("更新成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{id}/decrement-comment")
    @Operation(summary = "减少评论数", description = "内部接口：减少稿件的评论数")
    public Result<?> decrementCommentCount(@PathVariable Integer id) {
        try {
            manuscriptService.decrementCommentCount(id);
            return Result.success("更新成功");
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
            case "pending":
            case "reviewing":
            case "draft":
                return Manuscript.STATUS_PENDING_REVIEW;
            case "processing":
                return Manuscript.STATUS_PROCESSING;
            case "published":
                return Manuscript.STATUS_PUBLISHED;
            case "rejected":
                return Manuscript.STATUS_REJECTED;
            case "failed":
                return Manuscript.STATUS_PROCESS_FAILED;
            case "unpublished":
                return Manuscript.STATUS_UNPUBLISHED;
            default: return null;
        }
    }

    private List<String> extractTags(MultipartHttpServletRequest request) {
        String[] values = request.getParameterValues("tags");
        if (values == null || values.length == 0) {
            return List.of();
        }
        return Arrays.stream(values)
                .filter(StringUtils::hasText)
                .map(String::trim)
                .distinct()
                .collect(Collectors.toList());
    }

    private List<ManuscriptUploadDTO.VideoItemDTO> extractVideos(MultipartHttpServletRequest request) {
        return request.getFileMap().entrySet().stream()
                .map(entry -> toVideoItem(entry.getKey(), entry.getValue(), request))
                .filter(java.util.Objects::nonNull)
                .sorted(Comparator.comparingInt(ManuscriptUploadDTO.VideoItemDTO::getVideoOrder))
                .collect(Collectors.toList());
    }

    private ManuscriptUploadDTO.VideoItemDTO toVideoItem(String key, MultipartFile file, MultipartHttpServletRequest request) {
        Matcher matcher = VIDEO_FILE_PATTERN.matcher(key);
        if (!matcher.matches() || file == null || file.isEmpty()) {
            return null;
        }
        int index = Integer.parseInt(matcher.group(1));
        ManuscriptUploadDTO.VideoItemDTO item = new ManuscriptUploadDTO.VideoItemDTO();
        item.setVideo(file);
        item.setTitle(resolveVideoTitle(request.getParameter("videos[" + index + "].title"), file, index));
        Integer order = parseInteger(request.getParameter("videos[" + index + "].videoOrder"));
        item.setVideoOrder(order == null ? index : order);
        item.setDurationSeconds(parseInteger(request.getParameter("videos[" + index + "].durationSeconds")));
        return item;
    }

    private String resolveVideoTitle(String title, MultipartFile file, int index) {
        if (StringUtils.hasText(title)) {
            return title.trim();
        }
        String originalFilename = file.getOriginalFilename();
        if (StringUtils.hasText(originalFilename)) {
            String filename = originalFilename.replace("\\", "/");
            int slashIndex = filename.lastIndexOf('/');
            if (slashIndex >= 0) {
                filename = filename.substring(slashIndex + 1);
            }
            int dotIndex = filename.lastIndexOf('.');
            return dotIndex > 0 ? filename.substring(0, dotIndex) : filename;
        }
        return "P" + (index + 1);
    }

    private Integer parseInteger(String value) {
        if (!StringUtils.hasText(value)) {
            return null;
        }
        try {
            return Integer.parseInt(value.trim());
        } catch (NumberFormatException e) {
            return null;
        }
    }

    private String getClientIp(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip != null && !ip.isBlank()) {
            return ip.split(",")[0].trim();
        }
        ip = request.getHeader("X-Real-IP");
        if (ip != null && !ip.isBlank()) {
            return ip;
        }
        return request.getRemoteAddr();
    }
}
