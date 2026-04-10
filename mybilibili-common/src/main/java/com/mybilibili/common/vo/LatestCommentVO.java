package com.mybilibili.common.vo;

import lombok.Data;

@Data
public class LatestCommentVO {
    private Integer id;
    private String username;
    private String avatar;
    private String content;
    private String manuscriptTitle;
    private String time;
}
