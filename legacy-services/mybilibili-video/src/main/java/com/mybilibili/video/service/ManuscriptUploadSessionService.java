package com.mybilibili.video.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.operation.OperationTaskRecorder;
import com.mybilibili.common.utils.UploadFilePathUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.video.dto.CreateUploadSessionRequest;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardCopyOption;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.UUID;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Slf4j
@Service
public class ManuscriptUploadSessionService {

    private static final Pattern CHUNK_FILE_PATTERN = Pattern.compile("^chunk-(\\d+)$");
    private static final Pattern UPLOAD_ID_PATTERN = Pattern.compile("^[a-f0-9]{32}$");
    private static final String SESSION_FILE_NAME = "session.json";

    private final ManuscriptService manuscriptService;
    private final UploadFilePathUtils uploadFilePathUtils;
    private final ObjectMapper objectMapper;
    private final OperationTaskRecorder operationTaskRecorder;

    @Autowired
    public ManuscriptUploadSessionService(ManuscriptService manuscriptService,
                                          UploadFilePathUtils uploadFilePathUtils,
                                          ObjectMapper objectMapper,
                                          OperationTaskRecorder operationTaskRecorder) {
        this.manuscriptService = manuscriptService;
        this.uploadFilePathUtils = uploadFilePathUtils;
        this.objectMapper = objectMapper;
        this.operationTaskRecorder = operationTaskRecorder;
    }

    public UploadSessionInfo createSession(Integer userId, CreateUploadSessionRequest request) {
        validateRequest(request);
        String uploadId = UUID.randomUUID().toString().replace("-", "");
        UploadSessionMetadata metadata = new UploadSessionMetadata(
                uploadId,
                userId,
                request.getTitle().trim(),
                StringUtils.hasText(request.getDescription()) ? request.getDescription().trim() : null,
                request.getCategoryId(),
                normalizeTags(request.getTags()),
                normalizeVideos(request.getVideos())
        );
        storeSessionMetadata(metadata);
        recordPending(userId, uploadId, metadata.title(), metadata.videos().size());
        return new UploadSessionInfo(uploadId, metadata.videos().size());
    }

    public UploadSessionStatus getSessionStatus(Integer userId, String uploadId) {
        UploadSessionMetadata metadata = loadSessionMetadata(uploadId);
        ensureOwner(metadata, userId);
        List<PartStatus> parts = new ArrayList<>();
        int uploadedParts = 0;
        int totalChunks = 0;
        for (int i = 0; i < metadata.videos().size(); i++) {
            UploadVideoPartMetadata part = metadata.videos().get(i);
            List<Integer> uploadedChunks = listUploadedChunkIndexes(uploadId, i);
            boolean complete = isPartComplete(part, uploadedChunks);
            if (complete) {
                uploadedParts++;
            }
            totalChunks += Math.max(part.totalChunks(), 0);
            parts.add(new PartStatus(
                    i,
                    part.title(),
                    part.fileName(),
                    part.totalChunks(),
                    uploadedChunks,
                    complete
            ));
        }
        return new UploadSessionStatus(uploadId, metadata.videos().size(), uploadedParts, totalChunks, parts);
    }

