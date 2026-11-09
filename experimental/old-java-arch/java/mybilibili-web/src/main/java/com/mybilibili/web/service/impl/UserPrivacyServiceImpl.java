package com.mybilibili.web.service.impl;

import com.mybilibili.common.dto.UserPrivacySettingsDTO;
import com.mybilibili.common.entity.UserPrivacySettings;
import com.mybilibili.common.entity.UserTag;
import com.mybilibili.common.vo.UserPrivacySettingsVO;
import com.mybilibili.web.mapper.UserPrivacySettingsMapper;
import com.mybilibili.web.mapper.UserTagMapper;
import com.mybilibili.web.service.UserPrivacyService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;

@Service
public class UserPrivacyServiceImpl implements UserPrivacyService {

    @Autowired
    private UserPrivacySettingsMapper privacySettingsMapper;

    @Autowired
    private UserTagMapper userTagMapper;

    @Override
    public UserPrivacySettingsVO getPrivacySettings(Integer userId) {
        UserPrivacySettings settings = privacySettingsMapper.findByUserId(userId);

        UserPrivacySettingsVO vo = new UserPrivacySettingsVO();
        if (settings != null) {
            vo.setPublicCollection(settings.getPublicCollection());
            vo.setPublicBirthdayTags(settings.getPublicBirthdayTags());
            vo.setPublicCoinVideos(settings.getPublicCoinVideos());
            vo.setPublicLikeVideos(settings.getPublicLikeVideos());
            vo.setPublicFollowingList(settings.getPublicFollowingList());
            vo.setPublicFollowersList(settings.getPublicFollowersList());
        } else {
            // 默认设置
            vo.setPublicCollection(true);
            vo.setPublicBirthdayTags(false);
            vo.setPublicCoinVideos(false);
            vo.setPublicLikeVideos(false);
            vo.setPublicFollowingList(false);
            vo.setPublicFollowersList(false);
        }

        // 获取用户标签
        List<String> tags = getUserTags(userId);
        vo.setTags(tags);

        return vo;
    }

    @Override
    public void updatePrivacySettings(Integer userId, UserPrivacySettingsDTO dto) {
        UserPrivacySettings settings = privacySettingsMapper.findByUserId(userId);

        if (settings == null) {
            // 创建新设置
            settings = new UserPrivacySettings();
            settings.setUserId(userId);
            settings.setPublicCollection(dto.getPublicCollection() != null ? dto.getPublicCollection() : true);
            settings.setPublicBirthdayTags(dto.getPublicBirthdayTags() != null ? dto.getPublicBirthdayTags() : false);
            settings.setPublicCoinVideos(dto.getPublicCoinVideos() != null ? dto.getPublicCoinVideos() : false);
            settings.setPublicLikeVideos(dto.getPublicLikeVideos() != null ? dto.getPublicLikeVideos() : false);
            settings.setPublicFollowingList(dto.getPublicFollowingList() != null ? dto.getPublicFollowingList() : false);
            settings.setPublicFollowersList(dto.getPublicFollowersList() != null ? dto.getPublicFollowersList() : false);
            privacySettingsMapper.insert(settings);
        } else {
            // 更新设置
            if (dto.getPublicCollection() != null) {
                settings.setPublicCollection(dto.getPublicCollection());
            }
            if (dto.getPublicBirthdayTags() != null) {
                settings.setPublicBirthdayTags(dto.getPublicBirthdayTags());
            }
            if (dto.getPublicCoinVideos() != null) {
                settings.setPublicCoinVideos(dto.getPublicCoinVideos());
            }
            if (dto.getPublicLikeVideos() != null) {
                settings.setPublicLikeVideos(dto.getPublicLikeVideos());
            }
            if (dto.getPublicFollowingList() != null) {
                settings.setPublicFollowingList(dto.getPublicFollowingList());
            }
            if (dto.getPublicFollowersList() != null) {
                settings.setPublicFollowersList(dto.getPublicFollowersList());
            }
            privacySettingsMapper.update(settings);
        }
    }

    @Override
    public List<String> getUserTags(Integer userId) {
        List<UserTag> tags = userTagMapper.findByUserId(userId);
        return tags.stream().map(UserTag::getTagName).collect(Collectors.toList());
    }

    @Override
    public void addUserTag(Integer userId, String tagName) {
        // 检查标签数量限制
        int count = userTagMapper.countByUserId(userId);
        if (count >= 10) {
            throw new RuntimeException("个人标签最多只能添加10个");
        }

        UserTag tag = new UserTag();
        tag.setUserId(userId);
        tag.setTagName(tagName);
        userTagMapper.insert(tag);
    }

    @Override
    public void removeUserTag(Integer userId, String tagName) {
        userTagMapper.deleteByUserIdAndTagName(userId, tagName);
    }
}
