package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

@Data
public class ContentModerationMessage implements Serializable {
    private static final long serialVersionUID = 1L;

    public static final String TARGET_COMMENT = "COMMENT";
    public static final String TARGET_REPLY = "REPLY";
    public static final String TARGET_DYNAMIC_COMMENT = "DYNAMIC_COMMENT";

    private String targetType;
    private Integer targetId;
    private String content;

    public static ContentModerationMessage of(String targetType, Integer targetId, String content) {
        ContentModerationMessage message = new ContentModerationMessage();
        message.setTargetType(targetType);
        message.setTargetId(targetId);
        message.setContent(content);
        return message;
    }
}
