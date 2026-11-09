package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.ProcessMode;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessOrchestrator;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.common.entity.Video;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class AiTaskService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoProcessOrchestrator orchestrator;

    public void processSubtitleTask(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            return;
        }
        VideoProcessContext context = buildContext(video, VideoProcessStepType.GENERATE_SUBTITLE);
        orchestrator.executeStep(context);
    }

    public void processSummaryTask(Integer videoId) {
        Video video = videoMapper.selectById(videoId);
        if (video == null || video.getManuscriptId() == null) {
            return;
        }
        VideoProcessContext context = buildContext(video, VideoProcessStepType.AI_SUMMARY);
        orchestrator.executeStep(context);
    }

    private VideoProcessContext buildContext(Video video, VideoProcessStepType stepType) {
        VideoProcessContext context = new VideoProcessContext();
        context.setVideoId(video.getId());
        context.setManuscriptId(video.getManuscriptId());
        context.setVideoTitle(video.getTitle());
        context.setCurrentStep(stepType);
        context.setProcessMode(ProcessMode.INTERNAL_SINGLE);
        context.setTriggerSource("AI_TASK");
        return context;
    }
}
