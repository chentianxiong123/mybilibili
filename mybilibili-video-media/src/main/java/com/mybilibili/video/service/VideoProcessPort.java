package com.mybilibili.video.service;

import com.mybilibili.common.vo.Result;

public interface VideoProcessPort {

    Result<Void> triggerTranscode(Integer videoId, Integer manuscriptId);

    Result<Void> triggerAutoProcess(Integer videoId, Integer manuscriptId, Integer uploaderId);

    Result<Void> triggerAudioExtract(Integer videoId, Integer manuscriptId);

    Result<Void> triggerSubtitleGenerate(Integer videoId, Integer manuscriptId);

    Result<Void> triggerAiSummary(Integer videoId, Integer manuscriptId);
}
