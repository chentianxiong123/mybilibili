package com.mybilibili.ai.process;

public interface VideoProcessStep {

    VideoProcessStepType getStepType();

    StepExecutionResult execute(VideoProcessContext context);
}
