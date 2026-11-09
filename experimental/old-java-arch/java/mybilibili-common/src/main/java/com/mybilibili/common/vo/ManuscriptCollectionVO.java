package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class ManuscriptCollectionVO {
    private Integer id;
    private String title;
    private String description;
    private String coverUrl;
    private Integer userId;
    private Integer manuscriptCount;
    private Integer viewCount;
    private Integer status;
    private Date createdAt;
    private Date updatedAt;

    // 关联数据
    private UserInfo creator;                    // 创建者信息
    private List<ManuscriptItemVO> manuscripts;  // 合集中的稿件列表

    @Data
    public static class UserInfo {
        private Integer id;
        private String name;
        private String avatar;
        private Integer level;
    }

    @Data
    public static class ManuscriptItemVO {
        private Integer id;
        private String title;
        private String description;
        private String coverUrl;
        private Integer viewCount;
        private Integer likeCount;
        private Integer commentCount;
        private Integer collectionOrder;  // 在合集中的顺序
        private Date uploadTime;
        private String duration;          // 视频时长
    }
}
