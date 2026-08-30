package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.util.Date;

@Data
@TableName("admin_users")
public class AdminUser {
    private Integer id;
    private String username;
    private String password;
    private Integer adminLevel;
    private Date createdAt;
    private Date updatedAt;
}