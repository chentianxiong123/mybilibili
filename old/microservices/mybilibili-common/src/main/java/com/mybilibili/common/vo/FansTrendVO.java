package com.mybilibili.common.vo;

import lombok.Data;

import java.util.List;

@Data
public class FansTrendVO {
    private List<String> dates;
    private List<Integer> newFollowers;
    private List<Integer> unfollows;
    private List<Integer> totalFollowers;
    private Integer currentFollowers;
    private Integer newFollowersToday;
    private Integer unfollowsToday;
    private Double growthRate;
}
