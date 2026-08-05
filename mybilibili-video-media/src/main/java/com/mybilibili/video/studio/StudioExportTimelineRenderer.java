package com.mybilibili.video.studio;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.io.BufferedReader;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Locale;

@Service
@Slf4j
public class StudioExportTimelineRenderer {
    private static final int DEFAULT_WIDTH = 1920;
    private static final int DEFAULT_HEIGHT = 1080;
    private static final int DEFAULT_FRAME_RATE = 30;
    private static final int DEFAULT_BITRATE_KBPS = 8000;

    private final ObjectMapper objectMapper;
    private final StorageService storageService;

    public StudioExportTimelineRenderer(ObjectMapper objectMapper, StorageService storageService) {
        this.objectMapper = objectMapper;
        this.storageService = storageService;
    }

    public RenderResult render(StudioExportTaskRecord task, StudioExportRenderService.RenderPreparation preparation) {
        if (task == null || preparation == null || !StringUtils.hasText(preparation.manifestPath())) {
            throw new BusinessException("渲染任务无效");
        }

        try {
            JsonNode manifest = objectMapper.readTree(Path.of(preparation.manifestPath()).toFile());
            RenderSpec spec = buildRenderSpec(task, manifest, Path.of(preparation.workDir()));
            Files.createDirectories(spec.outputPath().getParent());

            List<String> command = buildCommand(spec);
            log.info("剪辑云端导出开始渲染: taskId={}, input={}, output={}", task.getTaskId(), spec.sourcePath(), spec.outputPath());
            int exitCode = run(command);
            if (exitCode != 0 || !Files.exists(spec.outputPath()) || Files.size(spec.outputPath()) == 0) {
                throw new BusinessException("FFmpeg 渲染失败，退出码: " + exitCode);
            }

            String objectKey = StorageKeys.studioExport(
                    task.getUserId(),
                    safeProjectId(task.getProjectId()),
                    task.getTaskId(),
                    "output.mp4"
            );
            String outputUrl;
            try (InputStream input = Files.newInputStream(spec.outputPath())) {
                outputUrl = storageService.upload(objectKey, input, Files.size(spec.outputPath()), "video/mp4");
            }

            log.info("剪辑云端导出渲染完成: taskId={}, objectKey={}, outputUrl={}", task.getTaskId(), objectKey, outputUrl);
            return new RenderResult(spec.outputPath().toString(), objectKey, outputUrl);
        } catch (BusinessException e) {
            throw e;
        } catch (Exception e) {
            throw new BusinessException("时间线渲染失败: " + e.getMessage());
        }
    }

    private RenderSpec buildRenderSpec(StudioExportTaskRecord task, JsonNode manifest, Path workDir) {
        JsonNode exportSettings = manifest.path("exportSettings");
        int width = positiveInt(exportSettings, "width", DEFAULT_WIDTH);
        int height = positiveInt(exportSettings, "height", DEFAULT_HEIGHT);
        int frameRate = positiveInt(exportSettings, "frameRate", DEFAULT_FRAME_RATE);
        int bitrate = positiveInt(exportSettings, "bitrate", DEFAULT_BITRATE_KBPS);
        JsonNode sourceClip = selectPrimaryClip(manifest.path("timeline").path("tracks"));
        JsonNode sourceAsset = findAsset(manifest.path("assets"), text(sourceClip, "mediaId"));
        Path sourcePath = Path.of(text(sourceAsset, "localPath"));
        if (!Files.exists(sourcePath)) {
            throw new BusinessException("渲染素材不存在: " + sourcePath);
        }

        double duration = positiveDouble(sourceClip, "duration", sourceAsset.path("duration").asDouble(0));
        if (duration <= 0) {
            duration = positiveDouble(manifest.path("timeline"), "duration", 0);
        }
        if (duration <= 0) {
            throw new BusinessException("时间线时长无效");
        }

        double inPoint = Math.max(0, sourceClip.path("inPoint").asDouble(0));
        Path outputPath = workDir.resolve("output").resolve("output.mp4").normalize();
        return new RenderSpec(sourcePath, text(sourceAsset, "type"), outputPath, width, height, frameRate, bitrate, duration, inPoint, manifest.path("textClips"));
    }

    private JsonNode selectPrimaryClip(JsonNode tracks) {
        if (!tracks.isArray()) {
            throw new BusinessException("时间线没有可渲染轨道");
        }

        List<JsonNode> clips = new ArrayList<>();
        for (JsonNode track : tracks) {
            if (track.path("hidden").asBoolean(false) || track.path("muted").asBoolean(false)) {
                continue;
            }
            String type = text(track, "type");
            if (!"video".equals(type) && !"image".equals(type)) {
                continue;
            }
            JsonNode trackClips = track.path("clips");
            if (!trackClips.isArray()) {
                continue;
            }
            for (JsonNode clip : trackClips) {
                if (StringUtils.hasText(text(clip, "mediaId"))) {
                    clips.add(clip);
                }
            }
        }
        return clips.stream()
                .min(Comparator.comparingDouble(clip -> clip.path("startTime").asDouble(0)))
                .orElseThrow(() -> new BusinessException("时间线没有视频或图片素材"));
    }

