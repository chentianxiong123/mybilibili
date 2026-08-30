package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;
import java.util.List;

@Data
@TableName("manuscript_collections")
public class ManuscriptCollection {
    @TableId(type = IdType.AUTO)
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
    
    // 非数据库字段
    private List<Manuscript> manuscripts; // 合集中的稿件列表
    private User user;
    
    // 状态常量
    public static final int STATUS_PRIVATE = 0;  // 私密
    public static final int STATUS_PUBLIC = 1;   // 公开
}
