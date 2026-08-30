package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class UserPrivacySettings {
    private Integer id;
    private Integer userId;
    private Boolean publicCollection;
    private Boolean publicBirthdayTags;
    private Boolean publicCoinVideos;
    private Boolean publicLikeVideos;
    private Boolean publicFollowingList;
    private Boolean publicFollowersList;
    private Date createdAt;
    private Date updatedAt;
}
