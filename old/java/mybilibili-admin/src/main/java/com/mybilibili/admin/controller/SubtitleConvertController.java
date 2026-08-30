package com.mybilibili.admin.controller;

import com.mybilibili.admin.dto.SrtParseRequest;
import com.mybilibili.admin.dto.SrtSaveRequest;
import com.mybilibili.admin.service.SubtitleConvertService;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

/**
 * 字幕转换控制器
 * 提供 SRT 字幕转 MongoDB 的独立测试接口
 */
@RestController
@RequestMapping("/subtitle-convert")
@Tag(name = "字幕转换", description = "SRT字幕转MongoDB独立接口")
public class SubtitleConvertController {

    @Autowired
    private SubtitleConvertService subtitleConvertService;

    /**
     * 上传 SRT 文件并保存到 MongoDB
     */
    @Operation(summary = "上传SRT文件到MongoDB", description = "上传SRT文件，解析并保存到MongoDB")
    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    public Result<Map<String, Object>> uploadSrtFile(
            @RequestParam("videoId") Integer videoId,
            @RequestParam("file") MultipartFile file,
            @RequestParam(value = "language", defaultValue = "zh-CN") String language,
            @RequestParam(value = "languageName", defaultValue = "中文") String languageName,
            @RequestParam(value = "source", defaultValue = "admin") String source,
            @RequestParam(value = "uploadedBy", defaultValue = "0") Integer uploadedBy) {
        try {
            // 参数校验
            if (videoId == null) {
                return Result.error("视频ID不能为空");
            }
            if (file.isEmpty()) {
                return Result.error("SRT文件不能为空");
            }

            // 读取文件内容
            String srtContent;
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(file.getInputStream(), StandardCharsets.UTF_8))) {
                srtContent = reader.lines().collect(Collectors.joining("\n"));
            }

            System.out.println("[字幕转换接口] 上传SRT文件: " + file.getOriginalFilename() +
                ", videoId=" + videoId + ", 大小=" + srtContent.length());

            // 保存到 MongoDB
            Subtitle subtitle = subtitleConvertService.saveSrtToMongo(
                videoId, srtContent, language, languageName, source, uploadedBy
            );

            // 构造返回结果
            Map<String, Object> result = new HashMap<>();
            result.put("id", subtitle.getId());
            result.put("videoId", subtitle.getVideoId());
            result.put("language", subtitle.getLanguage());
            result.put("languageName", subtitle.getLanguageName());
            result.put("itemCount", subtitle.getContent() != null ? subtitle.getContent().size() : 0);
            result.put("status", subtitle.getStatus());
            result.put("source", subtitle.getSource());

