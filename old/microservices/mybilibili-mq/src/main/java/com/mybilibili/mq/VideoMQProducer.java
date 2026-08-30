package com.mybilibili.mq;

import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Slf4j
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
                    log.debug("视频处理消息发送成功: {}", sendResult.getMsgId());
                }

                @Override
                public void onException(Throwable e) {
                    log.warn("视频处理消息发送失败: {}", e.getMessage());
                }
            }
        );
    }

    public void sendManuscriptAnalyticsEvent(ManuscriptAnalyticsEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_MANUSCRIPT_ANALYTICS,
            event,
            quietCallback("稿件统计事件", event == null ? null : event.getManuscriptId())
        );
    }

    public void sendManuscriptCommentCountEvent(ManuscriptCommentCountEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_MANUSCRIPT_COMMENT_COUNT,
            event,
            new org.apache.rocketmq.client.producer.SendCallback() {
                @Override
                public void onSuccess(org.apache.rocketmq.client.producer.SendResult sendResult) {
                    // Comment count events are background counters.
                }

                @Override
                public void onException(Throwable e) {
                    log.warn("稿件评论数事件发送失败: manuscriptId={}, error={}",
                            event == null ? null : event.getManuscriptId(), e.getMessage());
                }
            }
        );
    }

    public void sendVideoProcessAnalyticsEvent(VideoProcessAnalyticsEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_VIDEO_PROCESS_ANALYTICS,
            event,
            quietCallback("视频处理统计事件", event == null ? null : event.getVideoId())
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
                    log.warn("视频处理进度消息发送失败: {}", e.getMessage());
                }
            }
        );
    }

    public void sendVideoPublishEvent(VideoPublishEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_VIDEO_PUBLISH,
            event,
            quietCallback("稿件自动上架事件", event == null ? null : event.getManuscriptId())
        );
    }

    public void sendManuscriptIndexEvent(ManuscriptIndexEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_MANUSCRIPT_INDEX,
            event,
            quietCallback("稿件索引事件", event == null ? null : event.getManuscriptId())
        );
    }

    private org.apache.rocketmq.client.producer.SendCallback quietCallback(String label, Integer targetId) {
        return new org.apache.rocketmq.client.producer.SendCallback() {
            @Override
            public void onSuccess(org.apache.rocketmq.client.producer.SendResult sendResult) {
                // Background events are intentionally quiet on success.
            }

            @Override
            public void onException(Throwable e) {
                log.warn("{}发送失败: targetId={}, error={}", label, targetId, e.getMessage());
            }
        };
    }
}
