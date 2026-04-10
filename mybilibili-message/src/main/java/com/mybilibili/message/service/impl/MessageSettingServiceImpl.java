package com.mybilibili.message.service.impl;

import com.mybilibili.common.dto.MessageSettingDTO;
import com.mybilibili.common.entity.MessageSetting;
import com.mybilibili.common.vo.MessageSettingVO;
import com.mybilibili.message.mapper.MessageSettingMapper;
import com.mybilibili.message.service.MessageSettingService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class MessageSettingServiceImpl implements MessageSettingService {

    @Autowired
    private MessageSettingMapper messageSettingMapper;

    @Override
    public MessageSetting createDefaultSettings(Integer userId) {
        MessageSetting setting = new MessageSetting();
        setting.setUserId(userId);
        setting.setPrivateMessageNotify(true);
        setting.setReplyNotify(true);
        setting.setAtNotify(true);
        setting.setLikeNotify(true);
        setting.setSystemNotify(true);
        messageSettingMapper.insert(setting);
        return setting;
    }

    @Override
    public MessageSettingVO getSettingsByUserId(Integer userId) {
        return messageSettingMapper.selectVOByUserId(userId);
    }

    @Override
    public void updateSettings(Integer userId, MessageSettingDTO dto) {
        MessageSetting setting = messageSettingMapper.selectByUserId(userId);
        if (setting == null) {
            setting = createDefaultSettings(userId);
        }
        setting.setPrivateMessageNotify(dto.getPrivateMessageNotify());
        setting.setReplyNotify(dto.getReplyNotify());
        setting.setAtNotify(dto.getAtNotify());
        setting.setLikeNotify(dto.getLikeNotify());
        setting.setSystemNotify(dto.getSystemNotify());
        messageSettingMapper.update(setting);
    }

    @Override
    public MessageSetting getOrCreateSettings(Integer userId) {
        MessageSetting setting = messageSettingMapper.selectByUserId(userId);
        if (setting == null) {
            setting = createDefaultSettings(userId);
        }
        return setting;
    }
}