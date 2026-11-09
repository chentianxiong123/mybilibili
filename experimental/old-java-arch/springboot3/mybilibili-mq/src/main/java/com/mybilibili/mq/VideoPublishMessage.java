package com.mybilibili.mq;

import lombok.Data;
import java.io.Serializable;

@Data
public class VideoPublishMessage implements Serializable {

    private static final long serialVersionUID = 1L;

    private Integer manuscriptId;

    private Integer userId;

    private String action;

    public static final String ACTION_PUBLISH = "PUBLISH";
    public static final String ACTION_UNPUBLISH = "UNPUBLISH";

    public static VideoPublishMessage of(Integer manuscriptId, Integer userId, String action) {
        VideoPublishMessage message = new VideoPublishMessage();
        message.setManuscriptId(manuscriptId);
        message.setUserId(userId);
        message.setAction(action);
        return message;
    }
}