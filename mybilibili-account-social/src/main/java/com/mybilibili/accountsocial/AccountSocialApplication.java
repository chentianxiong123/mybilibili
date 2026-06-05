package com.mybilibili.accountsocial;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.context.annotation.FullyQualifiedAnnotationBeanNameGenerator;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication(
        nameGenerator = FullyQualifiedAnnotationBeanNameGenerator.class,
        scanBasePackages = {
        "com.mybilibili.accountsocial",
        "com.mybilibili.user",
        "com.mybilibili.message",
        "com.mybilibili.common"
})
@EnableFeignClients(clients = {
        com.mybilibili.user.feign.ManuscriptClient.class,
        com.mybilibili.user.feign.UserProfileClient.class
})
@MapperScan({
        "com.mybilibili.user.mapper",
        "com.mybilibili.message.mapper"
})
@EnableScheduling
public class AccountSocialApplication {

    public static void main(String[] args) {
        SpringApplication.run(AccountSocialApplication.class, args);
    }
}
