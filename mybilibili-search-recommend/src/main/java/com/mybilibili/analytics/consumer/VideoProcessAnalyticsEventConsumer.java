package com.mybilibili.analytics.consumer;

import com.mybilibili.analytics.mapper.VideoProcessAnalyticsMapper;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoProcessAnalyticsEvent;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_VIDEO_PROCESS_ANALYTICS,
        consumerGroup = MQConstants.GROUP_VIDEO_PROCESS_ANALYTICS
)
public class VideoProcessAnalyticsEventConsumer implements RocketMQListener<VideoProcessAnalyticsEvent> {

    private final VideoProcessAnalyticsMapper videoProcessAnalyticsMapper;

    public VideoProcessAnalyticsEventConsumer(VideoProcessAnalyticsMapper videoProcessAnalyticsMapper) {
        this.videoProcessAnalyticsMapper = videoProcessAnalyticsMapper;
    }

    @Override
    public void onMessage(VideoProcessAnalyticsEvent event) {
        if (event == null || event.getVideoId() == null || event.getToStatus() == null) {
            return;
        }
        LocalDateTime occurredAt = event.getOccurredAt() != null ? event.getOccurredAt() : LocalDateTime.now();
        videoProcessAnalyticsMapper.insertProcessEvent(
                event.getVideoId(),
                event.getManuscriptId(),
                event.getFromStatus(),
                event.getToStatus(),
                event.getStage(),
                event.getProgress(),
                event.getErrorMessage(),
                event.getOperatorType(),
                event.getOperatorId(),
                occurredAt
        );
    }
}
