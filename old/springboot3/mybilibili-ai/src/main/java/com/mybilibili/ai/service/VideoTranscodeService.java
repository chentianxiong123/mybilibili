package com.mybilibili.ai.service;

public interface VideoTranscodeService {

    interface TranscodeProgressListener {
        void onProgress(int progress, String stageText);
    }

    boolean transcode(Integer manuscriptId, Integer videoId);

    StepTranscodeResult transcodeStep(Integer manuscriptId, Integer videoId, TranscodeProgressListener listener);

    String getVideoPath(Integer manuscriptId, Integer videoId, String quality);

    boolean transcodeToHLS(Integer manuscriptId, Integer videoId);

    boolean transcodeToHLS(Integer manuscriptId, Integer videoId, TranscodeProgressListener listener);

    String getHLSPath(Integer manuscriptId, Integer videoId);

    class StepTranscodeResult {
        private final boolean success;
        private final String errorMessage;
        private final String playUrlHd;
        private final String playUrlSd;
        private final String playUrlLd;

        public StepTranscodeResult(boolean success, String errorMessage, String playUrlHd, String playUrlSd, String playUrlLd) {
            this.success = success;
            this.errorMessage = errorMessage;
            this.playUrlHd = playUrlHd;
            this.playUrlSd = playUrlSd;
            this.playUrlLd = playUrlLd;
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
}
