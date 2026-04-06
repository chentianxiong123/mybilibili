package com.mybilibili.common.vo;

import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
public class ReplyVO {
    private Integer id;
    private Integer commentId;
    private Integer userId;
    private String userName;
    private String userAvatar;
    private String content;
    private Integer likeCount;
    private Date createTime;
    private String replyToUserName;
    private Boolean liked;
    private Boolean hasProhibitedWords;
    private List<String> prohibitedWords;
}
