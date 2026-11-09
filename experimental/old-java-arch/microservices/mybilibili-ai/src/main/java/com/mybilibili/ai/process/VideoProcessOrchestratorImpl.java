package com.mybilibili.ai.process;

import com.mybilibili.ai.service.VideoProcessingStorageService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class VideoProcessOrchestratorImpl implements VideoProcessOrchestrator {

    @Autowired
    private VideoProcessStepRegistry stepRegistry;

    @Autowired
    private VideoProcessStateService stateService;

    @Autowired
    private VideoProcessDispatcher dispatcher;

    @Autowired
    private VideoProcessingStorageService processingStorageService;

    @Override
    public StepExecutionResult executeStep(VideoProcessContext context) {
        VideoProcessStep step = stepRegistry.getStep(context.getCurrentStep());
        stateService.markStepStarted(context);

        StepExecutionResult result;
        try {
            result = step.execute(context);
        } catch (Exception e) {
            result = StepExecutionResult.fail(e.getMessage());
        }

        if (!result.isSuccess()) {
            stateService.markStepFailed(context, result.getErrorMessage() == null ? "处理失败" : result.getErrorMessage());
            cleanup(context);
            return result;
        }

        if (context.getCurrentStep() == VideoProcessStepType.AI_SUMMARY) {
            stateService.markProcessCompleted(context);
            cleanup(context);
            return result;
        }

        stateService.markStepSucceeded(context);
        if (context.getProcessMode() == ProcessMode.AUTO_CHAIN) {
            context.getCurrentStep().next().ifPresent(next -> dispatcher.dispatchNext(context, next));
        } else {
            cleanup(context);
        }
        return result;
    }

    private void cleanup(VideoProcessContext context) {
        processingStorageService.cleanupWorkDir(context.getManuscriptId(), context.getVideoId());
    }
}
