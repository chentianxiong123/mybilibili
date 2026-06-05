package com.mybilibili.searchrecommend;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.cloud.openfeign.EnableFeignClients;
import org.springframework.context.annotation.FullyQualifiedAnnotationBeanNameGenerator;
import org.springframework.data.elasticsearch.repository.config.EnableElasticsearchRepositories;
import org.springframework.data.mongodb.repository.config.EnableMongoRepositories;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication(
        nameGenerator = FullyQualifiedAnnotationBeanNameGenerator.class,
        scanBasePackages = {
        "com.mybilibili.search",
        "com.mybilibili.analytics",
        "com.mybilibili.common",
        "com.mybilibili.mq"
})
@EnableFeignClients(basePackages = {
        "com.mybilibili.search.feign",
        "com.mybilibili.analytics.feign",
        "com.mybilibili.common.feign"
})
@MapperScan({
        "com.mybilibili.search.mapper",
        "com.mybilibili.analytics.mapper"
})
@EnableMongoRepositories(basePackages = "com.mybilibili.search.repository")
@EnableElasticsearchRepositories(basePackages = "com.mybilibili.search.repository")
@EnableScheduling
public class SearchRecommendApplication {

    public static void main(String[] args) {
        SpringApplication.run(SearchRecommendApplication.class, args);
    }
}
