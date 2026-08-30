package com.mybilibili.user.service;

import com.mybilibili.common.dto.UserPrivacySettingsDTO;
import com.mybilibili.common.vo.UserPrivacySettingsVO;
import com.mybilibili.common.vo.Result;
import java.util.List;

public interface UserPrivacySettingsService {
    Result<UserPrivacySettingsVO> getPrivacySettings(Integer userId);
    Result<?> updatePrivacySettings(Integer userId, UserPrivacySettingsDTO dto);
    Result<List<String>> getUserTags(Integer userId);
    Result<?> addUserTag(Integer userId, String tagName);
    Result<?> removeUserTag(Integer userId, String tagName);
}
