package com.mybilibili.ai.consumer;

import com.mybilibili.ai.service.ContentReviewService;
import com.mybilibili.mq.ContentReviewMessage;
import com.mybilibili.mq.ContentReviewMQProducer;
import com.mybilibili.mq.ContentReviewResultEvent;
import com.mybilibili.mq.MQConstants;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.Map;

@Component
@RocketMQMessageListener(
    topic = MQConstants.TOPIC_CONTENT_REVIEW,
    consumerGroup = MQConstants.GROUP_CONTENT_REVIEW
)
public class ContentReviewConsumer implements RocketMQListener<ContentReviewMessage> {

    @Autowired
    private ContentReviewService contentReviewService;

    @Autowired
    private ContentReviewMQProducer contentReviewMQProducer;

    @Override
    public void onMessage(ContentReviewMessage message) {
        Map<String, Object> reviewResult = contentReviewService.reviewContent(
            message.getContent(), message.getReason());

        String status = reviewResult.get("status") != null ? reviewResult.get("status").toString() : "FAILED";

        contentReviewMQProducer.sendContentReviewResultEvent(
            ContentReviewResultEvent.of(
                message.getReportId(),
                status,
                reviewResult.get("verdict"),
                reviewResult.get("riskLevel"))
        );
    }
}
