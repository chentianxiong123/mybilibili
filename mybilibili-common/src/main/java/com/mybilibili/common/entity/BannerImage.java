package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.io.Serializable;
import java.util.Date;

@Data
@TableName("banner_images")
public class BannerImage implements Serializable {
    private static final long serialVersionUID = 1L;

    @TableId(type = IdType.AUTO)
    private Integer id;

    private String title;

    private String imageUrl;

    private String linkUrl;

    private Integer sortOrder;

    private Integer status;

    private Date startTime;

    private Date endTime;

    private Integer type;

    private Integer categoryId;

    public static final Integer TYPE_HOME = 1;
    public static final Integer TYPE_CATEGORY = 2;
    public static final Integer TYPE_BACKGROUND = 3;
    public static final Integer TYPE_USER_PROFILE = 4;

    public static final Integer STATUS_DISABLED = 0;
    public static final Integer STATUS_ENABLED = 1;
}
