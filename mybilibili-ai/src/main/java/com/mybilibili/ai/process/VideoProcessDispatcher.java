package com.mybilibili.ai.process;

public interface VideoProcessDispatcher {

    void dispatchNext(VideoProcessContext context, VideoProcessStepType nextStep);
}
