package com.mybilibili.comment;

import com.mybilibili.common.feign.FeignRequestInterceptor;
import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.client.discovery.EnableDiscoveryClient;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Import;

@SpringBootApplication
@EnableDiscoveryClient
@EnableFeignClients(basePackages = {"com.mybilibili.comment.feign", "com.mybilibili.common.feign"})
@MapperScan("com.mybilibili.comment.mapper")
@ComponentScan({"com.mybilibili.comment", "com.mybilibili.mq"})
@Import(FeignRequestInterceptor.class)
public class CommentApplication {
    public static void main(String[] args) {
        SpringApplication.run(CommentApplication.class, args);
    }
}
