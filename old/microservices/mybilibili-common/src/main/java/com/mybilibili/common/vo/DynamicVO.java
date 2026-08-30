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
    private RefVideo refVideo;
    private Integer likeCount;
    private Integer commentCount;
    private Integer shareCount;
    private Date createdAt;
    private Integer status;
    private Boolean isLiked;
    private UserVO user;

    @Data
    public static class RefVideo {
        private Integer id;
        private String title;
        private String cover;
        private String duration;
        private Integer viewCount;
    }
}
