package com.mybilibili.admin;

import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.data.elasticsearch.ElasticsearchRepositoriesAutoConfiguration;
import org.springframework.boot.autoconfigure.elasticsearch.ElasticsearchRestClientAutoConfiguration;

@SpringBootApplication(scanBasePackages = {"com.mybilibili.admin", "com.mybilibili.common", "com.mybilibili.web"},
        exclude = {
                ElasticsearchRestClientAutoConfiguration.class,
                ElasticsearchRepositoriesAutoConfiguration.class
        })
public class MybilibiliAdminRunApplication {
    public static void main(String[] args) {
        org.springframework.boot.SpringApplication.run(MybilibiliAdminRunApplication.class, args);
    }
}
