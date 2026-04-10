package com.mybilibili.ai.service;

public interface AudioExtractService {
    boolean extractAudio(Integer manuscriptId, Integer videoId);
    String getAudioPath(Integer manuscriptId, Integer videoId);
}
