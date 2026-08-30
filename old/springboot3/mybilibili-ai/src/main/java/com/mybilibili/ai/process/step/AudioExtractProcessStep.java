package com.mybilibili.ai.process.step;

import com.mybilibili.ai.process.StepExecutionResult;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessStep;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.ai.service.AudioExtractService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class AudioExtractProcessStep implements VideoProcessStep {

    @Autowired
    private AudioExtractService audioExtractService;

    @Override
    public VideoProcessStepType getStepType() {
        return VideoProcessStepType.EXTRACT_AUDIO;
    }

    @Override
    public StepExecutionResult execute(VideoProcessContext context) {
        boolean success = audioExtractService.extractAudio(context.getManuscriptId(), context.getVideoId());
        return success ? StepExecutionResult.success() : StepExecutionResult.fail("音频提取失败");
    }
}
