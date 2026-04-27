package com.mybilibili.ai.process;

public interface VideoProcessOrchestrator {

    StepExecutionResult executeStep(VideoProcessContext context);
}
