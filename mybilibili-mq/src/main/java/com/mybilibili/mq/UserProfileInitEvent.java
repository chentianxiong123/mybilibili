package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;
import java.util.List;

@Data
public class UserProfileInitEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    private Integer userId;
    private List<String> tags;

    public static UserProfileInitEvent of(Integer userId, List<String> tags) {
        UserProfileInitEvent event = new UserProfileInitEvent();
        event.setUserId(userId);
        event.setTags(tags);
        return event;
    }
}
