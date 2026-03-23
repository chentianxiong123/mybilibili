package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.ManuscriptMapper;
import com.mybilibili.admin.mapper.VideoMapper;
import com.mybilibili.admin.repository.SubtitleRepository;
import com.mybilibili.admin.service.AiSummaryService;
import com.mybilibili.admin.service.VideoProcessService;
import com.mybilibili.admin.utils.FFmpegUtils;
import com.mybilibili.admin.utils.SubtitleTextUtils;
import com.mybilibili.admin.utils.UploadFilePathUtils;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Subtitle;
import com.mybilibili.common.entity.Video;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.*;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Date;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * 视频处理服务实现
 */
@Service
public class VideoProcessServiceImpl implements VideoProcessService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Autowired
    private FFmpegUtils ffmpegUtils;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Value("${ai.whisper.cli-path:whisper}")
    private String whisperCliPath;

    @Value("${ai.whisper.model-path:base}")
    private String whisperModelPath;

    @Value("${ai.whisper.language:zh}")
    private String whisperLanguage;

    @Value("${ai.whisper.threads:4}")
    private int whisperThreads;

    @Override
    public void transcodeVideo(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            System.err.println("[视频处理] 视频不存在: " + videoId);
            return;
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            System.err.println("[视频处理] 视频未关联稿件: " + videoId);
            return;
        }

        System.out.println("[视频处理] 开始转码视频: " + videoId);
        
        // 更新状态为转码中
        updateProcessStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);
        
        try {
            // 获取源视频路径
            String sourceVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, videoId);
            File sourceFile = new File(sourceVideoPath);
            if (!sourceFile.exists()) {
                throw new RuntimeException("源视频文件不存在: " + sourceVideoPath);
            }
            
            // 获取输出目录
            String transcodedDir = uploadFilePathUtils.getVideoTranscodedDir(manuscriptId, videoId);
            uploadFilePathUtils.ensureDirectoryExists(transcodedDir);
            
            System.out.println("[视频处理] 源视频: " + sourceVideoPath);
            System.out.println("[视频处理] 输出目录: " + transcodedDir);
            
            // 调用 FFmpeg 转码
            ffmpegUtils.transcodeVideo(sourceVideoPath, transcodedDir, videoId, 
                new FFmpegUtils.VideoTranscodeCallback() {
                    @Override
                    public void onTranscodeComplete(String hdPath, String sdPath, String ldPath) {
                        try {
                            // 获取视频时长
                            int duration = ffmpegUtils.getVideoDuration(hdPath);
                            int minutes = duration / 60;
                            int seconds = duration % 60;
                            String durationStr = String.format("%02d:%02d", minutes, seconds);
                            
                            // 更新视频信息
                            Video updateVideo = new Video();
                            updateVideo.setId(videoId);
                            updateVideo.setPlayUrlHd(uploadFilePathUtils.getHdVideoUrl(manuscriptId, videoId));
                            updateVideo.setPlayUrlSd(uploadFilePathUtils.getSdVideoUrl(manuscriptId, videoId));
                            updateVideo.setPlayUrlLd(uploadFilePathUtils.getLdVideoUrl(manuscriptId, videoId));
                            updateVideo.setDurationSeconds(duration);
                            // 转码成功后，状态改为11（转码成功，等待音频提取）
                            updateVideo.setProcessStatus(Video.PROCESS_STATUS_TRANSCODE_SUCCESS);
                            videoMapper.update(updateVideo);
                            
                            System.out.println("[视频处理] 转码完成: " + videoId);
                            System.out.println("[视频处理] 时长: " + durationStr);
                            
                        } catch (Exception e) {
                            System.err.println("[视频处理] 更新视频信息失败: " + e.getMessage());
                            e.printStackTrace();
                        }
                    }
                    
                    @Override
                    public void onTranscodeError(String errorMessage) {
                        System.err.println("[视频处理] 转码失败: " + errorMessage);
                        
                        Video updateVideo = new Video();
                        updateVideo.setId(videoId);
                        updateVideo.setProcessStatus(Video.PROCESS_STATUS_TRANSCODE_FAILED);
                        updateVideo.setProcessError(errorMessage);
                        videoMapper.update(updateVideo);
                    }
                });
            
        } catch (Exception e) {
            System.err.println("[视频处理] 转码异常: " + e.getMessage());
            e.printStackTrace();
            
            Video updateVideo = new Video();
            updateVideo.setId(videoId);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_TRANSCODE_FAILED);
            updateVideo.setProcessError(e.getMessage());
            videoMapper.update(updateVideo);
        }
    }

    @Override
    public void extractAudio(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            System.err.println("[视频处理] 视频不存在: " + videoId);
            return;
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            System.err.println("[视频处理] 视频未关联稿件: " + videoId);
            return;
        }

        System.out.println("[视频处理] 开始提取音频: " + videoId);
        
        updateProcessStatus(videoId, Video.PROCESS_STATUS_AUDIO_EXTRACTING);
        
        try {
            // 获取源视频路径
            String sourceVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, videoId);
            if (!new File(sourceVideoPath).exists()) {
                throw new RuntimeException("源视频文件不存在: " + sourceVideoPath);
            }
            
            // 获取音频输出路径
            String audioPath = uploadFilePathUtils.getAudioPath(manuscriptId, videoId);
            uploadFilePathUtils.ensureDirectoryExists(uploadFilePathUtils.getVideoAudioDir(manuscriptId, videoId));
            
            System.out.println("[视频处理] 源视频: " + sourceVideoPath);
            System.out.println("[视频处理] 音频输出: " + audioPath);
            
            // 提取音频
            boolean success = ffmpegUtils.extractAudio(sourceVideoPath, audioPath);
            
            if (!success) {
                throw new RuntimeException("音频提取失败");
            }
            
            // 检查音频文件
            File audioFile = new File(audioPath);
            if (!audioFile.exists() || audioFile.length() < 1024) {
                throw new RuntimeException("音频文件未生成或太小");
            }
            
            System.out.println("[视频处理] 音频提取完成: " + videoId);
            System.out.println("[视频处理] 音频大小: " + audioFile.length() + " 字节");
            
            // 音频提取成功后，状态改为21（音频提取成功，等待字幕生成）
            updateProcessStatus(videoId, Video.PROCESS_STATUS_AUDIO_SUCCESS);
            
        } catch (Exception e) {
            System.err.println("[视频处理] 音频提取失败: " + e.getMessage());
            e.printStackTrace();
            
            Video updateVideo = new Video();
            updateVideo.setId(videoId);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_AUDIO_FAILED);
            updateVideo.setProcessError(e.getMessage());
            videoMapper.update(updateVideo);
        }
    }
    
    /**
     * 内部音频提取方法 - 不更新状态，只返回是否成功
     */
    private boolean extractAudioInternal(Integer manuscriptId, Integer videoId) {
        try {
            System.out.println("[视频处理] 内部提取音频: " + videoId);
            
            // 获取源视频路径
            String sourceVideoPath = uploadFilePathUtils.getVideoSourcePath(manuscriptId, videoId);
            if (!new File(sourceVideoPath).exists()) {
                System.err.println("[视频处理] 源视频文件不存在: " + sourceVideoPath);
                return false;
            }
            
            // 获取音频输出路径
            String audioPath = uploadFilePathUtils.getAudioPath(manuscriptId, videoId);
            uploadFilePathUtils.ensureDirectoryExists(uploadFilePathUtils.getVideoAudioDir(manuscriptId, videoId));
            
            System.out.println("[视频处理] 源视频: " + sourceVideoPath);
            System.out.println("[视频处理] 音频输出: " + audioPath);
            
            // 提取音频
            boolean success = ffmpegUtils.extractAudio(sourceVideoPath, audioPath);
            
            if (!success) {
                System.err.println("[视频处理] 音频提取失败");
                return false;
            }
            
            // 检查音频文件
            File audioFile = new File(audioPath);
            if (!audioFile.exists() || audioFile.length() < 1024) {
                System.err.println("[视频处理] 音频文件未生成或太小");
                return false;
            }
            
            System.out.println("[视频处理] 音频提取完成: " + videoId);
            System.out.println("[视频处理] 音频大小: " + audioFile.length() + " 字节");
            
            return true;
            
        } catch (Exception e) {
            System.err.println("[视频处理] 音频提取异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    @Override
    public void generateSubtitle(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            System.err.println("[视频处理] 视频不存在: " + videoId);
            return;
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            System.err.println("[视频处理] 视频未关联稿件: " + videoId);
            return;
        }

        System.out.println("[视频处理] 开始生成字幕: " + videoId);
        
        updateProcessStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_GENERATING);
        
        try {
            // 检查音频文件
            String audioPath = uploadFilePathUtils.getAudioPath(manuscriptId, videoId);
            File audioFile = new File(audioPath);
            
            if (!audioFile.exists()) {
                System.out.println("[视频处理] 音频文件不存在，先提取音频...");
                // 直接提取音频，不通过 extractAudio 方法（避免状态更新冲突）
                boolean extractSuccess = extractAudioInternal(manuscriptId, videoId);
                
                if (!extractSuccess) {
                    throw new RuntimeException("音频提取失败，无法生成字幕");
                }
                
                // 重新检查
                if (!audioFile.exists()) {
                    throw new RuntimeException("音频文件未生成");
                }
            }
            
            // 获取字幕输出目录
            String subtitleDir = uploadFilePathUtils.getVideoSubtitleDir(manuscriptId, videoId);
            uploadFilePathUtils.ensureDirectoryExists(subtitleDir);
            String subtitlePath = uploadFilePathUtils.getChineseSubtitlePath(manuscriptId, videoId);
            
            System.out.println("[视频处理] 音频文件: " + audioPath);
            System.out.println("[视频处理] 字幕输出: " + subtitlePath);
            
            // 调用 Whisper 生成字幕
            boolean success = generateSubtitleWithWhisper(audioPath, subtitleDir);
            
            if (!success) {
                throw new RuntimeException("字幕生成失败");
            }
            
            // 检查字幕文件
            File srtFile = new File(subtitlePath);
            if (!srtFile.exists() || srtFile.length() == 0) {
                throw new RuntimeException("字幕文件未生成");
            }
            
            System.out.println("[视频处理] 字幕生成完成: " + videoId);
            System.out.println("[视频处理] 字幕文件路径: " + subtitlePath);
            System.out.println("[视频处理] 注意：字幕已保存为SRT文件，如需存入数据库请使用字幕管理功能");
            
            // 字幕生成成功后，状态改为31（字幕生成成功，等待AI总结）
            // 注意：不再自动保存到MongoDB，字幕入库将在字幕管理功能中实现
            Video updateVideo = new Video();
            updateVideo.setId(videoId);
            updateVideo.setHasSubtitle(1);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
            videoMapper.update(updateVideo);
            
        } catch (Exception e) {
            System.err.println("[视频处理] 字幕生成失败: " + e.getMessage());
            e.printStackTrace();
            
            Video updateVideo = new Video();
            updateVideo.setId(videoId);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_SUBTITLE_FAILED);
            updateVideo.setProcessError(e.getMessage());
            videoMapper.update(updateVideo);
        }
    }
    
    /**
     * 使用 Whisper 生成字幕
     */
    private boolean generateSubtitleWithWhisper(String audioPath, String outputDir) {
        try {
            System.out.println("[Whisper] 开始生成字幕...");
            System.out.println("[Whisper] 音频: " + audioPath);
            System.out.println("[Whisper] 输出目录: " + outputDir);
            
            // Whisper 命令
            String[] cmd = {
                whisperCliPath,
                "-m", whisperModelPath,
                "-f", audioPath,
                "-l", whisperLanguage,
                "-osrt",
                "-of", "zh-CN",
                "-t", String.valueOf(whisperThreads),
                "-pp"
            };
            
            ProcessBuilder pb = new ProcessBuilder(cmd);
            pb.directory(new File(outputDir));
            pb.redirectErrorStream(true);
            
            Process process = pb.start();
            
            // 读取输出
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println("Whisper: " + line);
                }
            }
            
            int exitCode = process.waitFor();
            System.out.println("[Whisper] 退出码: " + exitCode);
            
            return exitCode == 0;
            
        } catch (Exception e) {
            System.err.println("[Whisper] 生成字幕异常: " + e.getMessage());
            e.printStackTrace();
            return false;
        }
    }

    @Override
    public void aiSummary(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            System.err.println("[视频处理] 视频不存在: " + videoId);
            return;
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            System.err.println("[视频处理] 视频未关联稿件: " + videoId);
            return;
        }

        System.out.println("[视频处理] 开始AI总结: " + videoId);

        updateProcessStatus(videoId, Video.PROCESS_STATUS_AI_SUMMARIZING);

        try {
            // 读取字幕内容
            String subtitlePath = uploadFilePathUtils.getChineseSubtitlePath(manuscriptId, videoId);
            String subtitlePlainText = "";

            File subtitleFile = new File(subtitlePath);
            if (subtitleFile.exists()) {
                // 使用SubtitleTextUtils提取纯文本
                subtitlePlainText = SubtitleTextUtils.extractPlainText(subtitlePath);
                System.out.println("[视频处理] 字幕内容长度: " + subtitlePlainText.length() + " 字符");
                System.out.println("[视频处理] 估算token数: " + SubtitleTextUtils.estimateTokenCount(subtitlePlainText));
            } else {
                System.out.println("[视频处理] 字幕文件不存在，将使用视频信息生成降级摘要");
            }

            // 获取稿件信息
            Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
            String videoTitle = video.getTitle();
            String videoDescription = video.getDescription();

            if (manuscript != null) {
                if (videoTitle == null || videoTitle.isEmpty()) {
                    videoTitle = manuscript.getTitle();
                }
                if (videoDescription == null || videoDescription.isEmpty()) {
                    videoDescription = manuscript.getDescription();
                }
            }

            // 调用AI服务生成摘要
            String summary = aiSummaryService.generateSummary(subtitlePlainText, videoTitle, videoDescription);

            System.out.println("[视频处理] AI总结完成");
            System.out.println("[视频处理] 总结内容长度: " + summary.length() + " 字符");

            // 保存摘要到文件
            String summaryPath = uploadFilePathUtils.getAiSummaryPath(manuscriptId, videoId);
            boolean saved = aiSummaryService.saveSummaryToFile(summary, summaryPath, videoTitle);

            if (saved) {
                System.out.println("[视频处理] 摘要已保存到: " + summaryPath);
            } else {
                System.err.println("[视频处理] 摘要保存失败");
            }

            // 更新视频状态
            Video updateVideo = new Video();
            updateVideo.setId(videoId);
            updateVideo.setHasSummary(1);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_COMPLETED);
            videoMapper.update(updateVideo);

        } catch (Exception e) {
            System.err.println("[视频处理] AI总结失败: " + e.getMessage());
            e.printStackTrace();

            Video updateVideo = new Video();
            updateVideo.setId(videoId);
            updateVideo.setProcessStatus(Video.PROCESS_STATUS_AI_FAILED);
            updateVideo.setProcessError(e.getMessage());
            videoMapper.update(updateVideo);
        }
    }

    @Override
    public void processAll(Integer videoId) {
        System.out.println("[视频处理] 开始全流程处理: " + videoId);
        
        // 1. 视频转码
        transcodeVideo(videoId);
        
        // 2. 音频提取
        extractAudio(videoId);
        
        // 3. 字幕生成
        generateSubtitle(videoId);
        
        // 4. AI总结
        aiSummary(videoId);
        
        System.out.println("[视频处理] 全流程处理完成: " + videoId);
    }
    
    /**
     * 读取字幕文件内容
     */
    private String readSubtitleContent(String subtitlePath) throws IOException {
        StringBuilder content = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(new FileReader(subtitlePath))) {
            String line;
            while ((line = reader.readLine()) != null) {
                // 跳过时间码和序号行
                if (!line.matches("\\d+") && 
                    !line.matches("\\d{2}:\\d{2}:\\d{2},\\d{3} --> \\d{2}:\\d{2}:\\d{2},\\d{3}") &&
                    !line.trim().isEmpty()) {
                    content.append(line).append(" ");
                }
            }
        }
        return content.toString().trim();
    }
    
    /**
     * 模拟AI总结（TODO: 接入真实AI服务）
     */
    private String generateAiSummaryMock(String subtitleContent, Video video) {
        StringBuilder summary = new StringBuilder();
        summary.append("【视频摘要】\n");
        summary.append("视频标题: ").append(video.getTitle()).append("\n\n");
        summary.append("【内容概述】\n");
        summary.append("本视频");
        if (video.getDurationSeconds() != null) {
            int minutes = video.getDurationSeconds() / 60;
            summary.append("时长约 ").append(minutes).append("分钟,");
        }
        summary.append("主要内容包括相关主题的介绍和讲解。\n\n");
        summary.append("【关键要点】\n");
        summary.append("1. 视频涵盖了核心知识点\n");
        summary.append("2. 详细讲解了相关概念\n");
        summary.append("3. 提供了实用的示例和说明\n\n");
        summary.append("【字幕内容预览】\n");
        summary.append(subtitleContent.substring(0, Math.min(200, subtitleContent.length())));
        if (subtitleContent.length() > 200) {
            summary.append("...");
        }
        summary.append("\n\n");
        summary.append("【注意事项】\n");
        summary.append("当前为模拟AI总结，请接入 DeepSeek 或其他大模型API获取真实总结。");
        
        return summary.toString();
    }

    private void updateProcessStatus(Integer videoId, Integer status) {
        Video updateVideo = new Video();
        updateVideo.setId(videoId);
        updateVideo.setProcessStatus(status);
        videoMapper.update(updateVideo);
        System.out.println("[视频处理] 视频 " + videoId + " 状态更新为: " + getStatusText(status));
    }

    private String getStatusText(Integer status) {
        if (status == null) return "未知";
        switch (status) {
            // 0x: 初始状态
            case Video.PROCESS_STATUS_PENDING: return "待处理";
            
            // 1x: 视频转码
            case Video.PROCESS_STATUS_TRANSCODING: return "视频转码中";
            case Video.PROCESS_STATUS_TRANSCODE_FAILED: return "转码失败";
            case Video.PROCESS_STATUS_TRANSCODE_SUCCESS: return "转码成功";
            
            // 2x: 音频提取
            case Video.PROCESS_STATUS_AUDIO_EXTRACTING: return "音频提取中";
            case Video.PROCESS_STATUS_AUDIO_FAILED: return "音频提取失败";
            case Video.PROCESS_STATUS_AUDIO_SUCCESS: return "音频提取成功";
            
            // 3x: 字幕生成
            case Video.PROCESS_STATUS_SUBTITLE_GENERATING: return "字幕生成中";
            case Video.PROCESS_STATUS_SUBTITLE_FAILED: return "字幕生成失败";
            case Video.PROCESS_STATUS_SUBTITLE_SUCCESS: return "字幕生成成功";
            
            // 4x: AI总结
            case Video.PROCESS_STATUS_AI_SUMMARIZING: return "AI总结中";
            case Video.PROCESS_STATUS_AI_FAILED: return "AI总结失败";
            case Video.PROCESS_STATUS_AI_SUCCESS: return "AI总结成功";
            
            // 5: 全部完成
            case Video.PROCESS_STATUS_COMPLETED: return "处理完成";
            
            default: return "未知(" + status + ")";
        }
    }
    
    /**
     * 读取 SRT 字幕文件内容
     */
    private String readSrtFile(String subtitlePath) {
        try {
            File file = new File(subtitlePath);
            if (!file.exists()) {
                return null;
            }
            return new String(Files.readAllBytes(Paths.get(subtitlePath)));
        } catch (Exception e) {
            System.err.println("[视频处理] 读取字幕文件失败: " + e.getMessage());
            return null;
        }
    }
    
    /**
     * 解析 SRT 字幕内容
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
                continue;
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
}
