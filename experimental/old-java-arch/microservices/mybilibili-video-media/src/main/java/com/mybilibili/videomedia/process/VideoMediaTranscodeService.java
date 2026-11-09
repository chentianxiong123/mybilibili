package com.mybilibili.videomedia.process;

import org.springframework.stereotype.Service;

import java.nio.file.Path;

@Service
public class VideoMediaTranscodeService {

    public interface ProgressListener {
        void onProgress(int progress, String stageText);
    }

    public static class TranscodeResult {
        private final boolean success;
        private final String errorMessage;
        private final String playUrlHd;
        private final String playUrlSd;
        private final String playUrlLd;

        private TranscodeResult(boolean success, String errorMessage, String playUrlHd, String playUrlSd, String playUrlLd) {
            this.success = success;
            this.errorMessage = errorMessage;
            this.playUrlHd = playUrlHd;
            this.playUrlSd = playUrlSd;
            this.playUrlLd = playUrlLd;
        }

        public static TranscodeResult success(String playUrlHd, String playUrlSd, String playUrlLd) {
            return new TranscodeResult(true, null, playUrlHd, playUrlSd, playUrlLd);
        }

        public static TranscodeResult failed(String errorMessage) {
            return new TranscodeResult(false, errorMessage, null, null, null);
        }

        public boolean isSuccess() {
            return success;
        }

        public String getErrorMessage() {
            return errorMessage;
        }

        public String getPlayUrlHd() {
            return playUrlHd;
        }

        public String getPlayUrlSd() {
            return playUrlSd;
        }

        public String getPlayUrlLd() {
            return playUrlLd;
        }
    }

    private final VideoMediaProcessingStorageService storageService;
    private final FfmpegMediaTool ffmpeg;

    public VideoMediaTranscodeService(VideoMediaProcessingStorageService storageService, FfmpegMediaTool ffmpeg) {
        this.storageService = storageService;
        this.ffmpeg = ffmpeg;
    }

    public TranscodeResult transcode(Integer manuscriptId, Integer videoId, ProgressListener listener) {
        Path transcodedDir = null;
        try {
            Path sourceVideoPath = storageService.materializeSourceVideo(manuscriptId, videoId);
            transcodedDir = storageService.getTranscodedDir(manuscriptId, videoId);
            storageService.recreateDirectory(transcodedDir);

            notify(listener, 0, "TRANSCODING");
            Path hdDir = transcodedDir.resolve("1080p");
            Path sdDir = transcodedDir.resolve("720p");
            Path ldDir = transcodedDir.resolve("480p");

            if (!transcodeVariant(sourceVideoPath, hdDir, 1920, 1080, 0, 34, listener)) {
                return TranscodeResult.failed("1080p transcode failed");
            }
            if (!transcodeVariant(sourceVideoPath, sdDir, 1280, 720, 35, 69, listener)) {
                return TranscodeResult.failed("720p transcode failed");
            }
            if (!transcodeVariant(sourceVideoPath, ldDir, 854, 480, 70, 100, listener)) {
                return TranscodeResult.failed("480p transcode failed");
            }

            String playUrlHd = storageService.uploadHlsDirectory(manuscriptId, videoId, "1080p", hdDir);
            String playUrlSd = storageService.uploadHlsDirectory(manuscriptId, videoId, "720p", sdDir);
            String playUrlLd = storageService.uploadHlsDirectory(manuscriptId, videoId, "480p", ldDir);
            notify(listener, 100, "TRANSCODE_SUCCESS");
            return TranscodeResult.success(playUrlHd, playUrlSd, playUrlLd);
        } catch (Exception e) {
            return TranscodeResult.failed(e.getMessage());
        } finally {
            storageService.deleteDirectory(transcodedDir);
        }
    }

    private boolean transcodeVariant(Path source,
                                     Path outputDir,
                                     int width,
                                     int height,
                                     int startProgress,
                                     int endProgress,
                                     ProgressListener listener) {
        return ffmpeg.transcodeToHLS(source.toString(), outputDir.toString(), "playlist", width, height, ratio -> {
            int progress = startProgress + (int) Math.round((endProgress - startProgress) * ratio);
            notify(listener, progress, "TRANSCODING");
        });
    }

    private void notify(ProgressListener listener, int progress, String stageText) {
        if (listener != null) {
            listener.onProgress(progress, stageText);
        }
    }
}
