package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("favorite_folders")
public class FavoriteFolder {
    private Integer id;
    private Integer userId;
    private String name;
    private Integer collectCount;
    private Date createdAt;
    private Date updatedAt;
}