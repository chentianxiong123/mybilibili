package com.mybilibili.videomedia.mq;

import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.ManuscriptCommentCountEvent;
import com.mybilibili.video.service.ManuscriptService;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_MANUSCRIPT_COMMENT_COUNT,
        consumerGroup = MQConstants.GROUP_MANUSCRIPT_COMMENT_COUNT
)
public class ManuscriptCommentCountConsumer implements RocketMQListener<ManuscriptCommentCountEvent> {

    private final ManuscriptService manuscriptService;

    public ManuscriptCommentCountConsumer(ManuscriptService manuscriptService) {
        this.manuscriptService = manuscriptService;
    }

    @Override
    public void onMessage(ManuscriptCommentCountEvent event) {
        if (event == null || event.getManuscriptId() == null || event.getDelta() == null || event.getDelta() == 0) {
            log.warn("忽略无效稿件评论数事件: {}", event);
            return;
        }
        if (event.getDelta() > 0) {
            for (int i = 0; i < event.getDelta(); i++) {
                manuscriptService.incrementCommentCount(event.getManuscriptId());
            }
            return;
        }
        for (int i = 0; i < Math.abs(event.getDelta()); i++) {
            manuscriptService.decrementCommentCount(event.getManuscriptId());
        }
    }
}
