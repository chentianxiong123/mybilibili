package com.mybilibili.ai.service;

public interface AiSummaryService {

    String generateSummary(String subtitleText, String videoTitle, String videoDescription);

    String generateSummary(String subtitleText, String videoTitle);

    TestResult testApiConnection(String testText);

    boolean saveSummaryToFile(String summary, String filePath, String videoTitle);

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