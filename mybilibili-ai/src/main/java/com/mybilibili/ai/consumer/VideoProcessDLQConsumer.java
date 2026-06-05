package com.mybilibili.ai.consumer;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessAnalyticsEvent;
import com.mybilibili.mq.VideoProcessMessage;
import com.mybilibili.mq.VideoProcessProgressEvent;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
@RocketMQMessageListener(
    topic = "%DLQ%video-process-group",
    consumerGroup = "video-process-dlq-group"
)
public class VideoProcessDLQConsumer implements RocketMQListener<VideoProcessMessage> {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Override
    public void onMessage(VideoProcessMessage message) {
        Integer videoId = message.getVideoId();
        Video previous = videoMapper.selectById(videoId);
        VideoProcessStepType stepType;
        try {
            stepType = VideoProcessStepType.fromMqProcessType(message.getProcessType());
        } catch (Exception e) {
            stepType = VideoProcessStepType.AI_SUMMARY;
        }

        String errorMessage = "处理失败，已达到最大重试次数";
        videoMapper.updateProcessState(videoId, stepType.getFailedStatus(), 0, stepType.getRunningStageCode(), errorMessage);

        Integer manuscriptId = previous != null ? previous.getManuscriptId() : message.getManuscriptId();
        String title = previous != null ? previous.getTitle() : null;
        Integer fromStatus = previous != null ? previous.getProcessStatus() : null;
        videoMQProducer.sendVideoProcessAnalyticsEvent(VideoProcessAnalyticsEvent.of(
                videoId,
                manuscriptId,
                fromStatus,
                stepType.getFailedStatus(),
                stepType.getRunningStageCode(),
                0,
                errorMessage,
                "SYSTEM",
                null
        ));
        videoMQProducer.sendVideoProcessProgressEvent(VideoProcessProgressEvent.error(
                videoId,
                manuscriptId,
                title,
                stepType.getRunningStageCode(),
                errorMessage,
                stepType.getFailedStatus(),
                errorMessage,
                "mybilibili-ai"
        ));
    }
}
