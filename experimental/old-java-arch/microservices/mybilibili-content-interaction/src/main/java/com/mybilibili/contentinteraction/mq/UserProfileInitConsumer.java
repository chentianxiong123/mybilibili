package com.mybilibili.contentinteraction.mq;

import com.mybilibili.interaction.service.UserProfileService;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.UserProfileInitEvent;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_USER_PROFILE_INIT,
        consumerGroup = MQConstants.GROUP_USER_PROFILE_INIT
)
public class UserProfileInitConsumer implements RocketMQListener<UserProfileInitEvent> {

    private final UserProfileService userProfileService;

    public UserProfileInitConsumer(UserProfileService userProfileService) {
        this.userProfileService = userProfileService;
    }

    @Override
    public void onMessage(UserProfileInitEvent event) {
        if (event == null || event.getUserId() == null || event.getTags() == null || event.getTags().isEmpty()) {
            log.warn("忽略无效用户画像初始化事件: {}", event);
            return;
        }
        userProfileService.initProfileWithTags(event.getUserId(), event.getTags());
        log.debug("用户画像初始化事件处理完成: userId={}", event.getUserId());
    }
}
