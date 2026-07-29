package com.mybilibili.common.controller;

import com.mybilibili.common.audit.AuditLogRecorder;
import com.mybilibili.common.operation.OperationTaskRecorder;
import com.mybilibili.common.storage.StorageMigrationTool;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.nio.file.Path;

@RestController
@RequestMapping("/admin/storage")
@Tag(name = "存储管理", description = "文件存储迁移管理接口")
@ConditionalOnBean(StorageService.class)
@ConditionalOnProperty(name = "storage.migration.local-path")
public class StorageMigrationController {

    @Autowired
    private StorageService storageService;

    @Autowired
    private AuditLogRecorder auditLogRecorder;

    @Autowired
    private OperationTaskRecorder operationTaskRecorder;

    @Value("${storage.migration.local-path}")
    private String localPath;

    @PostMapping("/migrate")
    @Operation(summary = "迁移本地文件到MinIO", description = "将本地uploads目录下的文件迁移到MinIO存储")
    public Result<StorageMigrationTool.MigrationResult> migrateToMinio(HttpServletRequest request) {
        String taskKey = "storage-migration:minio:" + System.currentTimeMillis();
        operationTaskRecorder.running(request, taskKey, "STORAGE_MIGRATION", "本地文件迁移到MinIO",
                "storage", "minio", 0, "MIGRATING", "存储迁移已开始");
        try {
            StorageMigrationTool tool = new StorageMigrationTool(storageService, Path.of(localPath));
            StorageMigrationTool.MigrationResult result = tool.migrateAll();
            Result<StorageMigrationTool.MigrationResult> response = Result.success("迁移完成", result);
            operationTaskRecorder.success(request, taskKey, "STORAGE_MIGRATION", "本地文件迁移到MinIO",
                    "storage", "minio", "COMPLETED", response.getMessage());
            auditLogRecorder.record(request, "storage", "storage_migration_trigger", "storage", "minio", response);
            return response;
        } catch (Exception e) {
            Result<StorageMigrationTool.MigrationResult> response = Result.error("迁移失败: " + e.getMessage());
            operationTaskRecorder.failed(request, taskKey, "STORAGE_MIGRATION", "本地文件迁移到MinIO",
                    "storage", "minio", 0, "MIGRATING", response.getMessage(), e.getMessage());
            auditLogRecorder.record(request, "storage", "storage_migration_trigger", "storage", "minio", response);
            return response;
        }
    }
}
