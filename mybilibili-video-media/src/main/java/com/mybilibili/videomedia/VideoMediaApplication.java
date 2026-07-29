package com.mybilibili.videomedia;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.context.annotation.FullyQualifiedAnnotationBeanNameGenerator;
import org.springframework.data.mongodb.repository.config.EnableMongoRepositories;

@SpringBootApplication(
        nameGenerator = FullyQualifiedAnnotationBeanNameGenerator.class,
        scanBasePackages = {
        "com.mybilibili.videomedia",
        "com.mybilibili.video",
        "com.mybilibili.live",
        "com.mybilibili.common",
        "com.mybilibili.mq"
})
@EnableFeignClients(basePackages = {
        "com.mybilibili.video.feign",
        "com.mybilibili.live.feign",
        "com.mybilibili.common.feign"
})
@MapperScan({
        "com.mybilibili.video.mapper",
        "com.mybilibili.live.mapper"
})
@EnableMongoRepositories(basePackages = {
        "com.mybilibili.video.repository"
})
public class VideoMediaApplication {

    public static void main(String[] args) {
        SpringApplication.run(VideoMediaApplication.class, args);
    }
}
