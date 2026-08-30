package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class Like {
    private Integer id;
    private Integer userId;
    private Integer manuscriptId;  // 改为稿件ID
    private Date createdAt;
}
