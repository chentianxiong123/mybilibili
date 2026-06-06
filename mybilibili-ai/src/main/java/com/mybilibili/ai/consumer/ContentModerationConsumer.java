package com.mybilibili.ai.consumer;

import com.mybilibili.ai.config.DynamicModerationClient;
import com.mybilibili.ai.service.ModerationProvider;
import com.mybilibili.ai.service.ModerationProvider.ModerateRequest;
import com.mybilibili.mq.ContentModerationMessage;
import com.mybilibili.mq.ContentModerationResultEvent;
import com.mybilibili.mq.ContentReviewMQProducer;
import com.mybilibili.mq.MQConstants;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_CONTENT_MODERATION,
        consumerGroup = MQConstants.GROUP_CONTENT_MODERATION
)
public class ContentModerationConsumer implements RocketMQListener<ContentModerationMessage> {

    private final DynamicModerationClient moderationClient;
    private final ContentReviewMQProducer contentReviewMQProducer;

    public ContentModerationConsumer(DynamicModerationClient moderationClient,
                                     ContentReviewMQProducer contentReviewMQProducer) {
        this.moderationClient = moderationClient;
        this.contentReviewMQProducer = contentReviewMQProducer;
    }

    @Override
    public void onMessage(ContentModerationMessage message) {
        if (message == null || message.getTargetType() == null
                || message.getTargetId() == null || message.getContent() == null) {
            log.warn("忽略无效内容异步审核消息: {}", message);
            return;
        }

        ModerationProvider provider = moderationClient.getProvider();
        if (provider == null || !provider.isAvailable()) {
            log.debug("内容异步审核跳过: provider unavailable, targetType={}, targetId={}",
                    message.getTargetType(), message.getTargetId());
            return;
        }

        ModerationProvider.ModerationResult result = (ModerationProvider.ModerationResult) provider.invoke(
                ModerateRequest.of(message.getContent(), message.getTargetType()));
        contentReviewMQProducer.sendContentModerationResultEvent(
                ContentModerationResultEvent.of(
                        message.getTargetType(),
                        message.getTargetId(),
                        result.isPassed(),
                        result.getReason()));
    }
}
