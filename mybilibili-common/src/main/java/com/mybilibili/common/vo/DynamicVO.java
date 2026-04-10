package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class DynamicVO {
    private Integer id;
    private Integer userId;
    private String content;
    private Integer dynamicType;
    private List<String> imageUrls;
    private Integer refManuscriptId;
    private Integer likeCount;
    private Integer commentCount;
    private Integer shareCount;
    private Date createdAt;
    private Integer status;
    private Boolean isLiked;
    private UserVO user;
}
