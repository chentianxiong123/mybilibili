package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.service.VideoProcessingStorageService;
import com.mybilibili.ai.service.VideoProgressSseService;
import com.mybilibili.ai.service.VideoTranscodeService;
import com.mybilibili.ai.utils.FFmpegUtils;
import com.mybilibili.common.storage.StorageKeys;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.nio.file.Path;

@Service
public class VideoTranscodeServiceImpl implements VideoTranscodeService {

    @Autowired
    private VideoProcessingStorageService processingStorageService;

    @Autowired
    private FFmpegUtils ffmpegUtils;

    @Autowired
    private VideoProgressSseService progressSseService;

    @Override
    public boolean transcode(Integer manuscriptId, Integer videoId) {
        return transcodeStep(manuscriptId, videoId, null).isSuccess();
    }

    @Override
    public StepTranscodeResult transcodeStep(Integer manuscriptId, Integer videoId, TranscodeProgressListener listener) {
        Path transcodedDir = null;
        try {
            Path sourceVideoPath = processingStorageService.materializeSourceVideo(manuscriptId, videoId);
            transcodedDir = processingStorageService.getTranscodedDir(manuscriptId, videoId);
            processingStorageService.recreateDirectory(transcodedDir);

            Path hdDir = transcodedDir.resolve("1080p");
            Path sdDir = transcodedDir.resolve("720p");
            Path ldDir = transcodedDir.resolve("480p");

            notifyProgress(videoId, listener, 0, "开始HLS转码");

            boolean hdSuccess = transcodeVariant(videoId, sourceVideoPath.toString(), hdDir.toString(), 1920, 1080, 0, 34, "1080p", listener);
            if (!hdSuccess) {
                notifyProgress(videoId, listener, 0, "1080p转码失败");
                return new StepTranscodeResult(false, "1080p转码失败", null, null, null);
            }

            boolean sdSuccess = transcodeVariant(videoId, sourceVideoPath.toString(), sdDir.toString(), 1280, 720, 35, 69, "720p", listener);
            if (!sdSuccess) {
                notifyProgress(videoId, listener, 0, "720p转码失败");
                return new StepTranscodeResult(false, "720p转码失败", null, null, null);
            }

            boolean ldSuccess = transcodeVariant(videoId, sourceVideoPath.toString(), ldDir.toString(), 854, 480, 70, 100, "480p", listener);
            if (!ldSuccess) {
                notifyProgress(videoId, listener, 0, "480p转码失败");
                return new StepTranscodeResult(false, "480p转码失败", null, null, null);
            }

            String playUrlHd = processingStorageService.uploadHlsDirectory(manuscriptId, videoId, "1080p", hdDir);
            String playUrlSd = processingStorageService.uploadHlsDirectory(manuscriptId, videoId, "720p", sdDir);
            String playUrlLd = processingStorageService.uploadHlsDirectory(manuscriptId, videoId, "480p", ldDir);

            notifyProgress(videoId, listener, 100, "转码完成");
            return new StepTranscodeResult(
                    true,
                    null,
                    playUrlHd,
                    playUrlSd,
                    playUrlLd
            );

        } catch (Exception e) {
            notifyProgress(videoId, listener, 0, "转码失败");
            return new StepTranscodeResult(false, e.getMessage(), null, null, null);
        } finally {
            processingStorageService.deleteDirectory(transcodedDir);
        }
    }

    @Override
    public String getVideoPath(Integer manuscriptId, Integer videoId, String quality) {
        return processingStorageService.getPublicUrl(StorageKeys.videoHlsPlaylist(manuscriptId, videoId, quality));
    }

    @Override
    public boolean transcodeToHLS(Integer manuscriptId, Integer videoId) {
        return transcodeToHLS(manuscriptId, videoId, null);
    }

    @Override
    public boolean transcodeToHLS(Integer manuscriptId, Integer videoId, TranscodeProgressListener listener) {
        return transcodeStep(manuscriptId, videoId, listener).isSuccess();
    }

    @Override
    public String getHLSPath(Integer manuscriptId, Integer videoId) {
        return getVideoPath(manuscriptId, videoId, "1080p");
    }

    private boolean transcodeVariant(Integer videoId,
                                     String sourceVideoPath,
                                     String outputDir,
                                     int width,
                                     int height,
                                     int startProgress,
                                     int endProgress,
                                     String qualityLabel,
                                     TranscodeProgressListener listener) {
        notifyProgress(videoId, listener, startProgress, "转码" + qualityLabel + "中");
        return ffmpegUtils.transcodeToHLS(sourceVideoPath, outputDir, "playlist", width, height, ratio -> {
            int progress = mapRange(ratio, startProgress, endProgress);
            notifyProgress(videoId, listener, progress, "转码" + qualityLabel + "中");
        });
    }

    private int mapRange(double ratio, int startProgress, int endProgress) {
        if (ratio <= 0) {
            return startProgress;
        }
        if (ratio >= 1) {
            return endProgress;
        }
        return startProgress + (int) Math.round((endProgress - startProgress) * ratio);
    }

    private void notifyProgress(Integer videoId, TranscodeProgressListener listener, int progress, String stageText) {
        progressSseService.pushProgress(videoId, progress, stageText, "transcode");
        if (listener != null) {
            listener.onProgress(progress, stageText);
        }
    }
}
