package com.mybilibili.ai.service;

/**
 * AI 服务提供者接口（基类）。
 * 按类型分类：LLM（对话）、ASR（语音转文字）、TTS（文字转语音）等。
 */
public interface AiServiceProvider<C> {

    /**
     * 提供者类型标识 (LLM / ASR / TTS / IMAGE)
     */
    String getType();

    /**
     * 提供者名称（渠道名）
     */
    String getName();

    /**
     * 是否可用
     */
    boolean isAvailable();

    /**
     * 执行调用
     * @param config 配置
     * @return 调用结果
     */
    Object invoke(C config);
}
