package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.VideoMapper;
import com.mybilibili.admin.repository.SubtitleRepository;
import com.mybilibili.admin.service.AdminSubtitleService;
import com.mybilibili.admin.utils.UploadFilePathUtils;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.*;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

/**
 * 管理员字幕服务实现
 */
@Service
public class AdminSubtitleServiceImpl implements AdminSubtitleService {

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @Override
    public Result<Map<String, Object>> getVideosWithSubtitleInfo(Integer page, Integer size) {
        // 获取所有视频
        List<Video> videos = videoMapper.selectAll();

        // 获取字幕统计信息
        List<Map<String, Object>> videoList = videos.stream()
            .map(video -> {
                Map<String, Object> map = new HashMap<>();
                map.put("id", video.getId());
                map.put("title", video.getTitle());
                map.put("durationSeconds", video.getDurationSeconds());
                map.put("manuscriptId", video.getManuscriptId());

                // 查询该视频的字幕
                List<Subtitle> subtitles = subtitleRepository.findByVideoId(video.getId());

                // 统计各状态字幕数量
                long approvedCount = subtitles.stream().filter(s -> s.getStatus() != null && s.getStatus() == 1).count();
                long pendingCount = subtitles.stream().filter(s -> s.getStatus() != null && s.getStatus() == 0).count();
                long systemCount = subtitles.stream().filter(s -> s.getStatus() != null && s.getStatus() == 3).count();

                map.put("approvedCount", (int) approvedCount);
                map.put("pendingCount", (int) pendingCount);
                map.put("systemCount", (int) systemCount);

                // 是否有默认字幕
                boolean hasDefault = subtitles.stream().anyMatch(Subtitle::getIsDefault);
                map.put("hasDefaultSubtitle", hasDefault);

                // 扫描系统字幕目录，检查待入库字幕
                List<Map<String, Object>> pendingImportSubtitles = scanSystemSubtitles(video.getId());
                map.put("pendingImportSubtitles", pendingImportSubtitles);
                map.put("pendingImportCount", (int) pendingImportSubtitles.stream()
                    .filter(s -> "pending".equals(s.get("status"))).count());

                return map;
            })
            .collect(Collectors.toList());

        // 分页处理
        int total = videoList.size();
        int start = (page - 1) * size;
        int end = Math.min(start + size, total);
        List<Map<String, Object>> pageList = start < total ? videoList.subList(start, end) : new ArrayList<>();

        Map<String, Object> result = new HashMap<>();
        result.put("list", pageList);
        result.put("total", total);
        result.put("page", page);
        result.put("size", size);

        return Result.success(result);
    }

    @Override
    public List<Subtitle> getSubtitlesByVideoId(Integer videoId) {
        return subtitleRepository.findByVideoId(videoId);
    }

    @Override
    public Subtitle uploadSubtitle(Integer videoId, MultipartFile file, String language, String languageName, Integer uploadedBy) {
        try {
            // 读取SRT内容
            String srtContent;
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(file.getInputStream(), StandardCharsets.UTF_8))) {
                srtContent = reader.lines().collect(Collectors.joining("\n"));
            }
            
            // 解析SRT
            List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
            
            // 保存到用户上传目录
            Video video = videoMapper.selectById(videoId);
            String uploadId = UUID.randomUUID().toString();
            String userSubtitleDir = uploadFilePathUtils.getUserSubtitleDir(video.getManuscriptId(), videoId, uploadId);
            uploadFilePathUtils.ensureDirectoryExists(userSubtitleDir);
            String filePath = userSubtitleDir + "/subtitle.srt";
            Files.write(Paths.get(filePath), srtContent.getBytes(StandardCharsets.UTF_8));
            
            // 创建字幕记录
            Subtitle subtitle = new Subtitle();
            subtitle.setVideoId(videoId);
            subtitle.setLanguage(language);
            subtitle.setLanguageName(languageName.isEmpty() ? language : languageName);
            subtitle.setFormat("srt");
            subtitle.setContent(items);
            subtitle.setIsDefault(false); // 管理员上传也需要审核
            subtitle.setUploadedBy(uploadedBy != null ? uploadedBy : 0);
            subtitle.setUploadTime(new Date());
            subtitle.setStatus(1); // 管理员上传直接通过
            subtitle.setSource("admin");
            subtitle.setVersion(1);
            subtitle.setUploadId(uploadId);
            subtitle.setFilePath(filePath);
            
