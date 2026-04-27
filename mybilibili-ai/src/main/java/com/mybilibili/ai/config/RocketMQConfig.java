package com.mybilibili.ai.config;

import com.mybilibili.mq.VideoMQProducer;
import org.apache.rocketmq.client.producer.DefaultMQProducer;
import org.apache.rocketmq.spring.core.RocketMQTemplate;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class RocketMQConfig {

    @Value("${rocketmq.name-server:127.0.0.1:9876}")
    private String nameServer;

    @Bean
    public DefaultMQProducer defaultMQProducer() {
        DefaultMQProducer producer = new DefaultMQProducer("video-process-group");
        producer.setNamesrvAddr(nameServer);
        producer.setInstanceName("ai-video-producer");
        return producer;
    }

    @Bean
    public RocketMQTemplate rocketMQTemplate(DefaultMQProducer defaultMQProducer) {
        RocketMQTemplate template = new RocketMQTemplate();
        template.setProducer(defaultMQProducer);
        return template;
    }

    @Bean
    public VideoMQProducer videoMQProducer() {
        return new VideoMQProducer();
    }
}
