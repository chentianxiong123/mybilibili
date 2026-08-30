package com.mybilibili.analytics.consumer;

import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.ManuscriptIndexEvent;
import com.mybilibili.search.service.ManuscriptIndexService;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

/**
 * 稿件 ES 索引事件消费者
 * UPSERT: 上架/重新上架 → 增量索引该稿件
 * DELETE: 下架 → 从 ES 删除该稿件
 */
@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_MANUSCRIPT_INDEX,
        consumerGroup = "manuscript-index-group"
)
public class ManuscriptIndexEventConsumer implements RocketMQListener<ManuscriptIndexEvent> {

    private final ManuscriptIndexService manuscriptIndexService;

    public ManuscriptIndexEventConsumer(ManuscriptIndexService manuscriptIndexService) {
        this.manuscriptIndexService = manuscriptIndexService;
    }

    @Override
    public void onMessage(ManuscriptIndexEvent event) {
        if (event == null || event.getManuscriptId() == null || event.getOperation() == null) {
            log.warn("忽略无效稿件索引事件: {}", event);
            return;
        }
        log.info("收到稿件索引事件: manuscriptId={}, operation={}, trigger={}",
                event.getManuscriptId(), event.getOperation(), event.getTrigger());

        switch (event.getOperation()) {
            case ManuscriptIndexEvent.OPERATION_UPSERT:
                boolean indexed = manuscriptIndexService.indexOne(event.getManuscriptId());
                if (indexed) {
                    log.info("稿件索引成功: manuscriptId={}", event.getManuscriptId());
                } else {
                    log.warn("稿件索引失败(可能未上架): manuscriptId={}", event.getManuscriptId());
                }
                break;
            case ManuscriptIndexEvent.OPERATION_DELETE:
                boolean deleted = manuscriptIndexService.deleteOne(event.getManuscriptId());
                if (deleted) {
                    log.info("稿件索引删除成功: manuscriptId={}", event.getManuscriptId());
                } else {
                    log.warn("稿件索引删除异常: manuscriptId={}", event.getManuscriptId());
                }
                break;
            default:
                log.warn("未知操作类型: {} manuscriptId={}", event.getOperation(), event.getManuscriptId());
        }
    }
}
