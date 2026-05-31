package com.mybilibili.interaction.service;

import com.mybilibili.interaction.entity.UserProfile;

import java.util.List;

public interface UserProfileService {

    UserProfile getOrCreateProfile(Integer userId);

    void recordWatch(Integer userId, Integer categoryId, List<String> tags, Integer durationSeconds);

    void recordLike(Integer userId, Integer categoryId, List<String> tags);

    void recordCollect(Integer userId, Integer categoryId, List<String> tags);
}