            return subtitleRepository.save(subtitle);
        } catch (Exception e) {
            throw new RuntimeException("字幕上传失败: " + e.getMessage(), e);
        }
    }

    @Override
    public Subtitle importSrtToMongo(Integer videoId, String filePath, String language, String languageName, String source) {
        try {
            // 读取SRT文件
            File file = new File(filePath);
            if (!file.exists()) {
                throw new RuntimeException("字幕文件不存在: " + filePath);
            }
            
            String srtContent = new String(Files.readAllBytes(Paths.get(filePath)), StandardCharsets.UTF_8);
            
            // 解析SRT
            List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
            
            // 删除已存在的同语言字幕
            subtitleRepository.findByVideoIdAndLanguage(videoId, language)
                .ifPresent(existing -> subtitleRepository.delete(existing));
            
            // 创建字幕记录
            Subtitle subtitle = new Subtitle();
            subtitle.setVideoId(videoId);
            subtitle.setLanguage(language);
            subtitle.setLanguageName(languageName != null ? languageName : language);
            subtitle.setFormat("srt");
            subtitle.setContent(items);
            subtitle.setIsDefault(true);
            subtitle.setUploadedBy(0);
            subtitle.setUploadTime(new Date());
            subtitle.setStatus(3); // 系统生成
            subtitle.setSource(source != null ? source : "import");
            subtitle.setVersion(1);
            subtitle.setFilePath(filePath);
            
            return subtitleRepository.save(subtitle);
        } catch (Exception e) {
            throw new RuntimeException("字幕入库失败: " + e.getMessage(), e);
        }
    }

    @Override
    public void approveSubtitle(String subtitleId, Integer reviewerId) {
        Subtitle subtitle = subtitleRepository.findById(subtitleId)
            .orElseThrow(() -> new RuntimeException("字幕不存在"));
        
        Integer videoId = subtitle.getVideoId();
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            throw new RuntimeException("视频信息不完整");
        }
        
        // 如果是用户上传的字幕，复制到系统字幕目录
        if ("user".equals(subtitle.getSource()) && subtitle.getFilePath() != null) {
            try {
                String language = subtitle.getLanguage();
                String systemSubtitlePath = uploadFilePathUtils.getSubtitlePath(
                    video.getManuscriptId(), videoId, language);
                
                // 确保目录存在
                uploadFilePathUtils.ensureDirectoryExists(
                    uploadFilePathUtils.getVideoSubtitleDir(video.getManuscriptId(), videoId));
                
                // 复制文件到系统字幕目录
                File sourceFile = new File(subtitle.getFilePath());
                File targetFile = new File(systemSubtitlePath);
                
                if (sourceFile.exists()) {
                    Files.copy(sourceFile.toPath(), targetFile.toPath(), 
                        java.nio.file.StandardCopyOption.REPLACE_EXISTING);
                    System.out.println("[字幕审核] 用户字幕已复制到系统目录: " + systemSubtitlePath);
                    
                    // 更新字幕记录的filePath为系统路径
                    subtitle.setFilePath(systemSubtitlePath);
                }
            } catch (Exception e) {
                System.err.println("[字幕审核] 复制字幕文件失败: " + e.getMessage());
                // 不阻止审核流程，继续执行
            }
        }
        
        // 将该视频的其他字幕设为非默认
        List<Subtitle> videoSubtitles = subtitleRepository.findByVideoId(videoId);
        for (Subtitle sub : videoSubtitles) {
            if (!sub.getId().equals(subtitleId) && Boolean.TRUE.equals(sub.getIsDefault())) {
                sub.setIsDefault(false);
                subtitleRepository.save(sub);
            }
        }
        
        // 更新当前字幕
        subtitle.setStatus(1); // 审核通过
        subtitle.setReviewerId(reviewerId);
        subtitle.setReviewTime(new Date());
        subtitle.setIsDefault(true);
        subtitleRepository.save(subtitle);
    }

    @Override
    public void rejectSubtitle(String subtitleId, Integer reviewerId, String reason) {
        Subtitle subtitle = subtitleRepository.findById(subtitleId)
            .orElseThrow(() -> new RuntimeException("字幕不存在"));
        
        subtitle.setStatus(2); // 审核拒绝
        subtitle.setReviewerId(reviewerId);
        subtitle.setReviewTime(new Date());
        subtitle.setReviewReason(reason);
        subtitleRepository.save(subtitle);
    }

    @Override
    public void setDefaultSubtitle(String subtitleId) {
        Subtitle subtitle = subtitleRepository.findById(subtitleId)
            .orElseThrow(() -> new RuntimeException("字幕不存在"));
        
        // 将该视频的其他字幕设为非默认
        List<Subtitle> videoSubtitles = subtitleRepository.findByVideoId(subtitle.getVideoId());
        for (Subtitle sub : videoSubtitles) {
            if (!sub.getId().equals(subtitleId)) {
                sub.setIsDefault(false);
                subtitleRepository.save(sub);
            }
        }
        
        // 设置当前字幕为默认
        subtitle.setIsDefault(true);
        subtitleRepository.save(subtitle);
    }

    @Override
    public void deleteSubtitle(String subtitleId) {
        subtitleRepository.deleteById(subtitleId);
    }

    @Override
    public List<Map<String, Object>> getPendingSubtitles() {
        List<Subtitle> pendingSubtitles = subtitleRepository.findByStatus(0); // 0=待审核

        return pendingSubtitles.stream().map(subtitle -> {
            Map<String, Object> map = new HashMap<>();
            map.put("id", subtitle.getId());
            map.put("videoId", subtitle.getVideoId());
            map.put("language", subtitle.getLanguage());
            map.put("languageName", subtitle.getLanguageName());
            map.put("status", subtitle.getStatus());
            map.put("uploadId", subtitle.getUploadedBy());
            map.put("createTime", subtitle.getUploadTime());

            // 获取视频标题
            Video video = videoMapper.selectById(subtitle.getVideoId());
            map.put("videoTitle", video != null ? video.getTitle() : "未知视频");

            return map;
        }).collect(Collectors.toList());
    }

    @Override
    public Map<String, Object> previewSubtitle(String subtitleId) {
        Subtitle subtitle = subtitleRepository.findById(subtitleId)
            .orElseThrow(() -> new RuntimeException("字幕不存在"));

        Map<String, Object> result = new HashMap<>();

        // 字幕基本信息
        Map<String, Object> subtitleInfo = new HashMap<>();
        subtitleInfo.put("id", subtitle.getId());
        subtitleInfo.put("videoId", subtitle.getVideoId());
        subtitleInfo.put("language", subtitle.getLanguage());
        subtitleInfo.put("languageName", subtitle.getLanguageName());
        subtitleInfo.put("status", subtitle.getStatus());
        subtitleInfo.put("isDefault", subtitle.getIsDefault());
        subtitleInfo.put("uploadId", subtitle.getUploadedBy());
        subtitleInfo.put("createTime", subtitle.getUploadTime());
        result.put("subtitle", subtitleInfo);

        // 字幕内容（全部返回，前端控制显示）
        result.put("content", subtitle.getContent());

        return result;
    }
    
    /**
     * 解析SRT内容
     */
    private List<Subtitle.SubtitleItem> parseSrtContent(String srtContent) {
        List<Subtitle.SubtitleItem> items = new ArrayList<>();
        
        String[] blocks = srtContent.split("\\n\\s*\\n");
        Pattern timePattern = Pattern.compile("(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})\\s*-->\\s*(\\d{2}):\\s*(\\d{2}):\\s*(\\d{2})[,.](\\d{3})");
        
        for (String block : blocks) {
            block = block.trim();
            if (block.isEmpty()) continue;
            
            String[] lines = block.split("\\n");
            if (lines.length < 3) continue;
            
            try {
                Integer index = Integer.parseInt(lines[0].trim());
                
                Matcher matcher = timePattern.matcher(lines[1]);
                if (!matcher.find()) continue;
                
                double startTime = parseTimeToSeconds(
                    matcher.group(1), matcher.group(2),
                    matcher.group(3), matcher.group(4)
                );
                double endTime = parseTimeToSeconds(
                    matcher.group(5), matcher.group(6),
                    matcher.group(7), matcher.group(8)
                );
                
                StringBuilder textBuilder = new StringBuilder();
                for (int i = 2; i < lines.length; i++) {
                    if (textBuilder.length() > 0) {
                        textBuilder.append("\n");
                    }
                    textBuilder.append(lines[i].trim());
                }
                
                Subtitle.SubtitleItem item = new Subtitle.SubtitleItem();
                item.setIndex(index);
                item.setStartTime(startTime);
                item.setEndTime(endTime);
                item.setText(textBuilder.toString());
                items.add(item);
                
            } catch (Exception e) {
                // 跳过解析失败的块
            }
        }
        
        return items;
    }
    
    /**
     * 将时间字符串转换为秒
     */
    private double parseTimeToSeconds(String hours, String minutes, String seconds, String millis) {
        int h = Integer.parseInt(hours);
        int m = Integer.parseInt(minutes);
        int s = Integer.parseInt(seconds);
        int ms = Integer.parseInt(millis);
        return h * 3600 + m * 60 + s + ms / 1000.0;
    }

    @Override
    public List<Map<String, Object>> scanSystemSubtitles(Integer videoId) {
        List<Map<String, Object>> result = new ArrayList<>();
        
        // 获取视频信息
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            return result;
        }
        
        // 扫描系统字幕目录
        String subtitleDir = uploadFilePathUtils.getVideoSubtitleDir(video.getManuscriptId(), videoId);
        File dir = new File(subtitleDir);
        if (!dir.exists() || !dir.isDirectory()) {
            return result;
        }
        
        // 获取所有.srt文件
        File[] srtFiles = dir.listFiles((d, name) -> name.endsWith(".srt"));
        if (srtFiles == null || srtFiles.length == 0) {
            return result;
        }
        
        // 检查每个字幕文件的入库状态
        for (File file : srtFiles) {
            String fileName = file.getName();
            String language = fileName.replace(".srt", "");
            
            // 检查MongoDB中是否已存在该语言的字幕
            Optional<Subtitle> existing = subtitleRepository.findByVideoIdAndLanguage(videoId, language);
            String status = "pending"; // 默认待入库
            
            if (existing.isPresent()) {
                Subtitle sub = existing.get();
                // 如果来源是system或whisper，表示已入库
                if ("system".equals(sub.getSource()) || "whisper".equals(sub.getSource())) {
                    status = "imported";
                }
            }
            
            Map<String, Object> map = new HashMap<>();
            map.put("language", language);
            map.put("fileName", fileName);
            map.put("filePath", file.getAbsolutePath());
            map.put("status", status);
            map.put("fileSize", file.length());
            result.add(map);
        }
        
        return result;
    }

    @Override
    public Subtitle importSystemSubtitle(Integer videoId, String language) {
        // 获取视频信息
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            throw new RuntimeException("视频不存在");
        }
        
        // 读取SRT文件
        String subtitlePath = uploadFilePathUtils.getSubtitlePath(video.getManuscriptId(), videoId, language);
        File subtitleFile = new File(subtitlePath);
        if (!subtitleFile.exists()) {
            throw new RuntimeException("字幕文件不存在: " + subtitlePath);
        }
        
        try {
            // 读取SRT内容
            String srtContent = new String(Files.readAllBytes(Paths.get(subtitlePath)), StandardCharsets.UTF_8);
            
            // 解析SRT
            List<Subtitle.SubtitleItem> items = parseSrtContent(srtContent);
            if (items.isEmpty()) {
                throw new RuntimeException("字幕文件内容为空或格式错误");
            }
            
            // 删除已存在的同语言系统字幕
            subtitleRepository.findByVideoIdAndLanguage(videoId, language)
                .ifPresent(existing -> {
                    if ("system".equals(existing.getSource()) || "whisper".equals(existing.getSource())) {
                        subtitleRepository.delete(existing);
                    }
                });
            
            // 获取语言名称
            String languageName = getLanguageName(language);
            
            // 检查是否已有默认字幕
            boolean hasDefault = subtitleRepository.findByVideoId(videoId).stream()
                .anyMatch(Subtitle::getIsDefault);
            
            // 创建字幕记录
            Subtitle subtitle = new Subtitle();
            subtitle.setVideoId(videoId);
            subtitle.setLanguage(language);
            subtitle.setLanguageName(languageName);
            subtitle.setFormat("srt");
            subtitle.setContent(items);
            subtitle.setIsDefault(!hasDefault); // 如果没有其他默认字幕，设为默认
            subtitle.setUploadedBy(0); // 系统生成
            subtitle.setUploadTime(new Date());
            subtitle.setStatus(3); // 系统生成状态
            subtitle.setSource("system");
            subtitle.setVersion(1);
            subtitle.setFilePath(subtitlePath);
            
            // 保存到MongoDB
            Subtitle saved = subtitleRepository.save(subtitle);
            
            System.out.println("[字幕入库] 成功导入字幕: videoId=" + videoId + ", language=" + language + ", items=" + items.size());
            
            return saved;
            
        } catch (IOException e) {
            throw new RuntimeException("读取字幕文件失败: " + e.getMessage(), e);
        }
    }
    
    /**
     * 获取语言名称
     */
    private String getLanguageName(String language) {
        Map<String, String> languageMap = new HashMap<>();
        languageMap.put("zh-CN", "简体中文");
        languageMap.put("zh-TW", "繁体中文");
        languageMap.put("en", "English");
        languageMap.put("ja", "日本語");
        languageMap.put("ko", "한국어");
        return languageMap.getOrDefault(language, language);
    }
}
