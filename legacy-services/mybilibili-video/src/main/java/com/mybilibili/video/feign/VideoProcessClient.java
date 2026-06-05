package com.mybilibili.video.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

@FeignClient(name = "mybilibili-ai", contextId = "videoProcess")
public interface VideoProcessClient {

    @PostMapping("/ai/process/transcode/{videoId}")
    Result<Void> triggerTranscode(@PathVariable("videoId") Integer videoId,
                                   @RequestParam Integer manuscriptId);

    @PostMapping("/ai/process/audio/{videoId}")
    Result<Void> triggerAudioExtract(@PathVariable("videoId") Integer videoId,
                                      @RequestParam Integer manuscriptId);

    @PostMapping("/ai/process/subtitle/{videoId}")
    Result<Void> triggerSubtitleGenerate(@PathVariable("videoId") Integer videoId,
                                         @RequestParam Integer manuscriptId);

    @PostMapping("/ai/process/ai-summary/{videoId}")
    Result<Void> triggerAiSummary(@PathVariable("videoId") Integer videoId,
                                   @RequestParam Integer manuscriptId);
}
