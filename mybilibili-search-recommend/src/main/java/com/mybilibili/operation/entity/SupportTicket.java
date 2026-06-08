package com.mybilibili.operation.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.util.Date;

@Data
@TableName("support_tickets")
public class SupportTicket {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String ticketNo;
    private Long userId;
    private Long sessionId;
    private String source;
    private String category;
    private String priority;
    private String status;
    private String title;
    private String content;
    private String entryReply;
    private String adminReply;
    private Long assigneeAdminId;
    private Date processedAt;
    private Date createdAt;
    private Date updatedAt;
}