    public UploadChunkStatus uploadChunk(Integer userId, String uploadId, int partIndex, int chunkIndex,
                                         int totalChunks, MultipartFile file) throws IOException {
        UploadSessionMetadata metadata = loadSessionMetadata(uploadId);
        ensureOwner(metadata, userId);
        if (file == null || file.isEmpty()) {
            throw new BusinessException("分片文件不能为空");
        }
        if (partIndex < 0 || partIndex >= metadata.videos().size()) {
            throw new BusinessException("分片索引无效");
        }
        if (chunkIndex < 0) {
            throw new BusinessException("分片序号无效");
        }
        if (totalChunks <= 0) {
            throw new BusinessException("分片总数无效");
        }
        if (chunkIndex >= totalChunks) {
            throw new BusinessException("分片序号超出范围");
        }

        UploadVideoPartMetadata part = metadata.videos().get(partIndex);
        if (part.totalChunks() > 0 && totalChunks > 0 && part.totalChunks() != totalChunks) {
            throw new BusinessException("分片总数不一致");
        }

        try {
            Path chunkPath = chunkPath(uploadId, partIndex, chunkIndex);
            Files.createDirectories(chunkPath.getParent());
            file.transferTo(chunkPath.toFile());

            List<Integer> uploadedChunks = listUploadedChunkIndexes(uploadId, partIndex);
            boolean complete = isPartComplete(part, uploadedChunks);
            recordRunning(userId, uploadId, metadata.title(), metadata, "分片已上传");
            return new UploadChunkStatus(
                    uploadId,
                    partIndex,
                    chunkIndex,
                    uploadedChunks.size(),
                    part.totalChunks(),
                    complete
            );
        } catch (IOException e) {
            recordFailed(userId, uploadId, metadata.title(), 0, "UPLOAD_FAILED", "分片上传失败", e.getMessage());
            throw e;
        }
    }

    public ManuscriptVO completeSession(Integer userId, String uploadId, MultipartFile cover) throws Exception {
        UploadSessionMetadata metadata = loadSessionMetadata(uploadId);
        ensureOwner(metadata, userId);
        if (cover == null || cover.isEmpty()) {
            recordFailed(userId, uploadId, metadata.title(), calculateProgress(metadata, uploadId),
                    "COMPLETING", "稿件上传失败", "请上传封面");
            throw new BusinessException("请上传封面");
        }

        ManuscriptUploadDTO dto = new ManuscriptUploadDTO();
        dto.setTitle(metadata.title());
        dto.setDescription(metadata.description());
        dto.setCategoryId(metadata.categoryId());
        dto.setTags(metadata.tags());
        dto.setCover(cover);

        List<ManuscriptUploadDTO.VideoItemDTO> videos = new ArrayList<>();
        for (int i = 0; i < metadata.videos().size(); i++) {
            UploadVideoPartMetadata part = metadata.videos().get(i);
            List<Integer> uploadedChunks = listUploadedChunkIndexes(uploadId, i);
            if (!isPartComplete(part, uploadedChunks)) {
                recordFailed(userId, uploadId, metadata.title(), calculateProgress(metadata, uploadId),
                        "COMPLETING", "稿件上传失败", "视频分片尚未上传完成");
                throw new BusinessException("视频分片尚未上传完成");
            }
            File mergedFile = mergePart(uploadId, i, part);
            ManuscriptUploadDTO.VideoItemDTO videoItem = new ManuscriptUploadDTO.VideoItemDTO();
            videoItem.setTitle(part.title());
            videoItem.setVideoOrder(part.videoOrder() == null ? i : part.videoOrder());
            videoItem.setDurationSeconds(part.durationSeconds());
            videoItem.setVideo(new FileBackedMultipartFile(
                    part.fileName(),
                    mergedFile,
                    "application/octet-stream"
            ));
            videos.add(videoItem);
        }
        dto.setVideos(videos);

        try {
            ManuscriptVO vo = manuscriptService.uploadManuscript(dto, userId);
            recordSuccess(userId, uploadId, metadata.title(), "稿件上传完成");
            deleteSession(uploadId);
            return vo;
        } catch (Exception e) {
            recordFailed(userId, uploadId, metadata.title(), 100, "COMPLETING", "稿件上传失败", e.getMessage());
            log.error("分片稿件合并上传失败: uploadId={}", uploadId, e);
            throw e;
        }
    }

    public boolean cancelSession(Integer userId, String uploadId) {
        UploadSessionMetadata metadata = loadSessionMetadata(uploadId);
        ensureOwner(metadata, userId);
        deleteSession(uploadId);
        recordCancelled(userId, uploadId, metadata.title(), "上传会话已取消");
        return true;
    }

