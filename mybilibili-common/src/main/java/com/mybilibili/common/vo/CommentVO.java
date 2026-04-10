package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class CommentVO {
    private Integer id;
    private Integer videoId;
    private Integer userId;
    private String userName;
    private String userAvatar;
    private String content;
    private Integer likeCount;
    private Integer replyCount;
    private Date createTime;
    private List<ReplyVO> replies;
    private boolean liked;

    private Integer manuscriptId;

    // 违禁词检测相关字段
    private Boolean hasProhibitedWords;  // 是否包含违禁词
    private List<String> prohibitedWords; // 检测到的违禁词列表
}
