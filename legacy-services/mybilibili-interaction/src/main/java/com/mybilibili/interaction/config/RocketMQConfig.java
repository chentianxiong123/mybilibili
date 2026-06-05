package com.mybilibili.interaction.config;

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
    public DefaultMQProducer interactionMQProducer() {
        DefaultMQProducer producer = new DefaultMQProducer("interaction-producer-group");
        producer.setNamesrvAddr(nameServer);
        producer.setInstanceName("interaction-producer");
        return producer;
    }

    @Bean
    public RocketMQTemplate rocketMQTemplate(DefaultMQProducer interactionMQProducer) {
        RocketMQTemplate template = new RocketMQTemplate();
        template.setProducer(interactionMQProducer);
        return template;
    }
}
