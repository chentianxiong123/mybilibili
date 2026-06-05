package com.mybilibili.mq;

import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class VideoMQProducer {

    @Autowired
    private RocketMQTemplate rocketMQTemplate;

    public void sendVideoProcessMessage(VideoProcessMessage message) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_VIDEO_PROCESS + ":" + message.getProcessType(),
            message,
            new org.apache.rocketmq.client.producer.SendCallback() {
                @Override
                public void onSuccess(org.apache.rocketmq.client.producer.SendResult sendResult) {
                    System.out.println("视频处理消息发送成功: " + sendResult.getMsgId());
                }

                @Override
                public void onException(Throwable e) {
                    System.err.println("视频处理消息发送失败: " + e.getMessage());
                }
            }
        );
    }

    public void sendVideoProcessMessageSync(VideoProcessMessage message) {
        rocketMQTemplate.syncSend(
            MQConstants.TOPIC_VIDEO_PROCESS + ":" + message.getProcessType(),
            message
        );
    }

    public void sendVideoPublishMessage(VideoPublishMessage message) {
        rocketMQTemplate.syncSend(
            MQConstants.TOPIC_VIDEO_PUBLISH,
            message
        );
    }

    public void sendManuscriptAnalyticsEvent(ManuscriptAnalyticsEvent event) {
        rocketMQTemplate.syncSend(
            MQConstants.TOPIC_MANUSCRIPT_ANALYTICS,
            event
        );
    }

    public void sendVideoProcessAnalyticsEvent(VideoProcessAnalyticsEvent event) {
        rocketMQTemplate.syncSend(
            MQConstants.TOPIC_VIDEO_PROCESS_ANALYTICS,
            event
        );
    }

    public void sendVideoProcessProgressEvent(VideoProcessProgressEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_VIDEO_PROCESS_PROGRESS,
            event,
            new org.apache.rocketmq.client.producer.SendCallback() {
                @Override
                public void onSuccess(org.apache.rocketmq.client.producer.SendResult sendResult) {
                    // Progress events are high-frequency; keep success logging quiet.
                }

                @Override
                public void onException(Throwable e) {
                    System.err.println("视频处理进度消息发送失败: " + e.getMessage());
                }
            }
        );
    }
}
