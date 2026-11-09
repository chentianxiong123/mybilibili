package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;
import java.time.LocalDateTime;

@Data
public class ManuscriptAnalyticsEvent implements Serializable {

    private static final long serialVersionUID = 1L;

    public static final String TYPE_VIEW_INCREMENT = "VIEW_INCREMENT";
    public static final String TYPE_STATUS_CHANGE = "STATUS_CHANGE";
    public static final String TYPE_METRIC_INCREMENT = "METRIC_INCREMENT";

    public static final String METRIC_LIKE = "LIKE";
    public static final String METRIC_COIN = "COIN";
    public static final String METRIC_COLLECT = "COLLECT";
    public static final String METRIC_SHARE = "SHARE";
    public static final String METRIC_COMMENT = "COMMENT";
    public static final String METRIC_DANMAKU = "DANMAKU";

    private String eventType;
    private Integer manuscriptId;
    private Integer userId;
    private String metricType;
    private Integer delta;
    private Integer fromStatus;
    private Integer toStatus;
    private String action;
    private String operatorType;
    private Integer operatorId;
    private String reason;
    private LocalDateTime occurredAt;

    public static ManuscriptAnalyticsEvent viewIncrement(Integer manuscriptId, Integer userId) {
        ManuscriptAnalyticsEvent event = new ManuscriptAnalyticsEvent();
        event.setEventType(TYPE_VIEW_INCREMENT);
        event.setManuscriptId(manuscriptId);
        event.setUserId(userId);
        event.setOccurredAt(LocalDateTime.now());
        return event;
    }

    public static ManuscriptAnalyticsEvent statusChange(Integer manuscriptId,
                                                        Integer userId,
                                                        Integer fromStatus,
                                                        Integer toStatus,
                                                        String action,
                                                        String operatorType,
                                                        Integer operatorId,
                                                        String reason) {
        ManuscriptAnalyticsEvent event = new ManuscriptAnalyticsEvent();
        event.setEventType(TYPE_STATUS_CHANGE);
        event.setManuscriptId(manuscriptId);
        event.setUserId(userId);
        event.setFromStatus(fromStatus);
        event.setToStatus(toStatus);
        event.setAction(action);
        event.setOperatorType(operatorType);
        event.setOperatorId(operatorId);
        event.setReason(reason);
        event.setOccurredAt(LocalDateTime.now());
        return event;
    }

    public static ManuscriptAnalyticsEvent metricIncrement(Integer manuscriptId,
                                                           Integer userId,
                                                           String metricType,
                                                           Integer delta) {
        ManuscriptAnalyticsEvent event = new ManuscriptAnalyticsEvent();
        event.setEventType(TYPE_METRIC_INCREMENT);
        event.setManuscriptId(manuscriptId);
        event.setUserId(userId);
        event.setMetricType(metricType);
        event.setDelta(delta);
        event.setOccurredAt(LocalDateTime.now());
        return event;
    }
}
