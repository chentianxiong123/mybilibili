package com.mybilibili.common.vo;

import lombok.Data;
import java.util.List;

@Data
public class UserPrivacySettingsVO {
    private Boolean publicCollection;
    private Boolean publicBirthdayTags;
    private Boolean publicCoinVideos;
    private Boolean publicLikeVideos;
    private Boolean publicFollowingList;
    private Boolean publicFollowersList;
    private List<String> tags;
}
