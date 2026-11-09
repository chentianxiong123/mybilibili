package com.mybilibili.web.service;

import com.mybilibili.common.dto.CreatorSettingsDTO;
import com.mybilibili.common.entity.CreatorSettings;
import com.mybilibili.common.vo.CreatorSettingsVO;

public interface CreatorSettingsService {

    CreatorSettings createDefaultSettings(Integer userId);

    CreatorSettingsVO getSettingsByUserId(Integer userId);

    void updateSettings(Integer userId, CreatorSettingsDTO dto);

    CreatorSettings getOrCreateSettings(Integer userId);
}
