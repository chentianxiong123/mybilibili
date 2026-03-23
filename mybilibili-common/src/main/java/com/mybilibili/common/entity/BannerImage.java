package com.mybilibili.common.entity;

import lombok.Data;
import java.util.Date;

@Data
public class BannerImage {
    private Integer id;
    private String title;
    private String imageUrl;
    private String linkUrl;
    private Integer sortOrder;
    private Integer status;
    private Date startTime;
    private Date endTime;

    // 状态常量
    public static final Integer STATUS_DISABLED = 0;
    public static final Integer STATUS_ENABLED = 1;
}
