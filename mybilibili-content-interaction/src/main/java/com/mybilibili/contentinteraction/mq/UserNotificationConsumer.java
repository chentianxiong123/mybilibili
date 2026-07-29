package com.mybilibili.contentinteraction.mq;

import com.mybilibili.message.service.MessageService;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.UserNotificationEvent;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_USER_NOTIFICATION,
        consumerGroup = MQConstants.GROUP_USER_NOTIFICATION
)
public class UserNotificationConsumer implements RocketMQListener<UserNotificationEvent> {

    private final MessageService messageService;

    public UserNotificationConsumer(MessageService messageService) {
        this.messageService = messageService;
    }

    @Override
    public void onMessage(UserNotificationEvent event) {
        if (event == null || event.getType() == null) {
            log.warn("忽略无效通知事件: {}", event);
            return;
        }
        switch (event.getType()) {
            case UserNotificationEvent.TYPE_VIDEO_LIKE ->
                    messageService.sendLikeNotification(event.getSenderId(), event.getReceiverId(),
                            event.getTargetId(), event.getTitle());
            case UserNotificationEvent.TYPE_COMMENT_LIKE ->
                    messageService.sendCommentLikeNotification(event.getSenderId(), event.getReceiverId(),
                            event.getCommentId(), event.getContent());
            case UserNotificationEvent.TYPE_REPLY ->
                    messageService.sendReplyNotification(event.getSenderId(), event.getReceiverId(),
                            event.getContent(), event.getMessageType(), event.getTargetId(), event.getCommentId());
            case UserNotificationEvent.TYPE_SYSTEM ->
                    messageService.sendSystemNotification(event.getReceiverId(), event.getTitle(), event.getContent());
            default -> log.warn("忽略未知通知事件类型: {}", event.getType());
        }
    }
}

