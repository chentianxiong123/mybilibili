package com.mybilibili.mq;

import org.apache.rocketmq.client.producer.SendCallback;
import org.apache.rocketmq.client.producer.SendResult;
import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
public class UserEventMQProducer {

    @Autowired
    private RocketMQTemplate rocketMQTemplate;

    public void sendExperienceEvent(UserExperienceEvent event) {
        rocketMQTemplate.asyncSend(
                MQConstants.TOPIC_USER_EXPERIENCE,
                event,
                quietCallback("用户经验事件", event == null ? null : event.getUserId())
        );
    }

    public void sendNotificationEvent(UserNotificationEvent event) {
        rocketMQTemplate.asyncSend(
                MQConstants.TOPIC_USER_NOTIFICATION,
                event,
                quietCallback("用户通知事件", event == null ? null : event.getReceiverId())
        );
    }

    public void sendUserProfileInitEvent(UserProfileInitEvent event) {
        rocketMQTemplate.asyncSend(
                MQConstants.TOPIC_USER_PROFILE_INIT,
                event,
                quietCallback("用户画像初始化事件", event == null ? null : event.getUserId())
        );
    }

    private SendCallback quietCallback(String label, Integer userId) {
        return new SendCallback() {
            @Override
            public void onSuccess(SendResult sendResult) {
                // Side-effect events are intentionally quiet on success.
            }

            @Override
            public void onException(Throwable e) {
                System.err.println(label + "发送失败: userId=" + userId + ", error=" + e.getMessage());
            }
        };
    }
}
