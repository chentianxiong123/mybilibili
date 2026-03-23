package com.mybilibili.admin.dto;

import lombok.Data;
import java.util.Date;

@Data
public class BannerImageDTO {
    private String title;
    private String imageUrl;
    private String linkUrl;
    private Integer sortOrder;
    private Integer status;
    private Date startTime;
    private Date endTime;
}
