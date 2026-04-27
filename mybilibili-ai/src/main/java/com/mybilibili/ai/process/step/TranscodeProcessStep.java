package com.mybilibili.ai.process.step;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.StepExecutionResult;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessStateService;
import com.mybilibili.ai.process.VideoProcessStep;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.ai.service.VideoTranscodeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class TranscodeProcessStep implements VideoProcessStep {

    @Autowired
    private VideoTranscodeService videoTranscodeService;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoProcessStateService stateService;

    @Override
    public VideoProcessStepType getStepType() {
        return VideoProcessStepType.TRANSCODE;
    }

    @Override
    public StepExecutionResult execute(VideoProcessContext context) {
        VideoTranscodeService.StepTranscodeResult result = videoTranscodeService.transcodeStep(
                context.getManuscriptId(),
                context.getVideoId(),
                (progress, stageText) -> stateService.markStepProgress(context, progress, stageText)
        );
        if (!result.isSuccess()) {
            return StepExecutionResult.fail(result.getErrorMessage());
        }
        videoMapper.updateTranscodeResult(
                context.getVideoId(),
                VideoProcessStepType.TRANSCODE.getSuccessStatus(),
                VideoProcessStepType.TRANSCODE.getSuccessProgress(),
                VideoProcessStepType.TRANSCODE.getSuccessStageCode(),
                result.getPlayUrlHd(),
                result.getPlayUrlSd(),
                result.getPlayUrlLd()
        );
        return StepExecutionResult.success();
    }
}
