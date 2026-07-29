package com.mybilibili.ai.process;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessAnalyticsEvent;
import com.mybilibili.mq.VideoProcessProgressEvent;
import com.mybilibili.mq.VideoPublishEvent;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class VideoProcessStateServiceImpl implements VideoProcessStateService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Override
    public void markStepStarted(VideoProcessContext context) {
        VideoProcessStepType step = context.getCurrentStep();
        updateVideoState(context.getVideoId(), step.getStartedStatus(), step.getStartProgress(), step.getRunningStageCode(), null);
        push(context, step.getRunningStageCode(), step.getRunningStageText(), step.getStartProgress(), step.getStartedStatus());
    }

    @Override
    public void markStepProgress(VideoProcessContext context, int progress, String stageText) {
        VideoProcessStepType step = context.getCurrentStep();
        updateVideoState(context.getVideoId(), step.getStartedStatus(), progress, step.getRunningStageCode(), null);
        push(context, step.getRunningStageCode(), stageText, progress, step.getStartedStatus());
    }

    @Override
    public void markStepSucceeded(VideoProcessContext context) {
        VideoProcessStepType step = context.getCurrentStep();
        updateVideoState(context.getVideoId(), step.getSuccessStatus(), step.getSuccessProgress(), step.getSuccessStageCode(), null);
        push(context, step.getSuccessStageCode(), step.getSuccessStageText(), step.getSuccessProgress(), step.getSuccessStatus());
    }

    @Override
    public void markStepFailed(VideoProcessContext context, String errorMessage) {
        VideoProcessStepType step = context.getCurrentStep();
        updateVideoState(context.getVideoId(), step.getFailedStatus(), 0, step.getRunningStageCode(), errorMessage);
        videoMQProducer.sendVideoProcessProgressEvent(VideoProcessProgressEvent.error(
                context.getVideoId(),
                context.getManuscriptId(),
                context.getVideoTitle(),
                step.getRunningStageCode(),
                errorMessage,
                step.getFailedStatus(),
                errorMessage,
                "mybilibili-ai"
        ));
    }

    @Override
    public void markProcessCompleted(VideoProcessContext context) {
        updateVideoState(context.getVideoId(), Video.PROCESS_STATUS_COMPLETED, 100, "AI_SUCCESS", null);
        videoMQProducer.sendVideoProcessProgressEvent(VideoProcessProgressEvent.complete(
                context.getVideoId(),
                context.getManuscriptId(),
                context.getVideoTitle(),
                "AI_SUCCESS",
                "处理完成",
                Video.PROCESS_STATUS_COMPLETED,
                "mybilibili-ai"
        ));
        // AI 处理链全部完成,触发"自动上架"事件
        VideoPublishEvent publishEvent = new VideoPublishEvent();
        publishEvent.setManuscriptId(context.getManuscriptId());
        publishEvent.setVideoId(context.getVideoId());
        publishEvent.setTrigger("AUTO_CHAIN");
        videoMQProducer.sendVideoPublishEvent(publishEvent);
    }

    private void push(VideoProcessContext context, String stage, String stageText, int progress, Integer status) {
        videoMQProducer.sendVideoProcessProgressEvent(VideoProcessProgressEvent.progress(
                context.getVideoId(),
                context.getManuscriptId(),
                context.getVideoTitle(),
                stage,
                stageText,
                progress,
                status,
                "mybilibili-ai"
        ));
    }

    private void updateVideoState(Integer videoId, Integer status, Integer progress, String stage, String error) {
        Video previous = videoMapper.selectById(videoId);
        int updated = videoMapper.updateProcessState(videoId, status, progress, stage, error);
        if (updated <= 0) {
            return;
        }
        Integer fromStatus = previous != null ? previous.getProcessStatus() : null;
        Integer manuscriptId = previous != null ? previous.getManuscriptId() : null;
        VideoProcessAnalyticsEvent event = VideoProcessAnalyticsEvent.of(
                videoId,
                manuscriptId,
                fromStatus,
                status,
                stage,
                progress,
                error,
                "SYSTEM",
                null
        );
        videoMQProducer.sendVideoProcessAnalyticsEvent(event);
    }
}
