package com.mybilibili.videomedia.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoProcessMessage;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Component
@RocketMQMessageListener(
        topic = "%DLQ%" + MQConstants.GROUP_VIDEO_MEDIA_PROCESS,
        consumerGroup = "video-media-process-dlq-group"
)
public class VideoMediaProcessDLQConsumer implements RocketMQListener<VideoProcessMessage> {

    private final VideoMediaProcessStateService stateService;

    public VideoMediaProcessDLQConsumer(VideoMediaProcessStateService stateService) {
        this.stateService = stateService;
    }

    @Override
    public void onMessage(VideoProcessMessage message) {
        if (VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO.equals(message.getProcessType())) {
            stateService.markFailed(message.getVideoId(), Video.PROCESS_STATUS_AUDIO_FAILED, "AUDIO_EXTRACTING", "Processing failed after retries");
            return;
        }
        stateService.markFailed(message.getVideoId(), Video.PROCESS_STATUS_TRANSCODE_FAILED, "TRANSCODING", "Processing failed after retries");
    }
}
