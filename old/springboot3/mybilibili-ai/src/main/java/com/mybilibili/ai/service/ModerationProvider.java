package com.mybilibili.ai.service;

/**
 * 内容审核提供者接口（Moderation）
 * 用于检测文本内容是否违规（政治、色情、广告、暴力等）
 */
public interface ModerationProvider extends AiServiceProvider<ModerationProvider.ModerateRequest> {

    /**
     * 审核请求
     */
    class ModerateRequest {
        /** 待审核文本 */
        private String text;
        /** 审核场景：COMMENT / REPLY / MANUSCRIPT */
        private String scene;

        public String getText() { return text; }
        public void setText(String text) { this.text = text; }
        public String getScene() { return scene; }
        public void setScene(String scene) { this.scene = scene; }

        public static ModerateRequest of(String text) {
            ModerateRequest req = new ModerateRequest();
            req.text = text;
            return req;
        }

        public static ModerateRequest of(String text, String scene) {
            ModerateRequest req = new ModerateRequest();
            req.text = text;
            req.scene = scene;
            return req;
        }
    }

    /**
     * 审核结果
     */
    class ModerationResult {
        /** 是否通过 */
        private boolean passed;
        /** 违规类型列表 */
        private java.util.List<String> violationTypes;
        /** 详细描述 */
        private String reason;

        public ModerationResult() {}

        public ModerationResult(boolean passed, String reason) {
            this.passed = passed;
            this.reason = reason;
        }

        public boolean isPassed() { return passed; }
        public void setPassed(boolean passed) { this.passed = passed; }
        public java.util.List<String> getViolationTypes() { return violationTypes; }
        public void setViolationTypes(java.util.List<String> violationTypes) { this.violationTypes = violationTypes; }
        public String getReason() { return reason; }
        public void setReason(String reason) { this.reason = reason; }
    }

    @Override
    default String getType() { return "MODERATION"; }
}