package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("ai_bindings")
public class AiBinding {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String feature;
    private Long apiConfigId;
}
