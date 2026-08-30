package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class Tag {
    private Integer id;
    private String name;
    private String description;
    private Date createdAt;
    private Date updatedAt;
}