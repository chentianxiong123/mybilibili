package com.mybilibili.ai.consumer;

import com.mybilibili.ai.feign.ReportCallbackClient;
import com.mybilibili.ai.service.ContentReviewService;
import com.mybilibili.mq.ContentReviewMessage;
import com.mybilibili.mq.MQConstants;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.HashMap;
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
    private ReportCallbackClient reportCallbackClient;

    @Override
    public void onMessage(ContentReviewMessage message) {
        Map<String, Object> reviewResult = contentReviewService.reviewContent(
            message.getContent(), message.getReason());

        String status = reviewResult.get("status") != null ? reviewResult.get("status").toString() : "FAILED";

        Map<String, Object> callbackData = new HashMap<>();
        callbackData.put("reportId", message.getReportId());
        callbackData.put("aiReviewStatus", status);
        callbackData.put("aiVerdict", reviewResult.get("verdict"));
        callbackData.put("aiRiskLevel", reviewResult.get("riskLevel"));

        reportCallbackClient.updateAiReviewResult(callbackData);
    }
}