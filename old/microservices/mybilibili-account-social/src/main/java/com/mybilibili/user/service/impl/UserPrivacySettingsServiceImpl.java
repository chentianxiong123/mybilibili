package com.mybilibili.user.service.impl;

import com.mybilibili.common.dto.UserPrivacySettingsDTO;
import com.mybilibili.common.entity.UserPrivacySettings;
import com.mybilibili.common.entity.UserTag;
import com.mybilibili.common.vo.UserPrivacySettingsVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.user.mapper.UserPrivacySettingsMapper;
import com.mybilibili.user.mapper.UserTagMapper;
import com.mybilibili.user.service.UserPrivacySettingsService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class UserPrivacySettingsServiceImpl implements UserPrivacySettingsService {

    @Autowired
    private UserPrivacySettingsMapper privacySettingsMapper;

    @Autowired
    private UserTagMapper userTagMapper;

    @Override
    public Result<UserPrivacySettingsVO> getPrivacySettings(Integer userId) {
        UserPrivacySettings settings = privacySettingsMapper.selectByUserId(userId);
        
        if (settings == null) {
            settings = createDefaultSettings(userId);
            privacySettingsMapper.insert(settings);
        }
        
        List<UserTag> userTags = userTagMapper.selectByUserId(userId);
        List<String> tags = userTags.stream()
                .map(UserTag::getTagName)
                .collect(Collectors.toList());
        
        UserPrivacySettingsVO vo = new UserPrivacySettingsVO();
        vo.setPublicCollection(settings.getPublicCollection());
        vo.setPublicBirthdayTags(settings.getPublicBirthdayTags());
        vo.setPublicCoinVideos(settings.getPublicCoinVideos());
        vo.setPublicLikeVideos(settings.getPublicLikeVideos());
        vo.setPublicFollowingList(settings.getPublicFollowingList());
        vo.setPublicFollowersList(settings.getPublicFollowersList());
        vo.setTags(tags);
        
        return Result.success("获取成功", vo);
    }

    @Override
    @Transactional
    public Result<?> updatePrivacySettings(Integer userId, UserPrivacySettingsDTO dto) {
        UserPrivacySettings settings = privacySettingsMapper.selectByUserId(userId);
        
        if (settings == null) {
            settings = createDefaultSettings(userId);
            settings.setPublicCollection(dto.getPublicCollection());
            settings.setPublicBirthdayTags(dto.getPublicBirthdayTags());
            settings.setPublicCoinVideos(dto.getPublicCoinVideos());
            settings.setPublicLikeVideos(dto.getPublicLikeVideos());
            settings.setPublicFollowingList(dto.getPublicFollowingList());
            settings.setPublicFollowersList(dto.getPublicFollowersList());
            privacySettingsMapper.insert(settings);
        } else {
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
            privacySettingsMapper.updateById(settings);
        }
        
        return Result.success("更新成功");
    }

    @Override
    public Result<List<String>> getUserTags(Integer userId) {
        List<UserTag> userTags = userTagMapper.selectByUserId(userId);
        List<String> tags = userTags.stream()
                .map(UserTag::getTagName)
                .collect(Collectors.toList());
        return Result.success("获取成功", tags);
    }

    @Override
    @Transactional
    public Result<?> addUserTag(Integer userId, String tagName) {
        if (tagName == null || tagName.trim().isEmpty()) {
            return Result.error("标签名不能为空");
        }
        
        tagName = tagName.trim();
        
        UserTag existingTag = userTagMapper.selectByUserIdAndTagName(userId, tagName);
        if (existingTag != null) {
            return Result.error("该标签已存在");
        }
        
        List<UserTag> currentTags = userTagMapper.selectByUserId(userId);
        if (currentTags.size() >= 10) {
            return Result.error("最多只能添加10个标签");
        }
        
        UserTag userTag = new UserTag();
        userTag.setUserId(userId);
        userTag.setTagName(tagName);
        userTagMapper.insert(userTag);
        
        return Result.success("添加成功");
    }

    @Override
    @Transactional
    public Result<?> removeUserTag(Integer userId, String tagName) {
        if (tagName == null || tagName.trim().isEmpty()) {
            return Result.error("标签名不能为空");
        }
        
        int deleted = userTagMapper.deleteByUserIdAndTagName(userId, tagName.trim());
        if (deleted == 0) {
            return Result.error("标签不存在");
        }
        
        return Result.success("删除成功");
    }

    private UserPrivacySettings createDefaultSettings(Integer userId) {
        UserPrivacySettings settings = new UserPrivacySettings();
        settings.setUserId(userId);
        settings.setPublicCollection(true);
        settings.setPublicBirthdayTags(false);
        settings.setPublicCoinVideos(false);
        settings.setPublicLikeVideos(false);
        settings.setPublicFollowingList(false);
        settings.setPublicFollowersList(false);
        return settings;
    }
}
