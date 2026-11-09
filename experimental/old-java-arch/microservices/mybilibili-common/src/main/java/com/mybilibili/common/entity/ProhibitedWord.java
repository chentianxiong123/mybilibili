package com.mybilibili.common.entity;

import lombok.Data;

import java.util.Date;

@Data
public class ProhibitedWord {
    private Integer id;
    private String word;
    private String matchType;
    private String category;
    private Integer isEnabled;
    private Date createdAt;
    private Date updatedAt;
}
