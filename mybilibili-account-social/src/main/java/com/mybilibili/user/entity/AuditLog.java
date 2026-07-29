package com.mybilibili.user.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("audit_logs")
public class AuditLog {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Integer operatorId;
    private String operatorName;
    private String operatorRole;
    private String module;
    private String action;
    private String targetType;
    private String targetId;
    private String requestMethod;
    private String requestUri;
    private String clientIp;
    private String userAgent;
    private Integer result;
    private String message;
    private String detail;
    private LocalDateTime createdAt;
}
