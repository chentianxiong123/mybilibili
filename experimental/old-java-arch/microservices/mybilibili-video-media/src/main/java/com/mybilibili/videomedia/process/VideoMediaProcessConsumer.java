package com.mybilibili.videomedia.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessMessage;
import org.apache.rocketmq.spring.annotation.ConsumeMode;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_VIDEO_PROCESS,
        consumerGroup = MQConstants.GROUP_VIDEO_MEDIA_PROCESS,
        selectorExpression = VideoProcessMessage.PROCESS_TYPE_TRANSCODE + " || " + VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO,
        consumeMode = ConsumeMode.ORDERLY
)
public class VideoMediaProcessConsumer implements RocketMQListener<VideoProcessMessage> {

    private final VideoMediaTranscodeService transcodeService;
    private final VideoMediaAudioService audioService;
    private final VideoMediaProcessStateService stateService;
    private final VideoMQProducer videoMQProducer;

    public VideoMediaProcessConsumer(VideoMediaTranscodeService transcodeService,
                                     VideoMediaAudioService audioService,
                                     VideoMediaProcessStateService stateService,
                                     VideoMQProducer videoMQProducer) {
        this.transcodeService = transcodeService;
        this.audioService = audioService;
        this.stateService = stateService;
        this.videoMQProducer = videoMQProducer;
    }

    @Override
    public void onMessage(VideoProcessMessage message) {
        if (VideoProcessMessage.PROCESS_TYPE_TRANSCODE.equals(message.getProcessType())) {
            handleTranscode(message);
            return;
        }
        if (VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO.equals(message.getProcessType())) {
            handleAudioExtract(message);
        }
    }

    private void handleTranscode(VideoProcessMessage message) {
        Integer videoId = message.getVideoId();
        stateService.markProgress(videoId, Video.PROCESS_STATUS_TRANSCODING, 0, "TRANSCODING", null);
        VideoMediaTranscodeService.TranscodeResult result = transcodeService.transcode(
                message.getManuscriptId(),
                videoId,
                (progress, stageText) -> stateService.markProgress(videoId, Video.PROCESS_STATUS_TRANSCODING, progress, "TRANSCODING", null)
        );
        if (!result.isSuccess()) {
            stateService.markFailed(videoId, Video.PROCESS_STATUS_TRANSCODE_FAILED, "TRANSCODING", result.getErrorMessage());
            return;
        }
        stateService.markTranscodeSuccess(videoId, result.getPlayUrlHd(), result.getPlayUrlSd(), result.getPlayUrlLd());
        dispatchNextIfNeeded(message, VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO);
    }

    private void handleAudioExtract(VideoProcessMessage message) {
        Integer videoId = message.getVideoId();
        stateService.markProgress(videoId, Video.PROCESS_STATUS_AUDIO_EXTRACTING, 0, "AUDIO_EXTRACTING", null);
        boolean success = audioService.extractAudio(message.getManuscriptId(), videoId);
        if (!success) {
            stateService.markFailed(videoId, Video.PROCESS_STATUS_AUDIO_FAILED, "AUDIO_EXTRACTING", "Audio extraction failed");
            return;
        }
        stateService.markProgress(videoId, Video.PROCESS_STATUS_AUDIO_SUCCESS, 100, "AUDIO_SUCCESS", null);
        dispatchNextIfNeeded(message, VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE);
    }

    private void dispatchNextIfNeeded(VideoProcessMessage current, String nextProcessType) {
        if (VideoProcessMessage.PROCESS_MODE_MANUAL_SINGLE.equals(current.getProcessMode())) {
            return;
        }
        VideoProcessMessage next = VideoProcessMessage.of(
                current.getManuscriptId(),
                current.getVideoId(),
                current.getUploaderId(),
                nextProcessType
        );
        next.setProcessMode(current.getProcessMode());
        next.setPriority(current.getPriority());
        videoMQProducer.sendVideoProcessMessage(next);
    }
}
