package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class Coin {
    private Integer id;
    private Integer userId;
    private Integer manuscriptId;  // 改为稿件ID
    private Integer coinCount;
    private Date createdAt;
}
