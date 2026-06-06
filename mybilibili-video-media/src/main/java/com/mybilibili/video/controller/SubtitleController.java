package com.mybilibili.video.controller;

import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.SubtitleService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/subtitle")
@Tag(name = "字幕管理")
public class SubtitleController {

    @Autowired
    private SubtitleService subtitleService;

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

    @PostMapping("/upload")
    @Operation(summary = "上传字幕", description = "上传字幕文件")
    public Result<Subtitle> uploadSubtitle(@RequestBody Subtitle subtitle, HttpServletRequest request) {
        Integer userId = JwtUtils.getUserIdFromRequest(request);
        if (userId != null) {
            subtitle.setUploadedBy(userId);
        }
        Subtitle savedSubtitle = subtitleService.uploadSubtitle(subtitle);
        return Result.success("上传成功", savedSubtitle);
    }

    @PostMapping("/upload-srt")
    @Operation(summary = "上传SRT字幕", description = "上传并解析SRT格式字幕")
    public Result<Subtitle> uploadSrt(@RequestBody Map<String, Object> params, HttpServletRequest request) {
        Integer videoId = (Integer) params.get("videoId");
        String srtContent = (String) params.get("srtContent");
        String language = (String) params.get("language");
        String languageName = (String) params.get("languageName");
        Integer uploadedBy = (Integer) params.get("uploadedBy");

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
        Integer videoId = (Integer) params.get("videoId");
        String language = (String) params.get("language");

        if (videoId == null || language == null) {
            return Result.error("参数不完整");
        }

        subtitleService.setDefaultSubtitle(videoId, language);
        return Result.success("设置成功", null);
    }
}