    private JsonNode findAsset(JsonNode assets, String mediaId) {
        if (!StringUtils.hasText(mediaId) || !assets.isArray()) {
            throw new BusinessException("渲染素材索引无效");
        }
        for (JsonNode asset : assets) {
            if (mediaId.equals(text(asset, "mediaId"))) {
                return asset;
            }
        }
        throw new BusinessException("渲染素材不存在: " + mediaId);
    }

    private List<String> buildCommand(RenderSpec spec) {
        List<String> command = new ArrayList<>();
        command.add("ffmpeg");
        command.add("-y");
        if (spec.isImageSource()) {
            command.add("-loop");
            command.add("1");
        } else {
            command.add("-ss");
            command.add(formatSeconds(spec.inPoint()));
        }
        command.add("-i");
        command.add(spec.sourcePath().toString());
        command.add("-t");
        command.add(formatSeconds(spec.duration()));
        command.add("-vf");
        command.add(videoFilter(spec));
        command.add("-r");
        command.add(String.valueOf(spec.frameRate()));
        command.add("-c:v");
        command.add("libx264");
        command.add("-preset");
        command.add("fast");
        command.add("-b:v");
        command.add(spec.bitrate() + "k");
        command.add("-pix_fmt");
        command.add("yuv420p");
        if (!spec.isImageSource()) {
            command.add("-c:a");
            command.add("aac");
            command.add("-b:a");
            command.add("128k");
        }
        command.add("-movflags");
        command.add("+faststart");
        command.add(spec.outputPath().toString());
        return command;
    }

    private String videoFilter(RenderSpec spec) {
        List<String> filters = new ArrayList<>();
        filters.add("scale=%d:%d:force_original_aspect_ratio=decrease".formatted(spec.width(), spec.height()));
        filters.add("pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black".formatted(spec.width(), spec.height()));
        filters.add("setsar=1");

        if (spec.textClips().isArray()) {
            for (JsonNode textClip : spec.textClips()) {
                String text = text(textClip, "text");
                if (!StringUtils.hasText(text)) {
                    continue;
                }
                double start = Math.max(0, textClip.path("startTime").asDouble(0));
                double end = start + Math.max(0.1, textClip.path("duration").asDouble(5));
                JsonNode style = textClip.path("style");
                JsonNode transform = textClip.path("transform");
                int fontSize = positiveInt(style, "fontSize", 48);
                String color = normalizeColor(text(style, "color"), "white");
                String x = expressionPosition(transform.path("position").path("x").asDouble(0.5), "w", "text_w");
                String y = expressionPosition(transform.path("position").path("y").asDouble(0.5), "h", "text_h");
                filters.add("drawtext=text='%s':fontsize=%d:fontcolor=%s:x=%s:y=%s:enable='between(t\\,%s\\,%s)'"
                        .formatted(escapeDrawText(text), fontSize, color, x, y, formatSeconds(start), formatSeconds(end)));
            }
        }
        return String.join(",", filters);
    }

    private String expressionPosition(double normalized, String canvas, String item) {
        double clamped = Math.max(0, Math.min(1, normalized));
        return "(%s-%s)*%.4f".formatted(canvas, item, clamped);
    }

    private int run(List<String> command) throws Exception {
        ProcessBuilder builder = new ProcessBuilder(command);
        builder.redirectErrorStream(true);
        Process process = builder.start();
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
            while (reader.readLine() != null) {
                // Drain FFmpeg output to avoid blocking.
            }
        }
        return process.waitFor();
    }

    private String escapeDrawText(String value) {
        return value
                .replace("\\", "\\\\")
                .replace(":", "\\:")
                .replace(",", "\\,")
                .replace("'", "\\'")
                .replace("%", "\\%")
                .replace("\r", " ")
                .replace("\n", "\\n");
    }

    private String normalizeColor(String color, String fallback) {
        if (!StringUtils.hasText(color)) {
            return fallback;
        }
        String value = color.trim();
        if (value.matches("^#[0-9A-Fa-f]{6}$")) {
            return "0x" + value.substring(1);
        }
        if (value.matches("^[A-Za-z]+$")) {
            return value.toLowerCase(Locale.ROOT);
        }
        return fallback;
    }

    private int positiveInt(JsonNode node, String field, int fallback) {
        int value = node.path(field).asInt(fallback);
        return value > 0 ? value : fallback;
    }

    private double positiveDouble(JsonNode node, String field, double fallback) {
        double value = node.path(field).asDouble(fallback);
        return value > 0 ? value : fallback;
    }

    private String text(JsonNode node, String field) {
        JsonNode value = node == null ? null : node.get(field);
        return value != null && !value.isNull() ? value.asText() : null;
    }

    private String formatSeconds(double seconds) {
        return String.format(Locale.ROOT, "%.3f", Math.max(0, seconds));
    }

    private String safeProjectId(String projectId) {
        if (!StringUtils.hasText(projectId)) {
            return "default";
        }
        return projectId.matches("^[A-Za-z0-9_-]{1,80}$") ? projectId : "default";
    }

    public record RenderResult(String localPath, String objectKey, String outputUrl) {
    }

    private record RenderSpec(Path sourcePath,
                              String sourceType,
                              Path outputPath,
                              int width,
                              int height,
                              int frameRate,
                              int bitrate,
                              double duration,
                              double inPoint,
                              JsonNode textClips) {
        boolean isImageSource() {
            return "image".equals(sourceType);
        }
    }
}
