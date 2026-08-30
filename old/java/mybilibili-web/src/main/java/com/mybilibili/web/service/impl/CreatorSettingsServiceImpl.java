package com.mybilibili.web.service.impl;

import com.mybilibili.common.dto.CreatorSettingsDTO;
import com.mybilibili.common.entity.CreatorSettings;
import com.mybilibili.common.vo.CreatorSettingsVO;
import com.mybilibili.web.mapper.CreatorSettingsMapper;
import com.mybilibili.web.service.CreatorSettingsService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class CreatorSettingsServiceImpl implements CreatorSettingsService {

    @Autowired
    private CreatorSettingsMapper creatorSettingsMapper;

    @Override
    public CreatorSettings createDefaultSettings(Integer userId) {
        CreatorSettings setting = new CreatorSettings();
        setting.setUserId(userId);
        setting.setDefaultCategoryId(null);
        setting.setAutoPublish(false);
        setting.setCommentNotify(true);
        setting.setLikeNotify(true);
        setting.setFollowNotify(true);
        creatorSettingsMapper.insert(setting);
        return setting;
    }

    @Override
    public CreatorSettingsVO getSettingsByUserId(Integer userId) {
        return creatorSettingsMapper.selectVOByUserId(userId);
    }

    @Override
    public void updateSettings(Integer userId, CreatorSettingsDTO dto) {
        CreatorSettings setting = creatorSettingsMapper.selectByUserId(userId);
        if (setting == null) {
            setting = createDefaultSettings(userId);
        }
        setting.setDefaultCategoryId(dto.getDefaultCategoryId());
        setting.setAutoPublish(dto.getAutoPublish());
        setting.setCommentNotify(dto.getCommentNotify());
        setting.setLikeNotify(dto.getLikeNotify());
        setting.setFollowNotify(dto.getFollowNotify());
        creatorSettingsMapper.update(setting);
    }

    @Override
    public CreatorSettings getOrCreateSettings(Integer userId) {
        CreatorSettings setting = creatorSettingsMapper.selectByUserId(userId);
        if (setting == null) {
            setting = createDefaultSettings(userId);
        }
        return setting;
    }
}
