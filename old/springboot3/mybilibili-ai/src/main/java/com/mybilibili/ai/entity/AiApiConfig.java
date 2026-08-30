package com.mybilibili.ai.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.util.Date;

/**
 * AI API 配置实体。
 * 每个渠道代表一个 API 提供者，按 type 分类。
 *
 * type 可选值:
 * - LLM: 大模型对话（chat/completion）
 * - MODERATION: 内容审核（预过滤/细审核）
 * - ASR: 语音转文字（speech-to-text）
 * - TTS: 文字转语音（text-to-speech）
 * - IMAGE: 图像生成
 */
@Data
@TableName("ai_api_configs")
public class AiApiConfig {
    @TableId(type = IdType.AUTO)
    private Long id;

    /** 渠道名称 */
    private String name;

    /** API 类型: LLM / STT / TTS / IMAGE */
    private String type;

    /** API 根地址 */
    private String baseUrl;

    /** API 密钥 */
    private String apiKey;

    /** 模型名称 */
    private String model;

    /** 最大 token 数 */
    private Integer maxTokens;

    /** 温度参数 */
    private Double temperature;

    /** 是否启用 */
    private Boolean enabled;

    /** 扩展配置（JSON 格式，存储 type 特定配置） */
    private String extraConfig;

    private Date createdAt;
    private Date updatedAt;
}