package com.mybilibili.ai.service;

public interface VideoTranscodeService {
    
    boolean transcode(Integer manuscriptId, Integer videoId);
    
    String getVideoPath(Integer manuscriptId, Integer videoId, String quality);
}
