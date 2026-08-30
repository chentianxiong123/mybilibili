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
}