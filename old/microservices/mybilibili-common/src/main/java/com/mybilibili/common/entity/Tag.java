package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("tags")
public class Tag {
    private Integer id;
    private String name;
    private String description;
    private Date createdAt;
    private Date updatedAt;
}