    private void validateRequest(CreateUploadSessionRequest request) {
        if (request == null) {
            throw new BusinessException("上传请求不能为空");
        }
        if (!StringUtils.hasText(request.getTitle())) {
            throw new BusinessException("稿件标题不能为空");
        }
        if (request.getCategoryId() == null) {
            throw new BusinessException("请选择稿件分区");
        }
        if (request.getVideos() == null || request.getVideos().isEmpty()) {
            throw new BusinessException("至少需要一个视频分P");
        }
    }

    private List<String> normalizeTags(List<String> tags) {
        if (tags == null) {
            return List.of();
        }
        return tags.stream()
                .filter(StringUtils::hasText)
                .map(String::trim)
                .distinct()
                .toList();
    }

    private List<UploadVideoPartMetadata> normalizeVideos(List<CreateUploadSessionRequest.VideoPart> parts) {
        List<UploadVideoPartMetadata> videos = new ArrayList<>();
        for (int i = 0; i < parts.size(); i++) {
            CreateUploadSessionRequest.VideoPart part = parts.get(i);
            if (part == null || !StringUtils.hasText(part.getFileName())) {
                throw new BusinessException("视频分P信息无效");
            }
            String fileName = sanitizeFileName(part.getFileName());
            if (part.getTotalChunks() == null || part.getTotalChunks() <= 0) {
                throw new BusinessException("视频分片总数无效");
            }
            videos.add(new UploadVideoPartMetadata(
                    StringUtils.hasText(part.getTitle()) ? part.getTitle().trim() : stripExtension(fileName),
                    fileName,
                    part.getSize(),
                    part.getVideoOrder() == null ? i : part.getVideoOrder(),
                    part.getDurationSeconds(),
                    part.getTotalChunks()
            ));
        }
        return videos;
    }

    private static final java.util.Set<String> ALLOWED_VIDEO_EXTENSIONS = java.util.Set.of(
            ".mp4", ".avi", ".mkv", ".mov", ".flv", ".wmv", ".webm", ".m4v", ".ts", ".mpeg", ".mpg", ".3gp"
    );

    private String sanitizeFileName(String fileName) {
        String normalized = fileName.replace("\\", "/");
        int slashIndex = normalized.lastIndexOf('/');
        if (slashIndex >= 0) {
            normalized = normalized.substring(slashIndex + 1);
        }
        normalized = normalized.replace("\u0000", "").trim();
        if (!StringUtils.hasText(normalized)) {
            throw new BusinessException("视频文件名无效");
        }
        String ext = fileExtension(normalized).toLowerCase(java.util.Locale.ROOT);
        if (!ALLOWED_VIDEO_EXTENSIONS.contains(ext)) {
            throw new BusinessException("不支持的视频格式: " + ext + "，支持: mp4/avi/mkv/mov/flv/wmv/webm/m4v/ts/mpeg/3gp");
        }
        return normalized;
    }

    private String stripExtension(String fileName) {
        if (!StringUtils.hasText(fileName) || !fileName.contains(".")) {
            return fileName;
        }
        return fileName.substring(0, fileName.lastIndexOf('.'));
    }

    private Path sessionDir(String uploadId) {
        validateUploadId(uploadId);
        return Path.of(uploadFilePathUtils.getBasePath(), "manuscript-upload-sessions", uploadId);
    }

    private void validateUploadId(String uploadId) {
        if (!StringUtils.hasText(uploadId) || !UPLOAD_ID_PATTERN.matcher(uploadId).matches()) {
            throw new BusinessException("上传会话不存在");
        }
    }

    private Path metadataPath(String uploadId) {
        return sessionDir(uploadId).resolve(SESSION_FILE_NAME);
    }

    private Path chunkPath(String uploadId, int partIndex, int chunkIndex) {
        return sessionDir(uploadId)
                .resolve("chunks")
                .resolve("part-" + partIndex)
                .resolve("chunk-" + chunkIndex);
    }

