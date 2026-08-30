package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

@Data
public class UserExperienceEvent implements Serializable {
    private static final long serialVersionUID = 1L;

    private Integer userId;
    private Integer experienceAmount;
    private String sourceType;
    private Integer sourceId;
    private String description;

    public static UserExperienceEvent of(Integer userId, Integer experienceAmount,
                                         String sourceType, Integer sourceId, String description) {
        UserExperienceEvent event = new UserExperienceEvent();
        event.setUserId(userId);
        event.setExperienceAmount(experienceAmount);
        event.setSourceType(sourceType);
        event.setSourceId(sourceId);
        event.setDescription(description);
        return event;
    }
}
