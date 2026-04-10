package com.mybilibili.content.service;

import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.content.mapper.BannerImageMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class BannerImageService {

    @Autowired
    private BannerImageMapper bannerImageMapper;

    public List<BannerImage> getHomeBanners() {
        return bannerImageMapper.selectByType(BannerImage.TYPE_HOME);
    }

    public List<BannerImage> getCategoryBanners(Integer categoryId) {
        return bannerImageMapper.selectByTypeAndCategory(BannerImage.TYPE_CATEGORY, categoryId);
    }

    public BannerImage getBackgroundImage() {
        List<BannerImage> list = bannerImageMapper.selectByType(BannerImage.TYPE_BACKGROUND);
        return list.isEmpty() ? null : list.get(0);
    }

    public BannerImage getUserProfileBackground() {
        List<BannerImage> list = bannerImageMapper.selectByType(BannerImage.TYPE_USER_PROFILE);
        return list.isEmpty() ? null : list.get(0);
    }
}
