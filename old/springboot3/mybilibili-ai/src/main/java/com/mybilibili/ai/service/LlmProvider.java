package com.mybilibili.ai.service;

/**
 * LLM 提供者接口（对话补全）
 */
public interface LlmProvider extends AiServiceProvider<LlmProvider.ChatRequest> {

    /**
     * 对话请求
     */
    class ChatRequest {
        private String prompt;
        private String systemPrompt;
        private Integer maxTokens;
        private Double temperature;

        public String getPrompt() { return prompt; }
        public void setPrompt(String prompt) { this.prompt = prompt; }
        public String getSystemPrompt() { return systemPrompt; }
        public void setSystemPrompt(String systemPrompt) { this.systemPrompt = systemPrompt; }
        public Integer getMaxTokens() { return maxTokens; }
        public void setMaxTokens(Integer maxTokens) { this.maxTokens = maxTokens; }
        public Double getTemperature() { return temperature; }
        public void setTemperature(Double temperature) { this.temperature = temperature; }

        public static ChatRequest of(String prompt) {
            ChatRequest req = new ChatRequest();
            req.prompt = prompt;
            return req;
        }

        public static ChatRequest of(String prompt, String systemPrompt) {
            ChatRequest req = new ChatRequest();
            req.prompt = prompt;
            req.systemPrompt = systemPrompt;
            return req;
        }
    }

    @Override
    default String getType() { return "LLM"; }
}