    private Path mergedPath(String uploadId, int partIndex, UploadVideoPartMetadata part) {
        return sessionDir(uploadId)
                .resolve("assembled")
                .resolve("part-" + partIndex + fileExtension(part.fileName()));
    }

    private void storeSessionMetadata(UploadSessionMetadata metadata) {
        try {
            Files.createDirectories(sessionDir(metadata.uploadId()));
            objectMapper.writeValue(metadataPath(metadata.uploadId()).toFile(), metadata);
        } catch (IOException e) {
            throw new BusinessException("创建上传会话失败");
        }
    }

    private UploadSessionMetadata loadSessionMetadata(String uploadId) {
        try {
            File file = metadataPath(uploadId).toFile();
            if (!file.exists()) {
                throw new BusinessException("上传会话不存在");
            }
            return objectMapper.readValue(file, UploadSessionMetadata.class);
        } catch (BusinessException e) {
            throw e;
        } catch (Exception e) {
            throw new BusinessException("读取上传会话失败");
        }
    }

    private void ensureOwner(UploadSessionMetadata metadata, Integer userId) {
        if (metadata.userId() == null || !metadata.userId().equals(userId)) {
            throw new BusinessException("无权操作该上传会话");
        }
    }

    private List<Integer> listUploadedChunkIndexes(String uploadId, int partIndex) {
        Path partDir = sessionDir(uploadId).resolve("chunks").resolve("part-" + partIndex);
        if (!Files.exists(partDir)) {
            return List.of();
        }
        try (var stream = Files.list(partDir)) {
            return stream
                    .map(path -> path.getFileName().toString())
                    .map(CHUNK_FILE_PATTERN::matcher)
                    .filter(Matcher::matches)
                    .map(matcher -> Integer.parseInt(matcher.group(1)))
                    .sorted()
                    .toList();
        } catch (IOException e) {
            throw new BusinessException("读取分片状态失败");
        }
    }

    private boolean isPartComplete(UploadVideoPartMetadata part, List<Integer> uploadedChunks) {
        if (part.totalChunks() <= 0) {
            return !uploadedChunks.isEmpty();
        }
        if (uploadedChunks.size() != part.totalChunks()) {
            return false;
        }
        for (int i = 0; i < part.totalChunks(); i++) {
            if (!uploadedChunks.contains(i)) {
                return false;
            }
        }
        return true;
    }

    private File mergePart(String uploadId, int partIndex, UploadVideoPartMetadata part) throws IOException {
        Path targetPath = mergedPath(uploadId, partIndex, part);
        Files.createDirectories(targetPath.getParent());
        try (OutputStream outputStream = Files.newOutputStream(targetPath)) {
            for (int i = 0; i < Math.max(part.totalChunks(), 0); i++) {
                Path chunk = chunkPath(uploadId, partIndex, i);
                if (!Files.exists(chunk)) {
                    throw new BusinessException("分片不完整");
                }
                Files.copy(chunk, outputStream);
            }
        }
        return targetPath.toFile();
    }

    private void deleteSession(String uploadId) {
        Path dir = sessionDir(uploadId);
        if (!Files.exists(dir)) {
            return;
        }
        try (var stream = Files.walk(dir)) {
            stream
                    .sorted(Comparator.reverseOrder())
                    .forEach(path -> {
                        try {
                            Files.deleteIfExists(path);
                        } catch (IOException e) {
                            log.warn("删除上传会话文件失败: {}", path, e);
                        }
                    });
        } catch (IOException e) {
            log.warn("删除上传会话目录失败: {}", uploadId, e);
        }
    }

    private void recordPending(Integer userId, String uploadId, String title, int totalParts) {
        operationTaskRecorder.pending(userId, taskKey(uploadId), "UPLOAD", title, "upload_session", uploadId,
                "PENDING", "上传会话已创建，等待分片上传");
    }

