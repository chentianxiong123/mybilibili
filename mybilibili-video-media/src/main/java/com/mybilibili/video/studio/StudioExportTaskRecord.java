package com.mybilibili.video.studio;

import lombok.Data;

@Data
public class StudioExportTaskRecord {
    private String taskId;
    private Integer userId;
    private String projectId;
    private String projectName;
    private String status;
    private Integer progress;
    private String stage;
    private String message;
    private String errorMessage;
    private String outputUrl;
    private String projectJson;
    private String exportSettingsJson;
    private Long createdAt;
    private Long updatedAt;
}
