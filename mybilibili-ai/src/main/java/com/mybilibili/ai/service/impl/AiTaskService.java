package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.common.entity.Video;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.io.File;
import java.nio.file.Files;

@Service
public class AiTaskService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private AiSubtitleService aiSubtitleService;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Autowired
    private StringRedisTemplate redisTemplate;

    public void processSubtitleTask(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            System.err.println("[字幕任务] 视频不存在: " + videoId);
            return;
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            System.err.println("[字幕任务] 视频未关联稿件: " + videoId);
            return;
        }

        System.out.println("[字幕任务] 开始处理视频: " + videoId);

        updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_GENERATING);

        try {
            boolean success = aiSubtitleService.generateSubtitle(manuscriptId, videoId);

            if (success) {
                String subtitlePath = aiSubtitleService.getSubtitlePath(manuscriptId, videoId);
                String subtitleContent = SubtitleTextUtils.extractPlainText(subtitlePath);

                redisTemplate.opsForValue().set("subtitle:" + videoId, subtitleContent);

                updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
                updateVideoHasSubtitle(videoId, 1);

                System.out.println("[字幕任务] 处理完成: " + videoId);
            } else {
                updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
                System.err.println("[字幕任务] 处理失败: " + videoId);
            }

        } catch (Exception e) {
            System.err.println("[字幕任务] 处理异常: " + e.getMessage());
            e.printStackTrace();
            updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
        }
    }

    public void processSummaryTask(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            System.err.println("[摘要任务] 视频不存在: " + videoId);
            return;
        }

        Integer manuscriptId = video.getManuscriptId();
        if (manuscriptId == null) {
            System.err.println("[摘要任务] 视频未关联稿件: " + videoId);
            return;
        }

        System.out.println("[摘要任务] 开始处理视频: " + videoId);

        updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_SUMMARIZING);

        try {
            String subtitlePath = uploadFilePathConfig.getChineseSubtitlePath(manuscriptId, videoId);
            String subtitlePlainText = "";

            File subtitleFile = new File(subtitlePath);
            if (subtitleFile.exists()) {
                subtitlePlainText = SubtitleTextUtils.extractPlainText(subtitlePath);
                System.out.println("[摘要任务] 字幕内容长度: " + subtitlePlainText.length() + " 字符");
            } else {
                System.out.println("[摘要任务] 字幕文件不存在，将使用视频信息生成降级摘要");
            }

            String videoTitle = video.getTitle();
            String videoDescription = "";

            String summary = aiSummaryService.generateSummary(subtitlePlainText, videoTitle, videoDescription);

            System.out.println("[摘要任务] AI总结完成，内容长度: " + summary.length() + " 字符");

            String summaryPath = uploadFilePathConfig.getAiSummaryPath(manuscriptId, videoId);
            boolean saved = aiSummaryService.saveSummaryToFile(summary, summaryPath, videoTitle);

            if (saved) {
                System.out.println("[摘要任务] 摘要已保存到: " + summaryPath);
            }

            redisTemplate.opsForValue().set("summary:" + videoId, summary);

            updateVideoStatus(videoId, Video.PROCESS_STATUS_COMPLETED);
            updateVideoHasSummary(videoId, 1);

            System.out.println("[摘要任务] 处理完成: " + videoId);

        } catch (Exception e) {
            System.err.println("[摘要任务] 处理异常: " + e.getMessage());
            e.printStackTrace();
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
        }
    }

    private void updateVideoStatus(Integer videoId, Integer status) {
        try {
            videoMapper.updateProcessStatus(videoId, status);
            System.out.println("[视频状态] 视频 " + videoId + " 状态更新为: " + status);
        } catch (Exception e) {
            System.err.println("[视频状态] 更新失败: " + e.getMessage());
        }
    }

    private void updateVideoHasSubtitle(Integer videoId, Integer hasSubtitle) {
        try {
            videoMapper.updateHasSubtitle(videoId, hasSubtitle);
        } catch (Exception e) {
            System.err.println("[视频状态] 更新字幕标记失败: " + e.getMessage());
        }
    }

    private void updateVideoHasSummary(Integer videoId, Integer hasSummary) {
        try {
            videoMapper.updateHasSummary(videoId, hasSummary);
        } catch (Exception e) {
            System.err.println("[视频状态] 更新摘要标记失败: " + e.getMessage());
        }
    }
}