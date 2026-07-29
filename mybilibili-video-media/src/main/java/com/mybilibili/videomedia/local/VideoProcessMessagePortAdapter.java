package com.mybilibili.videomedia.local;

import com.mybilibili.common.vo.Result;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessMessage;
import com.mybilibili.video.service.VideoProcessPort;
import org.springframework.stereotype.Component;

@Component
public class VideoProcessMessagePortAdapter implements VideoProcessPort {

    private final VideoMQProducer videoMQProducer;

    public VideoProcessMessagePortAdapter(VideoMQProducer videoMQProducer) {
        this.videoMQProducer = videoMQProducer;
    }

    @Override
    public Result<Void> triggerTranscode(Integer videoId, Integer manuscriptId) {
        return sendManualStep(videoId, manuscriptId, VideoProcessMessage.PROCESS_TYPE_TRANSCODE, "转码任务已提交");
    }

    @Override
    public Result<Void> triggerAutoProcess(Integer videoId, Integer manuscriptId, Integer uploaderId) {
        if (videoId == null || manuscriptId == null) {
            return Result.error("视频处理任务缺少必要参数");
        }
        VideoProcessMessage message = VideoProcessMessage.of(
                manuscriptId,
                videoId,
                uploaderId,
                VideoProcessMessage.PROCESS_TYPE_TRANSCODE
        );
        videoMQProducer.sendVideoProcessMessage(message);
        return Result.success("全流程处理任务已提交", null);
    }

    @Override
    public Result<Void> triggerAudioExtract(Integer videoId, Integer manuscriptId) {
        return sendManualStep(videoId, manuscriptId, VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO, "音频提取任务已提交");
    }

    @Override
    public Result<Void> triggerSubtitleGenerate(Integer videoId, Integer manuscriptId) {
        return sendManualStep(videoId, manuscriptId, VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE, "字幕生成任务已提交");
    }

    @Override
    public Result<Void> triggerAiSummary(Integer videoId, Integer manuscriptId) {
        return sendManualStep(videoId, manuscriptId, VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY, "AI总结任务已提交");
    }

    private Result<Void> sendManualStep(Integer videoId, Integer manuscriptId, String processType, String successMessage) {
        if (videoId == null || manuscriptId == null) {
            return Result.error("视频处理任务缺少必要参数");
        }
        // 改为 AUTO_CHAIN：admin 单步触发后,成功会继续走后续步骤
        VideoProcessMessage message = VideoProcessMessage.of(
                manuscriptId,
                videoId,
                null,
                processType
        );
        videoMQProducer.sendVideoProcessMessage(message);
        return Result.success(successMessage, null);
    }
}
