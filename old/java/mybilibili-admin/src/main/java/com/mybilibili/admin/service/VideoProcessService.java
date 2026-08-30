package com.mybilibili.admin.service;

/**
 * 视频处理服务
 * 处理视频转码、音频提取、字幕生成、AI总结等流程
 */
public interface VideoProcessService {
    
    /**
     * 视频转码
     * @param videoId 视频ID
     */
    void transcodeVideo(Integer videoId);
    
    /**
     * 提取音频
     * @param videoId 视频ID
     */
    void extractAudio(Integer videoId);
    
    /**
     * 生成字幕
     * @param videoId 视频ID
     */
    void generateSubtitle(Integer videoId);
    
    /**
     * AI总结
     * @param videoId 视频ID
     */
    void aiSummary(Integer videoId);
    
    /**
     * 全流程处理
     * @param videoId 视频ID
     */
    void processAll(Integer videoId);
}
