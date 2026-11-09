package com.mybilibili.ai.process;

public class StepExecutionResult {

    private final boolean success;
    private final String errorMessage;
    private final Object data;

    private StepExecutionResult(boolean success, String errorMessage, Object data) {
        this.success = success;
        this.errorMessage = errorMessage;
        this.data = data;
    }

    public static StepExecutionResult success() {
        return new StepExecutionResult(true, null, null);
    }

    public static StepExecutionResult success(Object data) {
        return new StepExecutionResult(true, null, data);
    }

    public static StepExecutionResult fail(String errorMessage) {
        return new StepExecutionResult(false, errorMessage, null);
    }

    public boolean isSuccess() {
        return success;
    }

    public String getErrorMessage() {
        return errorMessage;
    }

    public Object getData() {
        return data;
    }
}
