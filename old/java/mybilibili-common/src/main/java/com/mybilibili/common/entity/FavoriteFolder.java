package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class FavoriteFolder {
    private Integer id;
    private Integer userId;
    private String name;
    private Integer videoCount;
    private Date createdAt;
    private Date updatedAt;
}