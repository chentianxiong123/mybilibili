package com.mybilibili.danmaku.config;

import org.apache.rocketmq.client.producer.DefaultMQProducer;
import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.apache.rocketmq.spring.support.RocketMQMessageConverter;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RocketMQConfig {

    @Value("${rocketmq.name-server:127.0.0.1:9876}")
    private String nameServer;

    @Bean
    public DefaultMQProducer danmakuMQProducer() {
        DefaultMQProducer producer = new DefaultMQProducer("danmaku-producer-group");
        producer.setNamesrvAddr(nameServer);
        producer.setInstanceName("danmaku-producer");
        return producer;
    }

    @Bean
    public RocketMQTemplate rocketMQTemplate(DefaultMQProducer danmakuMQProducer,
                                             RocketMQMessageConverter rocketMQMessageConverter) {
        RocketMQTemplate template = new RocketMQTemplate();
        template.setProducer(danmakuMQProducer);
        template.setMessageConverter(rocketMQMessageConverter.getMessageConverter());
        return template;
    }
}