            return Result.success("字幕保存成功", result);

        } catch (Exception e) {
            System.err.println("[字幕转换接口] 保存失败: " + e.getMessage());
            return Result.error("保存失败: " + e.getMessage());
        }
    }

    /**
     * 将 SRT 文本内容保存到 MongoDB
     */
    @Operation(summary = "SRT文本转MongoDB", description = "传入SRT文本内容，解析并保存到MongoDB")
    @PostMapping("/save-content")
    public Result<Map<String, Object>> saveSrtContent(@RequestBody SrtSaveRequest request) {
        try {
            // 参数校验
            if (request.getVideoId() == null) {
                return Result.error("视频ID不能为空");
            }
            if (request.getSrtContent() == null || request.getSrtContent().trim().isEmpty()) {
                return Result.error("SRT内容不能为空");
            }

            // 设置默认值
            String language = request.getLanguage() != null ? request.getLanguage() : "zh-CN";
            String languageName = request.getLanguageName() != null ? request.getLanguageName() : "中文";
            String source = request.getSource() != null ? request.getSource() : "admin";
            Integer uploadedBy = request.getUploadedBy() != null ? request.getUploadedBy() : 0;

            System.out.println("[字幕转换接口] 接收到请求: videoId=" + request.getVideoId() + ", language=" + language);

            // 保存到 MongoDB
            Subtitle subtitle = subtitleConvertService.saveSrtToMongo(
                request.getVideoId(), request.getSrtContent(), language, languageName, source, uploadedBy
            );

            // 构造返回结果
            Map<String, Object> result = new HashMap<>();
            result.put("id", subtitle.getId());
            result.put("videoId", subtitle.getVideoId());
            result.put("language", subtitle.getLanguage());
            result.put("languageName", subtitle.getLanguageName());
            result.put("itemCount", subtitle.getContent() != null ? subtitle.getContent().size() : 0);
            result.put("status", subtitle.getStatus());
            result.put("source", subtitle.getSource());

            return Result.success("字幕保存成功", result);

        } catch (Exception e) {
            System.err.println("[字幕转换接口] 保存失败: " + e.getMessage());
            return Result.error("保存失败: " + e.getMessage());
        }
    }

    /**
     * 从文件路径读取 SRT 并保存到 MongoDB
     */
    @Operation(summary = "SRT文件转MongoDB", description = "从指定文件路径读取SRT文件，解析并保存到MongoDB")
    @PostMapping("/save-file")
    public Result<Map<String, Object>> saveSrtFile(@RequestBody Map<String, Object> params) {
        try {
            Integer videoId = (Integer) params.get("videoId");
            String filePath = (String) params.get("filePath");
            String language = (String) params.get("language");
            String languageName = (String) params.get("languageName");
            String source = (String) params.get("source");
            Integer uploadedBy = (Integer) params.get("uploadedBy");

            // 参数校验
            if (videoId == null) {
                return Result.error("视频ID不能为空");
            }
            if (filePath == null || filePath.trim().isEmpty()) {
                return Result.error("文件路径不能为空");
            }
            if (language == null) {
                language = "zh-CN";
            }
            if (languageName == null) {
                languageName = "中文";
            }
            if (source == null) {
                source = "admin";
            }
            if (uploadedBy == null) {
                uploadedBy = 0;
            }

            System.out.println("[字幕转换接口] 从文件读取: videoId=" + videoId + ", path=" + filePath);

            // 保存到 MongoDB
            Subtitle subtitle = subtitleConvertService.saveSrtFileToMongo(
                videoId, filePath, language, languageName, source, uploadedBy
            );

            // 构造返回结果
            Map<String, Object> result = new HashMap<>();
            result.put("id", subtitle.getId());
            result.put("videoId", subtitle.getVideoId());
            result.put("language", subtitle.getLanguage());
            result.put("languageName", subtitle.getLanguageName());
            result.put("itemCount", subtitle.getContent() != null ? subtitle.getContent().size() : 0);
            result.put("status", subtitle.getStatus());
            result.put("source", subtitle.getSource());

            return Result.success("字幕保存成功", result);

        } catch (Exception e) {
            System.err.println("[字幕转换接口] 保存失败: " + e.getMessage());
            return Result.error("保存失败: " + e.getMessage());
        }
    }

    /**
     * 上传 SRT 文件并解析，不保存到数据库（测试解析逻辑）
     */
    @Operation(summary = "上传SRT文件解析", description = "上传SRT文件，解析返回结果，不保存到数据库")
    @PostMapping(value = "/parse", consumes = "multipart/form-data")
    public Result<Map<String, Object>> parseSrtFile(@RequestParam("file") MultipartFile file) {
        try {
            if (file.isEmpty()) {
                return Result.error("SRT文件不能为空");
            }

            // 读取文件内容
            String srtContent;
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(file.getInputStream(), StandardCharsets.UTF_8))) {
                srtContent = reader.lines().collect(Collectors.joining("\n"));
            }

            System.out.println("[字幕转换接口] 解析SRT文件: " + file.getOriginalFilename() + ", 大小: " + srtContent.length());

            // 解析内容
            List<Subtitle.SubtitleItem> items = subtitleConvertService.parseSrtContent(srtContent);

            // 构造返回结果
            Map<String, Object> result = new HashMap<>();
            result.put("fileName", file.getOriginalFilename());
            result.put("itemCount", items.size());
            result.put("items", items.subList(0, Math.min(10, items.size()))); // 只返回前10条
            result.put("hasMore", items.size() > 10);

            return Result.success("解析成功", result);

        } catch (Exception e) {
            System.err.println("[字幕转换接口] 解析失败: " + e.getMessage());
            return Result.error("解析失败: " + e.getMessage());
        }
    }

    /**
     * 解析 SRT 文本内容，不保存到数据库（测试解析逻辑）
     */
    @Operation(summary = "解析SRT文本", description = "传入SRT文本内容，解析返回结果，不保存到数据库")
    @PostMapping("/parse-text")
    public Result<Map<String, Object>> parseSrtContent(@RequestBody SrtParseRequest request) {
        try {
            if (request.getSrtContent() == null || request.getSrtContent().trim().isEmpty()) {
                return Result.error("SRT内容不能为空");
            }

            System.out.println("[字幕转换接口] 解析SRT内容，长度: " + request.getSrtContent().length());

            // 解析内容
            List<Subtitle.SubtitleItem> items = subtitleConvertService.parseSrtContent(request.getSrtContent());

            // 构造返回结果
            Map<String, Object> result = new HashMap<>();
            result.put("itemCount", items.size());
            result.put("items", items.subList(0, Math.min(10, items.size()))); // 只返回前10条
            result.put("hasMore", items.size() > 10);

            return Result.success("解析成功", result);

        } catch (Exception e) {
            System.err.println("[字幕转换接口] 解析失败: " + e.getMessage());
            return Result.error("解析失败: " + e.getMessage());
        }
    }

    /**
     * 查询字幕是否存在
     */
    @Operation(summary = "查询字幕是否存在", description = "查询指定视频和语言的字幕是否已存在")
    @GetMapping("/exists")
    public Result<Map<String, Object>> checkSubtitleExists(
            @RequestParam Integer videoId,
            @RequestParam String language) {
        try {
            boolean exists = subtitleConvertService.existsSubtitle(videoId, language);

            Map<String, Object> result = new HashMap<>();
            result.put("videoId", videoId);
            result.put("language", language);
            result.put("exists", exists);

            return Result.success("查询成功", result);

        } catch (Exception e) {
            return Result.error("查询失败: " + e.getMessage());
        }
    }

    /**
     * 删除字幕
     */
    @Operation(summary = "删除字幕", description = "从MongoDB删除指定视频和语言的字幕")
    @PostMapping("/delete")
    public Result<Void> deleteSubtitle(@RequestBody Map<String, Object> params) {
        try {
            Integer videoId = (Integer) params.get("videoId");
            String language = (String) params.get("language");

            if (videoId == null) {
                return Result.error("视频ID不能为空");
            }
            if (language == null) {
                language = "zh-CN";
            }

            subtitleConvertService.deleteSubtitleFromMongo(videoId, language);
            return Result.success("删除成功", null);

        } catch (Exception e) {
            return Result.error("删除失败: " + e.getMessage());
        }
    }
}
