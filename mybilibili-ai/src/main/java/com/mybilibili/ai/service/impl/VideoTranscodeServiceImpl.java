package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.service.VideoProgressSseService;
import com.mybilibili.ai.service.VideoTranscodeService;
import com.mybilibili.ai.utils.FFmpegUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class VideoTranscodeServiceImpl implements VideoTranscodeService {

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

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
        try {
            String sourceVideoPath = uploadFilePathConfig.getVideoSourcePath(manuscriptId, videoId);
            String transcodedDir = uploadFilePathConfig.getTranscodedDir(manuscriptId, videoId);

            java.io.File videoFile = new java.io.File(sourceVideoPath);
            if (!videoFile.exists()) {
                notifyProgress(videoId, listener, 0, "源视频不存在");
                return new StepTranscodeResult(false, "源视频不存在", null, null, null);
            }

            String hdDir = transcodedDir + "/1080p";
            String sdDir = transcodedDir + "/720p";
            String ldDir = transcodedDir + "/480p";

            notifyProgress(videoId, listener, 0, "开始HLS转码");

            boolean hdSuccess = transcodeVariant(videoId, sourceVideoPath, hdDir, 1920, 1080, 0, 34, "1080p", listener);
            if (!hdSuccess) {
                notifyProgress(videoId, listener, 0, "1080p转码失败");
                return new StepTranscodeResult(false, "1080p转码失败", null, null, null);
            }

            boolean sdSuccess = transcodeVariant(videoId, sourceVideoPath, sdDir, 1280, 720, 35, 69, "720p", listener);
            if (!sdSuccess) {
                notifyProgress(videoId, listener, 0, "720p转码失败");
                return new StepTranscodeResult(false, "720p转码失败", null, null, null);
            }

            boolean ldSuccess = transcodeVariant(videoId, sourceVideoPath, ldDir, 854, 480, 70, 100, "480p", listener);
            if (!ldSuccess) {
                notifyProgress(videoId, listener, 0, "480p转码失败");
                return new StepTranscodeResult(false, "480p转码失败", null, null, null);
            }

            notifyProgress(videoId, listener, 100, "转码完成");
            return new StepTranscodeResult(
                    true,
                    null,
                    buildPlayUrl(manuscriptId, videoId, "1080p"),
                    buildPlayUrl(manuscriptId, videoId, "720p"),
                    buildPlayUrl(manuscriptId, videoId, "480p")
            );

        } catch (Exception e) {
            notifyProgress(videoId, listener, 0, "转码失败");
            return new StepTranscodeResult(false, e.getMessage(), null, null, null);
        }
    }

    @Override
    public String getVideoPath(Integer manuscriptId, Integer videoId, String quality) {
        String basePath = uploadFilePathConfig.getTranscodedDir(manuscriptId, videoId);
        return basePath + "/" + quality + "/playlist.m3u8";
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

    private String buildPlayUrl(Integer manuscriptId, Integer videoId, String quality) {
        return "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/transcoded/" + quality + "/playlist.m3u8";
    }
}
