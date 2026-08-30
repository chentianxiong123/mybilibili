package com.mybilibili.videomedia.process;

import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoProcessProgressEvent;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_VIDEO_PROCESS_PROGRESS,
        consumerGroup = MQConstants.GROUP_VIDEO_PROCESS_PROGRESS
)
public class VideoProcessProgressConsumer implements RocketMQListener<VideoProcessProgressEvent> {

    private final VideoProcessProgressSseService sseService;

    public VideoProcessProgressConsumer(VideoProcessProgressSseService sseService) {
        this.sseService = sseService;
    }

    @Override
    public void onMessage(VideoProcessProgressEvent event) {
        sseService.publish(event);
    }
}
