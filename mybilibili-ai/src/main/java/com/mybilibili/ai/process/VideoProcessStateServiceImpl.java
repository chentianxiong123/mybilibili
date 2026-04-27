package com.mybilibili.ai.process;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.service.VideoProgressSseService;
import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;
import com.mybilibili.common.entity.Video;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class VideoProcessStateServiceImpl implements VideoProcessStateService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoProgressSseService progressSseService;

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
        VideoProcessWebSocketHandler.broadcastError(
                context.getVideoId(),
                context.getManuscriptId(),
                context.getVideoTitle(),
                step.getRunningStageCode(),
                errorMessage
        );
        progressSseService.pushProcessEvent(videoMapper.selectById(context.getVideoId()), "error", errorMessage);
    }

    @Override
    public void markProcessCompleted(VideoProcessContext context) {
        updateVideoState(context.getVideoId(), Video.PROCESS_STATUS_COMPLETED, 100, "AI_SUCCESS", null);
        VideoProcessWebSocketHandler.broadcastComplete(context.getVideoId(), context.getManuscriptId(), context.getVideoTitle());
        progressSseService.pushProcessEvent(videoMapper.selectById(context.getVideoId()), "complete", "处理完成");
        progressSseService.complete(context.getVideoId());
    }

    private void push(VideoProcessContext context, String stage, String stageText, int progress, Integer status) {
        VideoProcessWebSocketHandler.broadcastProgress(
                context.getVideoId(),
                context.getManuscriptId(),
                context.getVideoTitle(),
                stage,
                stageText,
                progress,
                status
        );
        progressSseService.pushProcessEvent(videoMapper.selectById(context.getVideoId()), "progress", stageText);
    }

    private void updateVideoState(Integer videoId, Integer status, Integer progress, String stage, String error) {
        videoMapper.updateProcessState(videoId, status, progress, stage, error);
    }
}
