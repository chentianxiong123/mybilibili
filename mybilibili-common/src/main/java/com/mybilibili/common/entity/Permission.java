package com.mybilibili.common.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.util.Date;

@Data
@TableName("permissions")
public class Permission {
    @TableId(type = IdType.AUTO)
    private Integer id;
    private String name;
    private String code;
    private String url;
    private String method;
    private Integer parentId;
    private String description;
    private Date createTime;
    private Date updateTime;
}
