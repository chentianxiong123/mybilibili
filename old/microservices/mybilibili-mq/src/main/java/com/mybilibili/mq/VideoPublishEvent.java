package com.mybilibili.mq;

import lombok.Data;

import java.io.Serializable;

/**
 * 视频处理完成 → 自动上架事件
 * 由 AI 服务(AI_SUMMARY 完成后)发送,video-media 服务监听后:
 * 1. 检查稿件下所有 video 是否全部 PROCESS_STATUS_COMPLETED
 * 2. 若是,改 manuscript.status = STATUS_PUBLISHED
 * 3. 转发到 search-recommend 增量索引该稿件
 */
@Data
public class VideoPublishEvent implements Serializable {

    private static final long serialVersionUID = 1L;

    private Integer manuscriptId;
    private Integer videoId;
    private String trigger;  // AUTO_CHAIN / MANUAL_RETRY
}
