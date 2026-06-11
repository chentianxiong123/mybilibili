package com.mybilibili.videomedia.mq;

import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoPublishEvent;
import com.mybilibili.video.service.ManuscriptService;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

/**
 * 稿件自动上架消费者
 * 监听 AI 服务在 AI_SUMMARY 完成后发送的 VideoPublishEvent:
 * 1. 调 ManuscriptService.autoPublishIfAllVideosCompleted
 * 2. 该 Service 检查稿件下所有 video.process_status == COMPLETED
 * 3. 若全部完成,改 manuscript.status = PUBLISHED
 * 4. 后续(下个 PR)会加 ES 索引事件触发
 */
@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_VIDEO_PUBLISH,
        consumerGroup = "manuscript-auto-publish-group"
)
public class ManuscriptAutoPublishConsumer implements RocketMQListener<VideoPublishEvent> {

    private final ManuscriptService manuscriptService;

    public ManuscriptAutoPublishConsumer(ManuscriptService manuscriptService) {
        this.manuscriptService = manuscriptService;
    }

    @Override
    public void onMessage(VideoPublishEvent event) {
        if (event == null || event.getManuscriptId() == null) {
            log.warn("忽略无效稿件上架事件: {}", event);
            return;
        }
        log.info("收到稿件自动上架事件: manuscriptId={}, videoId={}, trigger={}",
                event.getManuscriptId(), event.getVideoId(), event.getTrigger());
        try {
            manuscriptService.autoPublishIfAllVideosCompleted(event.getManuscriptId());
        } catch (Exception e) {
            log.error("处理稿件自动上架事件失败: manuscriptId={}, error={}",
                    event.getManuscriptId(), e.getMessage(), e);
        }
    }
}
