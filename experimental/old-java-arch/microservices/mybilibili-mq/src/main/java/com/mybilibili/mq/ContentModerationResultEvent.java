package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

@Data
public class ContentModerationResultEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    private String targetType;
    private Integer targetId;
    private boolean passed;
    private String reason;

    public static ContentModerationResultEvent of(String targetType, Integer targetId, boolean passed, String reason) {
        ContentModerationResultEvent event = new ContentModerationResultEvent();
        event.setTargetType(targetType);
        event.setTargetId(targetId);
        event.setPassed(passed);
        event.setReason(reason);
        return event;
    }
}
