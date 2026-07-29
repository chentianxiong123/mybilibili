package com.mybilibili.accountsocial.mq;

import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.UserExperienceEvent;
import com.mybilibili.user.service.UserService;
import lombok.extern.slf4j.Slf4j;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RocketMQMessageListener(
        topic = MQConstants.TOPIC_USER_EXPERIENCE,
        consumerGroup = MQConstants.GROUP_USER_EXPERIENCE
)
public class UserExperienceConsumer implements RocketMQListener<UserExperienceEvent> {

    private final UserService userService;

    public UserExperienceConsumer(UserService userService) {
        this.userService = userService;
    }

    @Override
    public void onMessage(UserExperienceEvent event) {
        if (event == null || event.getUserId() == null || event.getExperienceAmount() == null) {
            log.warn("忽略无效经验值事件: {}", event);
            return;
        }
        userService.addExperience(event.getUserId(), event.getExperienceAmount());
        log.debug("经验值事件处理完成: userId={}, amount={}, sourceType={}, sourceId={}",
                event.getUserId(), event.getExperienceAmount(), event.getSourceType(), event.getSourceId());
    }
}
