package com.mybilibili.contentinteraction;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.context.annotation.FullyQualifiedAnnotationBeanNameGenerator;
import org.springframework.data.mongodb.repository.config.EnableMongoRepositories;

@SpringBootApplication(
        nameGenerator = FullyQualifiedAnnotationBeanNameGenerator.class,
        scanBasePackages = {
        "com.mybilibili.contentinteraction",
        "com.mybilibili.comment",
        "com.mybilibili.interaction",
        "com.mybilibili.danmaku",
        "com.mybilibili.common",
        "com.mybilibili.mq"
})
@EnableFeignClients(clients = {
        com.mybilibili.comment.feign.ContentReviewClient.class,
        com.mybilibili.comment.feign.ManuscriptClient.class,
        com.mybilibili.comment.feign.MessageClient.class,
        com.mybilibili.comment.feign.UserClient.class,
        com.mybilibili.danmaku.feign.VideoClient.class,
        com.mybilibili.interaction.feign.MessageClient.class,
        com.mybilibili.interaction.feign.VideoClient.class
})
@MapperScan({
        "com.mybilibili.comment.mapper",
        "com.mybilibili.interaction.mapper",
        "com.mybilibili.danmaku.mapper"
})
@EnableMongoRepositories(basePackages = {
        "com.mybilibili.interaction.repository",
        "com.mybilibili.danmaku.repository"
})
public class ContentInteractionApplication {

    public static void main(String[] args) {
        SpringApplication.run(ContentInteractionApplication.class, args);
    }
}
