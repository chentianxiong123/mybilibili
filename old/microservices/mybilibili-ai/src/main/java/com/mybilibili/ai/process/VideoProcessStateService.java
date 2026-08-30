package com.mybilibili.ai.process;

public interface VideoProcessStateService {

    void markStepStarted(VideoProcessContext context);

    void markStepProgress(VideoProcessContext context, int progress, String stageText);

    void markStepSucceeded(VideoProcessContext context);

    void markStepFailed(VideoProcessContext context, String errorMessage);

    void markProcessCompleted(VideoProcessContext context);
}
