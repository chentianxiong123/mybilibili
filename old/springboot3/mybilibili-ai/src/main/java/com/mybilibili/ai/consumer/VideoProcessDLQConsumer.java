package com.mybilibili.ai.consumer;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.mq.VideoProcessMessage;
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

    @Override
    public void onMessage(VideoProcessMessage message) {
        Integer videoId = message.getVideoId();
        try {
            VideoProcessStepType stepType = VideoProcessStepType.fromMqProcessType(message.getProcessType());
            videoMapper.updateProcessError(videoId, stepType.getFailedStatus(), "处理失败，已达到最大重试次数");
        } catch (Exception e) {
            videoMapper.updateProcessError(videoId, VideoProcessStepType.TRANSCODE.getFailedStatus(), "处理失败，已达到最大重试次数");
        }
    }
}
