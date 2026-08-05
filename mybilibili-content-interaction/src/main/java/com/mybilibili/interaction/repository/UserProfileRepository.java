package com.mybilibili.interaction.repository;

import com.mybilibili.interaction.entity.UserProfile;
import org.springframework.data.mongodb.repository.MongoRepository;

import java.util.Optional;

public interface UserProfileRepository extends MongoRepository<UserProfile, String> {
    Optional<UserProfile> findByUserId(Integer userId);
}
