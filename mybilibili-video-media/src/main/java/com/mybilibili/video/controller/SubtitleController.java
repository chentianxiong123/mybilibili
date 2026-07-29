package com.mybilibili.video.controller;

import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.SubtitleService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/subtitle")
@Tag(name = "字幕管理")
public class SubtitleController {

    @Autowired
    private SubtitleService subtitleService;

    @GetMapping("/videos")
    @Operation(summary = "获取视频字幕管理列表", description = "后台字幕管理页使用，返回视频和字幕统计")
    public Result<List<Map<String, Object>>> getVideosWithSubtitleInfo() {
        return Result.success(subtitleService.getVideosWithSubtitleInfo());
    }

    @GetMapping("/video/{videoId}")
    @Operation(summary = "获取视频字幕列表", description = "获取指定视频的所有字幕")
    public Result<List<Subtitle>> getSubtitles(@PathVariable Integer videoId) {
        List<Subtitle> subtitles = subtitleService.getSubtitlesByVideoId(videoId);
        return Result.success(subtitles);
    }

    @GetMapping("/video/{videoId}/{language}")
    @Operation(summary = "获取指定语言字幕", description = "获取指定视频的指定语言字幕")
    public Result<Subtitle> getSubtitle(@PathVariable Integer videoId,
                                        @PathVariable String language) {
        try {
            Subtitle subtitle = subtitleService.getSubtitleByVideoIdAndLanguage(videoId, language);
            return Result.success(subtitle);
        } catch (Exception e) {
            return Result.error("获取字幕失败: " + e.getMessage());
        }
    }

