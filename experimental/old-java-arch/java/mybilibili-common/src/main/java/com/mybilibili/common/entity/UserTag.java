package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class UserTag {
    private Integer id;
    private Integer userId;
    private String tagName;
    private Date createdAt;
}
