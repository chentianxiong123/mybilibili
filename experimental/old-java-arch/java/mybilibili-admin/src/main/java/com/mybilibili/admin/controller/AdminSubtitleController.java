package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.AdminSubtitleService;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 管理员字幕管理控制器
 */
@RestController
@RequestMapping("/subtitle")
@Tag(name = "管理员字幕管理", description = "字幕管理、审核、入库等功能")
public class AdminSubtitleController {

    @Autowired
    private AdminSubtitleService adminSubtitleService;

    /**
     * 获取带字幕信息的视频列表
     */
    @Operation(summary = "获取视频列表（带字幕信息）", description = "分页查询视频，附带字幕统计信息")
    @GetMapping("/videos")
    public Result<Map<String, Object>> getVideosWithSubtitleInfo(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "20") Integer size) {
        return adminSubtitleService.getVideosWithSubtitleInfo(page, size);
    }

    /**
     * 获取指定视频的所有字幕
     */
    @Operation(summary = "获取视频字幕列表", description = "获取指定视频的所有字幕")
    @GetMapping("/video/{videoId}")
    public Result<List<Subtitle>> getSubtitlesByVideoId(@PathVariable Integer videoId) {
        List<Subtitle> subtitles = adminSubtitleService.getSubtitlesByVideoId(videoId);
        return Result.success(subtitles);
    }

    /**
     * 管理员上传字幕（直接入库，无需审核）
     */
    @Operation(summary = "管理员上传字幕", description = "管理员直接上传字幕，无需审核")
    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    public Result<Subtitle> uploadSubtitle(
            @RequestParam Integer videoId,
            @RequestParam MultipartFile file,
            @RequestParam String language,
            @RequestParam(required = false, defaultValue = "") String languageName,
            @RequestParam(required = false, defaultValue = "false") Boolean isDefault) {
        try {
            Subtitle subtitle = adminSubtitleService.uploadSubtitle(videoId, file, language, languageName, null);
            // 如果需要设为默认
            if (isDefault != null && isDefault && subtitle.getId() != null) {
                adminSubtitleService.setDefaultSubtitle(subtitle.getId());
            }
            return Result.success("字幕上传成功", subtitle);
        } catch (Exception e) {
            return Result.error("字幕上传失败: " + e.getMessage());
        }
    }

    /**
     * SRT文件入库（将磁盘上的SRT文件导入MongoDB）
     */
    @Operation(summary = "SRT文件入库", description = "将磁盘上的SRT文件导入MongoDB数据库")
    @PostMapping("/import-srt")
    public Result<Subtitle> importSrtToMongo(@RequestBody Map<String, Object> params) {
        try {
            Integer videoId = (Integer) params.get("videoId");
            String filePath = (String) params.get("srtFilePath");
            if (filePath == null) {
                filePath = (String) params.get("filePath");
            }
            String language = (String) params.get("language");
            String languageName = (String) params.get("languageName");
            Boolean isDefault = (Boolean) params.get("isDefault");

            if (videoId == null || filePath == null || language == null) {
                return Result.error("参数不完整");
            }

            Subtitle subtitle = adminSubtitleService.importSrtToMongo(
                videoId, filePath, language, languageName, "admin"
            );
            // 如果需要设为默认
            if (isDefault != null && isDefault && subtitle.getId() != null) {
                adminSubtitleService.setDefaultSubtitle(subtitle.getId());
            }
            return Result.success("字幕入库成功", subtitle);
        } catch (Exception e) {
            return Result.error("字幕入库失败: " + e.getMessage());
        }
    }

    /**
     * 审核通过字幕
     */
    @Operation(summary = "审核通过字幕", description = "将待审核字幕状态改为已通过，并设为默认")
    @PostMapping("/{subtitleId}/approve")
    public Result<Void> approveSubtitle(@PathVariable String subtitleId) {
        try {
            // 从当前登录用户获取reviewerId，这里暂时传null由service处理
            adminSubtitleService.approveSubtitle(subtitleId, null);
            return Result.success("审核通过", null);
        } catch (Exception e) {
            return Result.error("审核失败: " + e.getMessage());
        }
    }

    /**
     * 审核拒绝字幕
     */
    @Operation(summary = "审核拒绝字幕", description = "将待审核字幕状态改为已拒绝")
    @PostMapping("/{subtitleId}/reject")
    public Result<Void> rejectSubtitle(
            @PathVariable String subtitleId,
            @RequestBody Map<String, String> params) {
        try {
            String reason = params.get("reason");
            // 从当前登录用户获取reviewerId，这里暂时传null由service处理
            adminSubtitleService.rejectSubtitle(subtitleId, null, reason);
            return Result.success("已拒绝", null);
        } catch (Exception e) {
            return Result.error("操作失败: " + e.getMessage());
        }
    }

    /**
     * 设为默认字幕
     */
    @Operation(summary = "设为默认字幕", description = "将指定字幕设为默认，其他字幕设为非默认")
    @PostMapping("/{subtitleId}/set-default")
    public Result<Void> setDefaultSubtitle(@PathVariable String subtitleId) {
        try {
            adminSubtitleService.setDefaultSubtitle(subtitleId);
            return Result.success("设置成功", null);
        } catch (Exception e) {
            return Result.error("设置失败: " + e.getMessage());
        }
    }

    /**
     * 删除字幕
     */
    @Operation(summary = "删除字幕", description = "从MongoDB删除指定字幕")
    @DeleteMapping("/{subtitleId}")
    public Result<Void> deleteSubtitle(@PathVariable String subtitleId) {
        try {
            adminSubtitleService.deleteSubtitle(subtitleId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error("删除失败: " + e.getMessage());
        }
    }

    /**
     * 获取待审核字幕列表
     */
    @Operation(summary = "获取待审核字幕", description = "获取所有状态为待审核的字幕列表")
    @GetMapping("/pending")
    public Result<List<Map<String, Object>>> getPendingSubtitles() {
        List<Map<String, Object>> subtitles = adminSubtitleService.getPendingSubtitles();
        return Result.success(subtitles);
    }

    /**
     * 预览字幕内容
     */
    @Operation(summary = "预览字幕", description = "获取字幕内容预览（前10条）")
    @GetMapping("/{subtitleId}/preview")
    public Result<Map<String, Object>> previewSubtitle(@PathVariable String subtitleId) {
        try {
            Map<String, Object> preview = adminSubtitleService.previewSubtitle(subtitleId);
            return Result.success(preview);
        } catch (Exception e) {
            return Result.error("预览失败: " + e.getMessage());
        }
    }

    /**
     * 扫描系统字幕文件
     */
    @Operation(summary = "扫描系统字幕", description = "扫描视频的系统字幕目录，返回待入库的字幕文件列表")
    @GetMapping("/scan/{videoId}")
    public Result<List<Map<String, Object>>> scanSystemSubtitles(@PathVariable Integer videoId) {
        try {
            List<Map<String, Object>> subtitles = adminSubtitleService.scanSystemSubtitles(videoId);
            return Result.success(subtitles);
        } catch (Exception e) {
            return Result.error("扫描失败: " + e.getMessage());
        }
    }

    /**
     * 系统字幕入库
     */
    @Operation(summary = "系统字幕入库", description = "将系统生成的字幕文件解析并入库到MongoDB")
    @PostMapping("/import-system")
    public Result<Subtitle> importSystemSubtitle(@RequestBody Map<String, Object> params) {
        try {
            Integer videoId = (Integer) params.get("videoId");
            String language = (String) params.get("language");

            if (videoId == null || language == null) {
                return Result.error("参数不完整: 需要videoId和language");
            }

            Subtitle subtitle = adminSubtitleService.importSystemSubtitle(videoId, language);
            return Result.success("字幕入库成功", subtitle);
        } catch (Exception e) {
            return Result.error("字幕入库失败: " + e.getMessage());
        }
    }
}
