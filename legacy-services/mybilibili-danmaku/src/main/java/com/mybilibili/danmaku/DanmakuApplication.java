package com.mybilibili.danmaku;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.context.annotation.ComponentScan;

@SpringBootApplication
@EnableDiscoveryClient
@EnableFeignClients
@ComponentScan(basePackages = {"com.mybilibili.danmaku", "com.mybilibili.mq"})
public class DanmakuApplication {
    public static void main(String[] args) {
        SpringApplication.run(DanmakuApplication.class, args);
    }
}
