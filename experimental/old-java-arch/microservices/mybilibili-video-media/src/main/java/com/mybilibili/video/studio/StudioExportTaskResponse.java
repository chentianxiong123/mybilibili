package com.mybilibili.video.studio;

import lombok.Data;

@Data
public class StudioExportTaskResponse {
    private String taskId;
    private String projectId;
    private String projectName;
    private String status;
    private Integer progress;
    private String stage;
    private String message;
    private String errorMessage;
    private String outputUrl;
    private Long createdAt;
    private Long updatedAt;

    public static StudioExportTaskResponse fromRecord(StudioExportTaskRecord record) {
        StudioExportTaskResponse response = new StudioExportTaskResponse();
        response.setTaskId(record.getTaskId());
        response.setProjectId(record.getProjectId());
        response.setProjectName(record.getProjectName());
        response.setStatus(record.getStatus());
        response.setProgress(record.getProgress());
        response.setStage(record.getStage());
        response.setMessage(record.getMessage());
        response.setErrorMessage(record.getErrorMessage());
        response.setOutputUrl(record.getOutputUrl());
        response.setCreatedAt(record.getCreatedAt());
        response.setUpdatedAt(record.getUpdatedAt());
        return response;
    }
}
