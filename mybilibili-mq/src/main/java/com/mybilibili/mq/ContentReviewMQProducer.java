package com.mybilibili.mq;

import org.apache.rocketmq.client.producer.SendCallback;
import org.apache.rocketmq.client.producer.SendResult;
import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class ContentReviewMQProducer {

    @Autowired
    private RocketMQTemplate rocketMQTemplate;

    public void sendContentReviewMessage(ContentReviewMessage message) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_CONTENT_REVIEW,
            message,
            new SendCallback() {
                @Override
                public void onSuccess(SendResult sendResult) {
                    System.out.println("内容审核消息发送成功: reportId=" + message.getReportId()
                        + ", msgId=" + sendResult.getMsgId());
                }

                @Override
                public void onException(Throwable e) {
                    System.err.println("内容审核消息发送失败: reportId=" + message.getReportId()
                        + ", error=" + e.getMessage());
                }
            }
        );
    }

    public void sendContentReviewResultEvent(ContentReviewResultEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_CONTENT_REVIEW_RESULT,
            event,
            new SendCallback() {
                @Override
                public void onSuccess(SendResult sendResult) {
                    // Review result delivery is a background side effect.
                }

                @Override
                public void onException(Throwable e) {
                    System.err.println("内容审核结果发送失败: reportId="
                        + (event == null ? null : event.getReportId())
                        + ", error=" + e.getMessage());
                }
            }
        );
    }

    public void sendContentModerationMessage(ContentModerationMessage message) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_CONTENT_MODERATION,
            message,
            quietCallback("内容异步审核消息", message == null ? null : message.getTargetId())
        );
    }

    public void sendContentModerationResultEvent(ContentModerationResultEvent event) {
        rocketMQTemplate.asyncSend(
            MQConstants.TOPIC_CONTENT_MODERATION_RESULT,
            event,
            quietCallback("内容异步审核结果", event == null ? null : event.getTargetId())
        );
    }

    private SendCallback quietCallback(String label, Integer targetId) {
        return new SendCallback() {
            @Override
            public void onSuccess(SendResult sendResult) {
                // Background moderation events are intentionally quiet on success.
            }

            @Override
            public void onException(Throwable e) {
                System.err.println(label + "发送失败: targetId=" + targetId + ", error=" + e.getMessage());
            }
        };
    }
}
