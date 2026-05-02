package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.util.Date;

@Data
@TableName("user_tags")
public class UserTag {
    private Integer id;
    private Integer userId;
    private String tagName;
    private Date createdAt;
}
