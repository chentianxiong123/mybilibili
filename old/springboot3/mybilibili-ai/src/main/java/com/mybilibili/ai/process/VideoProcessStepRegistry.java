package com.mybilibili.ai.process;

import org.springframework.stereotype.Component;

import java.util.EnumMap;
import java.util.List;
import java.util.Map;

@Component
public class VideoProcessStepRegistry {

    private final Map<VideoProcessStepType, VideoProcessStep> stepMap = new EnumMap<>(VideoProcessStepType.class);

    public VideoProcessStepRegistry(List<VideoProcessStep> steps) {
        for (VideoProcessStep step : steps) {
            stepMap.put(step.getStepType(), step);
        }
    }

    public VideoProcessStep getStep(VideoProcessStepType stepType) {
        VideoProcessStep step = stepMap.get(stepType);
        if (step == null) {
            throw new IllegalArgumentException("未找到处理步骤: " + stepType);
        }
        return step;
    }
}
