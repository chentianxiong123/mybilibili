package com.mybilibili.web.service;

import com.mybilibili.common.dto.UserPrivacySettingsDTO;
import com.mybilibili.common.vo.UserPrivacySettingsVO;

import java.util.List;

public interface UserPrivacyService {

    UserPrivacySettingsVO getPrivacySettings(Integer userId);

    void updatePrivacySettings(Integer userId, UserPrivacySettingsDTO dto);

    List<String> getUserTags(Integer userId);

    void addUserTag(Integer userId, String tagName);

    void removeUserTag(Integer userId, String tagName);
}
