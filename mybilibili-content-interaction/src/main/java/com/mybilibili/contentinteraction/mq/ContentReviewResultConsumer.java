package com.mybilibili.contentinteraction.mq;

import com.mybilibili.comment.service.ReportService;
import com.mybilibili.mq.ContentReviewResultEvent;
import com.mybilibili.mq.MQConstants;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_CONTENT_REVIEW_RESULT,
        consumerGroup = MQConstants.GROUP_CONTENT_REVIEW_RESULT
)
public class ContentReviewResultConsumer implements RocketMQListener<ContentReviewResultEvent> {

    private final ReportService reportService;

    public ContentReviewResultConsumer(ReportService reportService) {
        this.reportService = reportService;
    }

    @Override
    public void onMessage(ContentReviewResultEvent event) {
        if (event == null || event.getReportId() == null) {
            log.warn("忽略无效内容审核结果事件: {}", event);
            return;
        }
        Map<String, Object> result = new HashMap<>();
        result.put("reportId", event.getReportId());
        result.put("aiReviewStatus", event.getAiReviewStatus());
        result.put("aiVerdict", event.getAiVerdict());
        result.put("aiRiskLevel", event.getAiRiskLevel());
        reportService.updateAiReviewResult(result);
    }
}
