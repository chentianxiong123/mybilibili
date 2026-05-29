package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.util.Date;

@Data
@TableName("ai_api_configs")
public class AiApiConfig {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String name;
    private String baseUrl;
    private String apiKey;
    private String model;
    private Integer maxTokens;
    private Double temperature;
    private Boolean enabled;
    private Date createdAt;
    private Date updatedAt;
}
