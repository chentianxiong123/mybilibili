package com.mybilibili.user.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("login_logs")
public class LoginLog {
    @TableId(type = IdType.AUTO)
    private Integer id;
    private Integer userId;
    private String ip;
    private String userAgent;
    private String location;
    private LocalDateTime loginTime;
    private Integer status; // 1成功 0失败
}