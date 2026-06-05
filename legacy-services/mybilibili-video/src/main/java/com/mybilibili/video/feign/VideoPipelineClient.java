package com.mybilibili.video.feign;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;

/**
 * 视频全流程处理 Feign 客户端
 */
@FeignClient(name = "mybilibili-ai", contextId = "videoPipeline")
public interface VideoPipelineClient {

    /**
     * 提交全流程处理任务
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @param uploaderId 上传者ID
     * @return 处理结果
     */
    @PostMapping("/api/pipeline/submit")
    Result<String> submitPipelineTask(
            @RequestParam("manuscriptId") Integer manuscriptId,
            @RequestParam("videoId") Integer videoId,
            @RequestParam(value = "uploaderId", required = false) Integer uploaderId);

    /**
     * 取消任务
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @return 处理结果
     */
    @PostMapping("/api/pipeline/cancel")
    Result<String> cancelTask(
            @RequestParam("manuscriptId") Integer manuscriptId,
            @RequestParam("videoId") Integer videoId);

    /**
     * 重试失败的任务
     *
     * @param manuscriptId 稿件ID
     * @param videoId 视频ID
     * @return 处理结果
     */
    @PostMapping("/api/pipeline/retry")
    Result<String> retryTask(
            @RequestParam("manuscriptId") Integer manuscriptId,
            @RequestParam("videoId") Integer videoId);
}
