package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

@Data
public class ManuscriptCommentCountEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    private Integer manuscriptId;
    private Integer delta;

    public static ManuscriptCommentCountEvent of(Integer manuscriptId, Integer delta) {
        ManuscriptCommentCountEvent event = new ManuscriptCommentCountEvent();
        event.setManuscriptId(manuscriptId);
        event.setDelta(delta);
        return event;
    }
}
