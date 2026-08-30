package com.mybilibili.admin.service;

/**
 * AI摘要服务接口
 * 用于调用AI大模型API生成视频摘要
 */
public interface AiSummaryService {

    /**
     * 生成视频摘要
     *
     * @param subtitleText 字幕文本内容
     * @param videoTitle   视频标题
     * @return 生成的摘要内容
     */
    String generateSummary(String subtitleText, String videoTitle);

    /**
     * 生成视频摘要（带描述）
     *
     * @param subtitleText    字幕文本内容
     * @param videoTitle      视频标题
     * @param videoDescription 视频描述
     * @return 生成的摘要内容
     */
    String generateSummary(String subtitleText, String videoTitle, String videoDescription);

    /**
     * 测试AI API连接
     *
     * @param testText 测试文本
     * @return 测试结果
     */
    TestResult testApiConnection(String testText);

    /**
     * 保存摘要到文件
     *
     * @param summary    摘要内容
     * @param filePath   文件路径
     * @param videoTitle 视频标题
     * @return 是否保存成功
     */
    boolean saveSummaryToFile(String summary, String filePath, String videoTitle);

    /**
     * 测试结果类
     */
    class TestResult {
        private boolean success;
        private String message;
        private String response;
        private long responseTime;

        public TestResult(boolean success, String message) {
            this.success = success;
            this.message = message;
        }

        public TestResult(boolean success, String message, String response, long responseTime) {
            this.success = success;
            this.message = message;
            this.response = response;
            this.responseTime = responseTime;
        }

        public boolean isSuccess() {
            return success;
        }

        public void setSuccess(boolean success) {
            this.success = success;
        }

        public String getMessage() {
            return message;
        }

        public void setMessage(String message) {
            this.message = message;
        }

        public String getResponse() {
            return response;
        }

        public void setResponse(String response) {
            this.response = response;
        }

        public long getResponseTime() {
            return responseTime;
        }

        public void setResponseTime(long responseTime) {
            this.responseTime = responseTime;
        }
    }
}
