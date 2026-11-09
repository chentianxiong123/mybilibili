package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

@Data
public class ContentReviewResultEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    private Integer reportId;
    private String aiReviewStatus;
    private String aiVerdict;
    private String aiRiskLevel;

    public static ContentReviewResultEvent of(Integer reportId, String aiReviewStatus,
                                              Object aiVerdict, Object aiRiskLevel) {
        ContentReviewResultEvent event = new ContentReviewResultEvent();
        event.setReportId(reportId);
        event.setAiReviewStatus(aiReviewStatus);
        event.setAiVerdict(aiVerdict == null ? null : aiVerdict.toString());
        event.setAiRiskLevel(aiRiskLevel == null ? null : aiRiskLevel.toString());
        return event;
    }
}
