package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

@Data
public class UserNotificationEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    public static final String TYPE_VIDEO_LIKE = "VIDEO_LIKE";
    public static final String TYPE_COMMENT_LIKE = "COMMENT_LIKE";
    public static final String TYPE_REPLY = "REPLY";
    public static final String TYPE_SYSTEM = "SYSTEM";

    private String type;
    private Integer senderId;
    private Integer receiverId;
    private String title;
    private String content;
    private Integer messageType;
    private Integer targetId;
    private Integer commentId;

    public static UserNotificationEvent videoLike(Integer senderId, Integer receiverId,
                                                  Integer videoId, String videoTitle) {
        UserNotificationEvent event = new UserNotificationEvent();
        event.setType(TYPE_VIDEO_LIKE);
        event.setSenderId(senderId);
        event.setReceiverId(receiverId);
        event.setTargetId(videoId);
        event.setTitle(videoTitle);
        return event;
    }

    public static UserNotificationEvent commentLike(Integer senderId, Integer receiverId,
                                                    Integer commentId, String commentContent) {
        UserNotificationEvent event = new UserNotificationEvent();
        event.setType(TYPE_COMMENT_LIKE);
        event.setSenderId(senderId);
        event.setReceiverId(receiverId);
        event.setCommentId(commentId);
        event.setContent(commentContent);
        return event;
    }

    public static UserNotificationEvent reply(Integer senderId, Integer receiverId, String content,
                                              Integer messageType, Integer targetId, Integer commentId) {
        UserNotificationEvent event = new UserNotificationEvent();
        event.setType(TYPE_REPLY);
        event.setSenderId(senderId);
        event.setReceiverId(receiverId);
        event.setContent(content);
        event.setMessageType(messageType);
        event.setTargetId(targetId);
        event.setCommentId(commentId);
        return event;
    }

    public static UserNotificationEvent system(Integer userId, String title, String content) {
        UserNotificationEvent event = new UserNotificationEvent();
        event.setType(TYPE_SYSTEM);
        event.setReceiverId(userId);
        event.setTitle(title);
        event.setContent(content);
        return event;
    }
}
