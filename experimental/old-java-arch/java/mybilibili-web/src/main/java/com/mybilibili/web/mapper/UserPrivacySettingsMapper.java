package com.mybilibili.web.mapper;

import com.mybilibili.common.entity.UserPrivacySettings;
import org.apache.ibatis.annotations.*;

@Mapper
public interface UserPrivacySettingsMapper {

    @Select("SELECT * FROM user_privacy_settings WHERE user_id = #{userId}")
    UserPrivacySettings findByUserId(Integer userId);

    @Insert("INSERT INTO user_privacy_settings (user_id, public_collection, public_birthday_tags, " +
            "public_coin_videos, public_like_videos, public_following_list, public_followers_list) " +
            "VALUES (#{userId}, #{publicCollection}, #{publicBirthdayTags}, #{publicCoinVideos}, " +
            "#{publicLikeVideos}, #{publicFollowingList}, #{publicFollowersList})")
    @Options(useGeneratedKeys = true, keyProperty = "id")
    int insert(UserPrivacySettings settings);

    @Update("UPDATE user_privacy_settings SET " +
            "public_collection = #{publicCollection}, " +
            "public_birthday_tags = #{publicBirthdayTags}, " +
            "public_coin_videos = #{publicCoinVideos}, " +
            "public_like_videos = #{publicLikeVideos}, " +
            "public_following_list = #{publicFollowingList}, " +
            "public_followers_list = #{publicFollowersList} " +
            "WHERE user_id = #{userId}")
    int update(UserPrivacySettings settings);
}
