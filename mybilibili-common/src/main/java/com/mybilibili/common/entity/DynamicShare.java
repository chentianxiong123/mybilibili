package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

/**
 * 动态转发关系实体
 */
@Data
public class DynamicShare {
    
    private Integer id;
    
    /**
     * 被转发的动态ID（原动态）
     */
    private Integer dynamicId;
    
    /**
     * 转发用户ID
     */
    private Integer userId;
    
    /**
     * 转发时附加的评论内容
     */
    private String content;
    
    /**
     * 创建时间
     */
    private Date createdAt;
    
    /**
     * 状态：0-正常，1-已删除
     */
    private Integer status;
}
