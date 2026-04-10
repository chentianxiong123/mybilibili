package com.mybilibili.ai.service;

public interface AiSubtitleService {

    boolean generateSubtitle(Integer manuscriptId, Integer videoId);

    String getSubtitlePath(Integer manuscriptId, Integer videoId);
}