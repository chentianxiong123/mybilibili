package com.mybilibili.ai.service;

public interface AiSubtitleService {

    boolean generateSubtitle(Integer manuscriptId, Integer videoId);

    boolean generateSubtitle(Integer manuscriptId, Integer videoId, ProgressListener progressListener);

    String getSubtitlePath(Integer manuscriptId, Integer videoId);

    @FunctionalInterface
    interface ProgressListener {
        void onProgress(int percent, String stageText);
    }
}
