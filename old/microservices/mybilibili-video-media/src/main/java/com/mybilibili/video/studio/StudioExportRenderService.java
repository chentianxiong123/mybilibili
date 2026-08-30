package com.mybilibili.video.studio;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.storage.StorageService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.LinkedHashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

@Service
@Slf4j
public class StudioExportRenderService {
    private final ObjectMapper objectMapper;
    private final StorageService storageService;
    private final Path workRoot;

    public StudioExportRenderService(ObjectMapper objectMapper,
                                     StorageService storageService,
                                     @Value("${studio.export.work-dir:${video.processing.work-dir:${upload.base-path:mybilibili-processing}}/studio-export}") String workDir) {
        this.objectMapper = objectMapper;
        this.storageService = storageService;
        this.workRoot = Path.of(workDir);
    }

    public RenderPreparation prepare(StudioExportTaskRecord task) {
        if (task == null || !StringUtils.hasText(task.getTaskId())) {
            throw new BusinessException("导出任务无效");
        }

        try {
            JsonNode project = objectMapper.readTree(task.getProjectJson());
            JsonNode exportSettings = objectMapper.readTree(task.getExportSettingsJson());
            Map<String, JsonNode> mediaById = indexMedia(project);
            Set<String> referencedMediaIds = collectReferencedMediaIds(project);
            List<PreparedAsset> preparedAssets = new ArrayList<>();

            Path taskDir = workRoot.resolve(task.getTaskId()).normalize();
            Path assetDir = taskDir.resolve("assets");
            Files.createDirectories(assetDir);

            for (String mediaId : referencedMediaIds) {
                JsonNode media = mediaById.get(mediaId);
                if (media == null) {
                    throw new BusinessException("时间线引用了不存在的素材: " + mediaId);
                }

                String objectKey = text(media, "cloudObjectKey");
                if (!StringUtils.hasText(objectKey)) {
                    throw new BusinessException("素材尚未上传到云端: " + safeName(media, mediaId));
                }

                Path localPath = assetDir.resolve(mediaId + extension(safeName(media, mediaId))).normalize();
                if (!localPath.startsWith(assetDir)) {
                    throw new BusinessException("素材路径无效: " + mediaId);
                }

                try (InputStream input = storageService.download(objectKey)) {
                    Files.copy(input, localPath, StandardCopyOption.REPLACE_EXISTING);
                }

                preparedAssets.add(new PreparedAsset(
                        mediaId,
                        safeName(media, mediaId),
                        text(media, "type"),
                        objectKey,
                        text(media, "cloudUrl"),
                        localPath.toString(),
                        media.path("metadata").path("duration").asDouble(0),
                        media.path("metadata").path("width").asInt(0),
                        media.path("metadata").path("height").asInt(0),
                        media.path("metadata").path("fileSize").asLong(0)
                ));
            }

            Path manifestPath = taskDir.resolve("render-manifest.json");
            Map<String, Object> manifest = new LinkedHashMap<>();
            manifest.put("taskId", task.getTaskId());
            manifest.put("userId", task.getUserId());
            manifest.put("projectId", task.getProjectId());
            manifest.put("projectName", task.getProjectName());
            manifest.put("exportSettings", exportSettings);
            manifest.put("assets", preparedAssets);
            manifest.put("timeline", project.path("timeline"));
            manifest.put("textClips", project.path("textClips"));
            manifest.put("shapeClips", project.path("shapeClips"));
            manifest.put("createdAt", System.currentTimeMillis());
            objectMapper.writerWithDefaultPrettyPrinter().writeValue(manifestPath.toFile(), manifest);

            log.info("剪辑云端导出准备完成: taskId={}, assets={}, manifest={}", task.getTaskId(), preparedAssets.size(), manifestPath);
            return new RenderPreparation(taskDir.toString(), manifestPath.toString(), preparedAssets.size());
        } catch (BusinessException e) {
            throw e;
        } catch (Exception e) {
            throw new BusinessException("渲染准备失败: " + e.getMessage());
        }
    }

    private Map<String, JsonNode> indexMedia(JsonNode project) {
        Map<String, JsonNode> mediaById = new LinkedHashMap<>();
        JsonNode items = project.path("mediaLibrary").path("items");
        if (!items.isArray()) {
            return mediaById;
        }
        for (JsonNode item : items) {
            String id = text(item, "id");
            if (StringUtils.hasText(id)) {
                mediaById.put(id, item);
            }
        }
        return mediaById;
    }

    private Set<String> collectReferencedMediaIds(JsonNode project) {
        Set<String> ids = new LinkedHashSet<>();
        JsonNode tracks = project.path("timeline").path("tracks");
        if (!tracks.isArray()) {
            return ids;
        }
        for (JsonNode track : tracks) {
            JsonNode clips = track.path("clips");
            if (!clips.isArray()) {
                continue;
            }
            for (JsonNode clip : clips) {
                String mediaId = text(clip, "mediaId");
                if (StringUtils.hasText(mediaId)) {
                    ids.add(mediaId);
                }
            }
        }
        return ids;
    }

    private String text(JsonNode node, String field) {
        JsonNode value = node == null ? null : node.get(field);
        return value != null && !value.isNull() ? value.asText() : null;
    }

    private String safeName(JsonNode media, String fallback) {
        String name = text(media, "name");
        if (!StringUtils.hasText(name)) {
            return fallback;
        }
        return name.replace("\\", "/").replaceAll(".*/", "").replaceAll("[^A-Za-z0-9._-]", "_");
    }

    private String extension(String fileName) {
        if (!StringUtils.hasText(fileName)) {
            return ".bin";
        }
        int index = fileName.lastIndexOf('.');
        if (index < 0 || index == fileName.length() - 1) {
            return ".bin";
        }
        String ext = fileName.substring(index).toLowerCase();
        return ext.matches("\\.[a-z0-9]{1,12}") ? ext : ".bin";
    }

    public record RenderPreparation(String workDir, String manifestPath, int assetCount) {
    }

    public record PreparedAsset(String mediaId,
                                String name,
                                String type,
                                String objectKey,
                                String cloudUrl,
                                String localPath,
                                double duration,
                                int width,
                                int height,
                                long fileSize) {
    }
}
