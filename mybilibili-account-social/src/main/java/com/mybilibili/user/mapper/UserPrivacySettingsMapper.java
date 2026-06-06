package com.mybilibili.user.mapper;

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.mybilibili.common.entity.UserPrivacySettings;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

@Mapper
public interface UserPrivacySettingsMapper extends BaseMapper<UserPrivacySettings> {

    @Select("SELECT * FROM user_privacy_settings WHERE user_id = #{userId}")
    UserPrivacySettings selectByUserId(Integer userId);
}
