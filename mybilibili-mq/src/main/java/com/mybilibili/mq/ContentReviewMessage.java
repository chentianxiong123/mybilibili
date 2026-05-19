package com.mybilibili.mq;

import lombok.Data;
import java.io.Serializable;

@Data
public class ContentReviewMessage implements Serializable {
    private static final long serialVersionUID = 1L;

    private Integer reportId;
    private String targetType;
    private Integer targetId;
    private Integer manuscriptId;
    private String content;
    private String reason;

    public static ContentReviewMessage of(Integer reportId, String targetType, Integer targetId,
                                           Integer manuscriptId, String content, String reason) {
        ContentReviewMessage msg = new ContentReviewMessage();
        msg.setReportId(reportId);
        msg.setTargetType(targetType);
        msg.setTargetId(targetId);
        msg.setManuscriptId(manuscriptId);
        msg.setContent(content);
        msg.setReason(reason);
        return msg;
    }
}