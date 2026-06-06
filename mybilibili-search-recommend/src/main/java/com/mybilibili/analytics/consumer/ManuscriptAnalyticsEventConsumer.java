package com.mybilibili.analytics.consumer;

import com.mybilibili.analytics.mapper.ManuscriptAnalyticsMapper;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.ManuscriptAnalyticsEvent;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;

@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_MANUSCRIPT_ANALYTICS,
        consumerGroup = MQConstants.GROUP_MANUSCRIPT_ANALYTICS
)
public class ManuscriptAnalyticsEventConsumer implements RocketMQListener<ManuscriptAnalyticsEvent> {

    private final ManuscriptAnalyticsMapper manuscriptAnalyticsMapper;

    public ManuscriptAnalyticsEventConsumer(ManuscriptAnalyticsMapper manuscriptAnalyticsMapper) {
        this.manuscriptAnalyticsMapper = manuscriptAnalyticsMapper;
    }

    @Override
    public void onMessage(ManuscriptAnalyticsEvent event) {
        if (event == null || event.getEventType() == null) {
            return;
        }
        LocalDateTime occurredAt = event.getOccurredAt() != null ? event.getOccurredAt() : LocalDateTime.now();
        switch (event.getEventType()) {
            case ManuscriptAnalyticsEvent.TYPE_VIEW_INCREMENT:
                manuscriptAnalyticsMapper.incrementDailyView(event.getManuscriptId(), event.getUserId(), occurredAt);
                break;
            case ManuscriptAnalyticsEvent.TYPE_STATUS_CHANGE:
                manuscriptAnalyticsMapper.insertStatusEvent(
                        event.getManuscriptId(),
                        event.getUserId(),
                        event.getFromStatus(),
                        event.getToStatus(),
                        event.getAction(),
                        event.getOperatorType(),
                        event.getOperatorId(),
                        event.getReason(),
                        occurredAt
                );
                break;
            case ManuscriptAnalyticsEvent.TYPE_METRIC_INCREMENT:
                incrementMetric(event, occurredAt);
                break;
            default:
                throw new IllegalArgumentException("Unsupported manuscript analytics event type: " + event.getEventType());
        }
    }

    private void incrementMetric(ManuscriptAnalyticsEvent event, LocalDateTime occurredAt) {
        Integer delta = event.getDelta() != null ? event.getDelta() : 1;
        switch (event.getMetricType()) {
            case ManuscriptAnalyticsEvent.METRIC_LIKE:
                manuscriptAnalyticsMapper.incrementDailyLike(event.getManuscriptId(), event.getUserId(), delta, occurredAt);
                break;
            case ManuscriptAnalyticsEvent.METRIC_COIN:
                manuscriptAnalyticsMapper.incrementDailyCoin(event.getManuscriptId(), event.getUserId(), delta, occurredAt);
                break;
            case ManuscriptAnalyticsEvent.METRIC_COLLECT:
                manuscriptAnalyticsMapper.incrementDailyCollect(event.getManuscriptId(), event.getUserId(), delta, occurredAt);
                break;
            case ManuscriptAnalyticsEvent.METRIC_SHARE:
                manuscriptAnalyticsMapper.incrementDailyShare(event.getManuscriptId(), event.getUserId(), delta, occurredAt);
                break;
            case ManuscriptAnalyticsEvent.METRIC_COMMENT:
                manuscriptAnalyticsMapper.incrementDailyComment(event.getManuscriptId(), event.getUserId(), delta, occurredAt);
                break;
            case ManuscriptAnalyticsEvent.METRIC_DANMAKU:
                manuscriptAnalyticsMapper.incrementDailyDanmaku(event.getManuscriptId(), event.getUserId(), delta, occurredAt);
                break;
            default:
                throw new IllegalArgumentException("Unsupported manuscript metric type: " + event.getMetricType());
        }
    }
}
