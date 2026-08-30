package com.mybilibili.ai.process.step;

import com.mybilibili.ai.process.StepExecutionResult;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessStep;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.ai.process.VideoProcessStateService;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.mapper.VideoMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class SubtitleProcessStep implements VideoProcessStep {

    @Autowired
    private AiSubtitleService aiSubtitleService;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoProcessStateService stateService;

    @Override
    public VideoProcessStepType getStepType() {
        return VideoProcessStepType.GENERATE_SUBTITLE;
    }

    @Override
    public StepExecutionResult execute(VideoProcessContext context) {
        boolean success = aiSubtitleService.generateSubtitle(
                context.getManuscriptId(),
                context.getVideoId(),
                (percent, stageText) -> {
                    stateService.markStepProgress(context, percent, stageText);
                }
        );
        if (!success) {
            return StepExecutionResult.fail("字幕生成失败");
        }
        videoMapper.updateHasSubtitle(context.getVideoId(), 1);
        return StepExecutionResult.success();
    }
}
