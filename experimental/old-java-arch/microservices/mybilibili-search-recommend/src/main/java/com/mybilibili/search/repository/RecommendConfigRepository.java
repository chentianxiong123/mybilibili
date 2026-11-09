package com.mybilibili.search.repository;

import com.mybilibili.search.entity.RecommendConfig;
import org.springframework.data.mongodb.repository.MongoRepository;

public interface RecommendConfigRepository extends MongoRepository<RecommendConfig, String> {
}
