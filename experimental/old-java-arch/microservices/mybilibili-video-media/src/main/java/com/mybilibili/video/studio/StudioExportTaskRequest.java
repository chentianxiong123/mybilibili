package com.mybilibili.video.studio;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.Data;

import java.util.Map;

@Data
public class StudioExportTaskRequest {
    private String projectId;
    private String projectName;
    private JsonNode project;
    private Map<String, Object> exportSettings;
}
