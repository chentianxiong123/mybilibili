package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class FavoriteVideo {
    private Integer id;
    private Integer folderId;
    private Integer manuscriptId;  // 改为稿件ID
    private Date createdAt;
}
