package com.mybilibili.web.controller;

import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.SubtitleService;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/subtitle")
@Tag(name = "字幕管理")
public class SubtitleController {

    @Autowired
    private SubtitleService subtitleService;

    @GetMapping("/video/{videoId}")
    public Result<List<Subtitle>> getSubtitles(@PathVariable Integer videoId) {
        List<Subtitle> subtitles = subtitleService.getSubtitlesByVideoId(videoId);
        return Result.success(subtitles);
    }

    @GetMapping("/video/{videoId}/{language}")
    public Result<Subtitle> getSubtitle(@PathVariable Integer videoId,
                                        @PathVariable String language) {
        Subtitle subtitle = subtitleService.getSubtitleByVideoIdAndLanguage(videoId, language);
        return Result.success(subtitle);
    }

    @GetMapping("/video/{videoId}/{language}/srt")
    public String getSubtitleSrt(@PathVariable Integer videoId,
                                 @PathVariable String language) {
        Subtitle subtitle = subtitleService.getSubtitleByVideoIdAndLanguage(videoId, language);
        if (subtitle == null || subtitle.getContent() == null) {
            return "";
        }
        // 将字幕内容转换为 SRT 格式
        return convertToSrt(subtitle.getContent());
    }

    private String convertToSrt(List<Subtitle.SubtitleItem> items) {
        StringBuilder srt = new StringBuilder();
        for (int i = 0; i < items.size(); i++) {
            Subtitle.SubtitleItem item = items.get(i);
            srt.append(i + 1).append("\n");
            srt.append(formatTime(item.getStartTime())).append(" --> ").append(formatTime(item.getEndTime())).append("\n");
            srt.append(item.getText()).append("\n\n");
        }
        return srt.toString();
    }

    private String formatTime(double seconds) {
        int hours = (int) (seconds / 3600);
        int minutes = (int) ((seconds % 3600) / 60);
        int secs = (int) (seconds % 60);
        int millis = (int) ((seconds - (int) seconds) * 1000);
        return String.format("%02d:%02d:%02d,%03d", hours, minutes, secs, millis);
    }

    @PostMapping("/upload")
    public Result<Subtitle> uploadSubtitle(@RequestBody Subtitle subtitle) {
        Subtitle savedSubtitle = subtitleService.uploadSubtitle(subtitle);
        return Result.success("上传成功", savedSubtitle);
    }

    @PostMapping("/upload-srt")
    public Result<Subtitle> uploadSrt(@RequestBody Map<String, Object> params) {
        Integer videoId = (Integer) params.get("videoId");
        String srtContent = (String) params.get("srtContent");
        String language = (String) params.get("language");
        String languageName = (String) params.get("languageName");
        Integer uploadedBy = (Integer) params.get("uploadedBy");

        if (videoId == null || srtContent == null || language == null) {
            return Result.error("参数不完整");
        }

        Subtitle subtitle = subtitleService.parseAndSaveSrt(videoId, srtContent, language, languageName, uploadedBy);
        return Result.success("上传成功", subtitle);
    }

    @DeleteMapping("/{subtitleId}")
    public Result<Void> deleteSubtitle(@PathVariable String subtitleId) {
        subtitleService.deleteSubtitle(subtitleId);
        return Result.success("删除成功", null);
    }

    @PostMapping("/set-default")
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
