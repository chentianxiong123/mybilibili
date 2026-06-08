package com.mybilibili.video.controller;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.studio.StudioAssetUploadResponse;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;

import java.util.Set;

@RestController
@RequestMapping("/studio/assets")
@Tag(name = "剪辑工作室素材", description = "剪辑工作室云端素材接口")
public class StudioAssetController {
    private static final long MAX_ASSET_SIZE = 500L * 1024 * 1024;
    private static final Set<String> ALLOWED_PREFIXES = Set.of("video/", "audio/", "image/");

    private final StorageService storageService;

    public StudioAssetController(StorageService storageService) {
        this.storageService = storageService;
    }

    @PostMapping(value = "/upload", consumes = "multipart/form-data")
    @Operation(summary = "上传剪辑素材")
    public Result<StudioAssetUploadResponse> uploadAsset(@RequestParam("projectId") String projectId,
                                                         @RequestParam("mediaId") String mediaId,
                                                         @RequestParam("file") MultipartFile file,
                                                         HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            if (userId == null) {
                throw new BusinessException("用户未登录");
            }
            validateProjectId(projectId);
            validateMediaId(mediaId);
            validateFile(file);

            String safeFileName = sanitizeFileName(file.getOriginalFilename());
            String contentType = StringUtils.hasText(file.getContentType()) ? file.getContentType() : "application/octet-stream";
            String objectKey = StorageKeys.studioAsset(userId, projectId, mediaId, safeFileName);
            String url = storageService.upload(objectKey, file.getInputStream(), file.getSize(), contentType);

            return Result.success("上传成功", new StudioAssetUploadResponse(
                    mediaId,
                    objectKey,
                    url,
                    safeFileName,
                    contentType,
                    file.getSize()
            ));
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    private void validateProjectId(String projectId) {
        if (!StringUtils.hasText(projectId) || !projectId.matches("^[A-Za-z0-9_-]{1,80}$")) {
            throw new BusinessException("项目 ID 无效");
        }
    }

    private void validateMediaId(String mediaId) {
        if (!StringUtils.hasText(mediaId) || !mediaId.matches("^[A-Za-z0-9_-]{1,80}$")) {
            throw new BusinessException("素材 ID 无效");
        }
    }

    private void validateFile(MultipartFile file) {
        if (file == null || file.isEmpty()) {
            throw new BusinessException("素材文件不能为空");
        }
        if (file.getSize() > MAX_ASSET_SIZE) {
            throw new BusinessException("素材文件不能超过 500MB");
        }
        String contentType = file.getContentType();
        boolean allowed = StringUtils.hasText(contentType) && ALLOWED_PREFIXES.stream().anyMatch(contentType::startsWith);
        if (!allowed) {
            throw new BusinessException("仅支持视频、音频和图片素材");
        }
    }

    private String sanitizeFileName(String fileName) {
        String normalized = StringUtils.hasText(fileName) ? fileName.replace("\\", "/") : "asset.bin";
        int slashIndex = normalized.lastIndexOf('/');
        if (slashIndex >= 0) {
            normalized = normalized.substring(slashIndex + 1);
        }
        normalized = normalized.replace("\u0000", "").trim();
        if (!StringUtils.hasText(normalized)) {
            return "asset.bin";
        }
        return normalized.replaceAll("[^A-Za-z0-9._-]", "_");
    }
}
