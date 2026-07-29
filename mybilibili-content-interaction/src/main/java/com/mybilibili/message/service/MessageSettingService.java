package com.mybilibili.message.service;

import com.mybilibili.common.dto.MessageSettingDTO;
import com.mybilibili.common.entity.MessageSetting;
import com.mybilibili.common.vo.MessageSettingVO;

public interface MessageSettingService {

    MessageSetting createDefaultSettings(Integer userId);

    MessageSettingVO getSettingsByUserId(Integer userId);

    void updateSettings(Integer userId, MessageSettingDTO dto);

    MessageSetting getOrCreateSettings(Integer userId);
}