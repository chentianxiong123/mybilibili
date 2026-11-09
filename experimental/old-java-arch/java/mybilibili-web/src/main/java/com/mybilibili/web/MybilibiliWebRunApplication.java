package com.mybilibili.web;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableAsync;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication(scanBasePackages = {"com.mybilibili.web", "com.mybilibili.common"})
@MapperScan("com.mybilibili.web.mapper")
@EnableAsync
@EnableScheduling
public class MybilibiliWebRunApplication {
    public static void main(String[] args) {
        org.springframework.boot.SpringApplication.run(MybilibiliWebRunApplication.class, args);
    }
}
