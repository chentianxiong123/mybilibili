package com.mybilibili.common.dto;

import lombok.Data;

@Data
public class UserPrivacySettingsDTO {
    private Boolean publicCollection;
    private Boolean publicBirthdayTags;
    private Boolean publicCoinVideos;
    private Boolean publicLikeVideos;
    private Boolean publicFollowingList;
    private Boolean publicFollowersList;
}
