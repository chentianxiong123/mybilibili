package com.mybilibili.mq;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class StudioExportTaskMessage {
    private String taskId;
    private Integer userId;
    private String projectId;
    private Long createdAt;
}