    private void recordRunning(Integer userId, String uploadId, String title,
                               UploadSessionMetadata metadata, String message) {
        operationTaskRecorder.running(userId, taskKey(uploadId), "UPLOAD", title, "upload_session", uploadId,
                calculateProgress(metadata, uploadId), "UPLOADING", message);
    }

    private void recordSuccess(Integer userId, String uploadId, String title, String message) {
        operationTaskRecorder.success(userId, taskKey(uploadId), "UPLOAD", title, "upload_session", uploadId,
                "COMPLETED", message);
    }

    private void recordFailed(Integer userId, String uploadId, String title, int progress,
                              String stage, String message, String errorMessage) {
        operationTaskRecorder.failed(userId, taskKey(uploadId), "UPLOAD", title, "upload_session", uploadId,
                progress, stage, message, errorMessage);
    }

    private void recordCancelled(Integer userId, String uploadId, String title, String message) {
        operationTaskRecorder.cancelled(userId, taskKey(uploadId), "UPLOAD", title, "upload_session", uploadId,
                "CANCELLED", message);
    }

    private String taskKey(String uploadId) {
        return "upload:" + uploadId;
    }

    private int calculateProgress(UploadSessionMetadata metadata, String uploadId) {
        int totalChunks = 0;
        int uploadedChunks = 0;
        for (int i = 0; i < metadata.videos().size(); i++) {
            UploadVideoPartMetadata part = metadata.videos().get(i);
            List<Integer> chunks = listUploadedChunkIndexes(uploadId, i);
            totalChunks += Math.max(part.totalChunks(), 0);
            uploadedChunks += chunks.size();
        }
        if (totalChunks <= 0) {
            return metadata.videos().isEmpty() ? 0 : 100;
        }
        return (int) Math.min(100, Math.round(uploadedChunks * 100.0 / totalChunks));
    }

    private String fileExtension(String fileName) {
        if (!StringUtils.hasText(fileName) || !fileName.contains(".")) {
            return ".mp4";
        }
        return fileName.substring(fileName.lastIndexOf('.'));
    }

    public record UploadSessionInfo(String uploadId, int totalParts) {
    }

    public record UploadChunkStatus(String uploadId, int partIndex, int chunkIndex, int uploadedChunks,
                                    int totalChunks, boolean complete) {
    }

    public record PartStatus(int partIndex, String title, String fileName, int totalChunks,
                             List<Integer> uploadedChunks, boolean complete) {
    }

    public record UploadSessionStatus(String uploadId, int totalParts, int uploadedParts, int totalChunks,
                                      List<PartStatus> parts) {
    }

    public record UploadSessionMetadata(String uploadId, Integer userId, String title, String description,
                                        Integer categoryId, List<String> tags, List<UploadVideoPartMetadata> videos) {
    }

    public record UploadVideoPartMetadata(String title, String fileName, Long size, Integer videoOrder,
                                          Integer durationSeconds, Integer totalChunks) {
    }

    private static final class FileBackedMultipartFile implements MultipartFile {
        private final String name;
        private final File file;
        private final String contentType;

        private FileBackedMultipartFile(String name, File file, String contentType) {
            this.name = name;
            this.file = file;
            this.contentType = contentType;
        }

        @Override
        public String getName() {
            return name;
        }

        @Override
        public String getOriginalFilename() {
            return name;
        }

        @Override
        public String getContentType() {
            return contentType;
        }

        @Override
        public boolean isEmpty() {
            return !file.exists() || file.length() == 0;
        }

        @Override
        public long getSize() {
            return file.length();
        }

        @Override
        public byte[] getBytes() throws IOException {
            return Files.readAllBytes(file.toPath());
        }

        @Override
        public InputStream getInputStream() throws IOException {
            return Files.newInputStream(file.toPath());
        }

        @Override
        public void transferTo(File dest) throws IOException, IllegalStateException {
            Files.copy(file.toPath(), dest.toPath(), StandardCopyOption.REPLACE_EXISTING);
        }
    }
}