    @PostMapping(value = "/upload", consumes = MediaType.APPLICATION_JSON_VALUE)
    @Operation(summary = "上传字幕", description = "上传字幕文件")
    public Result<Subtitle> uploadSubtitle(@RequestBody Subtitle subtitle, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId != null) {
            subtitle.setUploadedBy(userId);
        }
        Subtitle savedSubtitle = subtitleService.uploadSubtitle(subtitle);
        return Result.success("上传成功", savedSubtitle);
    }

    @PostMapping(value = "/upload", consumes = MediaType.MULTIPART_FORM_DATA_VALUE)
    @Operation(summary = "上传SRT字幕文件", description = "后台上传SRT字幕文件并入库")
    public Result<Subtitle> uploadSubtitleFile(@RequestParam Integer videoId,
                                               @RequestParam MultipartFile file,
                                               @RequestParam String language,
                                               @RequestParam(defaultValue = "false") Boolean isDefault,
                                               HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            Subtitle subtitle = subtitleService.uploadSrtFile(videoId, file, language, isDefault, userId);
            return Result.success("上传成功", subtitle);
        } catch (Exception e) {
            return Result.error("上传失败: " + e.getMessage());
        }
    }

    @PostMapping("/upload-srt")
    @Operation(summary = "上传SRT字幕", description = "上传并解析SRT格式字幕")
    public Result<Subtitle> uploadSrt(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        Integer videoId = toInteger(params.get("videoId"));
        String srtContent = (String) params.get("srtContent");
        String language = (String) params.get("language");
        String languageName = (String) params.get("languageName");
        Integer uploadedBy = toInteger(params.get("uploadedBy"));

        if (videoId == null || srtContent == null || language == null) {
            return Result.error("参数不完整");
        }

        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId != null && uploadedBy == null) {
            uploadedBy = userId;
        }

        Subtitle subtitle = subtitleService.parseAndSaveSrt(videoId, srtContent, language, languageName, uploadedBy);
        return Result.success("上传成功", subtitle);
    }

    @PostMapping("/import-srt")
    @Operation(summary = "导入SRT字幕", description = "从本地路径或对象存储Key导入SRT字幕")
    public Result<Subtitle> importSrt(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        try {
            Integer videoId = toInteger(params.get("videoId"));
            String srtFilePath = (String) params.get("srtFilePath");
            String language = (String) params.get("language");
            Boolean isDefault = toBoolean(params.get("isDefault"));

            if (videoId == null || srtFilePath == null || language == null) {
                return Result.error("参数不完整");
            }

            Integer userId = JwtUtils.getUserIdFromRequest(request);
            Subtitle subtitle = subtitleService.importSrtFilePath(videoId, srtFilePath, language, isDefault, userId);
            return Result.success("导入成功", subtitle);
        } catch (Exception e) {
            return Result.error("导入失败: " + e.getMessage());
        }
    }

    @GetMapping("/pending")
    @Operation(summary = "获取待审核字幕列表", description = "获取所有待审核字幕")
    public Result<List<Subtitle>> getPendingSubtitles() {
        return Result.success(subtitleService.getPendingSubtitles());
    }

    @PostMapping("/{subtitleId}/approve")
    @Operation(summary = "审核通过字幕", description = "将待审核字幕标记为审核通过")
    public Result<Subtitle> approveSubtitle(@PathVariable String subtitleId, HttpServletRequest request) {
        try {
            Integer reviewerId = JwtUtils.getUserIdFromRequest(request);
            return Result.success("审核通过", subtitleService.approveSubtitle(subtitleId, reviewerId));
        } catch (Exception e) {
            return Result.error("审核失败: " + e.getMessage());
        }
    }

    @PostMapping("/{subtitleId}/reject")
    @Operation(summary = "审核拒绝字幕", description = "将待审核字幕标记为审核拒绝")
    public Result<Subtitle> rejectSubtitle(@PathVariable String subtitleId,
                                           @RequestBody(required = false) Map<String, Object> params,
                                           HttpServletRequest request) {
        try {
            String reason = params == null ? null : (String) params.get("reason");
            Integer reviewerId = JwtUtils.getUserIdFromRequest(request);
            return Result.success("已拒绝", subtitleService.rejectSubtitle(subtitleId, reason, reviewerId));
        } catch (Exception e) {
            return Result.error("审核失败: " + e.getMessage());
        }
    }

    @GetMapping("/{subtitleId}/preview")
    @Operation(summary = "预览字幕", description = "获取字幕内容用于后台预览")
    public Result<Map<String, Object>> previewSubtitle(@PathVariable String subtitleId) {
        try {
            return Result.success(subtitleService.previewSubtitle(subtitleId));
        } catch (Exception e) {
            return Result.error("获取字幕内容失败: " + e.getMessage());
        }
    }

    @GetMapping("/scan/{videoId}")
    @Operation(summary = "扫描系统字幕文件", description = "检查对象存储中是否存在尚未入库的系统字幕")
    public Result<List<Map<String, Object>>> scanSystemSubtitles(@PathVariable Integer videoId) {
        return Result.success(subtitleService.scanSystemSubtitles(videoId));
    }

    @PostMapping("/import-system")
    @Operation(summary = "系统字幕入库", description = "将AI生成并上传到对象存储的SRT字幕导入MongoDB")
    public Result<Subtitle> importSystemSubtitle(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        try {
            Integer videoId = toInteger(params.get("videoId"));
            String language = (String) params.get("language");
            Boolean isDefault = toBoolean(params.get("isDefault"));
            if (videoId == null || language == null) {
                return Result.error("参数不完整");
            }
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            Subtitle subtitle = subtitleService.importSystemSubtitle(videoId, language, isDefault, userId);
            return Result.success("入库成功", subtitle);
        } catch (Exception e) {
            return Result.error("入库失败: " + e.getMessage());
        }
    }

    @PostMapping("/{subtitleId}/set-default")
    @Operation(summary = "按字幕ID设置默认字幕", description = "后台字幕管理页使用")
    public Result<Void> setDefaultSubtitleById(@PathVariable String subtitleId) {
        try {
            subtitleService.setDefaultSubtitleById(subtitleId);
            return Result.success("设置成功", null);
        } catch (Exception e) {
            return Result.error("设置失败: " + e.getMessage());
        }
    }

    @DeleteMapping("/{subtitleId}")
    @Operation(summary = "删除字幕", description = "删除指定的字幕")
    public Result<Void> deleteSubtitle(@PathVariable String subtitleId, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId == null) {
            return Result.error("用户未登录");
        }
        subtitleService.deleteSubtitle(subtitleId);
        return Result.success("删除成功", null);
    }

    @PostMapping("/set-default")
    @Operation(summary = "设置默认字幕", description = "设置指定视频的默认字幕")
    public Result<Void> setDefaultSubtitle(@RequestBody Map<String, Object> params) {
        Integer videoId = toInteger(params.get("videoId"));
        String language = (String) params.get("language");

        if (videoId == null || language == null) {
            return Result.error("参数不完整");
        }

        subtitleService.setDefaultSubtitle(videoId, language);
        return Result.success("设置成功", null);
    }

    private Integer toInteger(Object value) {
        if (value == null) {
            return null;
        }
        if (value instanceof Number number) {
            return number.intValue();
        }
        if (value instanceof String text && !text.isBlank()) {
            return Integer.parseInt(text);
        }
        return null;
    }

    private Boolean toBoolean(Object value) {
        if (value == null) {
            return false;
        }
        if (value instanceof Boolean bool) {
            return bool;
        }
        if (value instanceof String text) {
            return Boolean.parseBoolean(text);
        }
        return false;
    }
}
