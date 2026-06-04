package com.mybilibili.video.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.video.mapper.BannerImageMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.Date;
import java.util.List;

@Service
public class BannerImageService {

    @Autowired
    private BannerImageMapper bannerImageMapper;

    public List<BannerImage> getHomeBanners() {
        return listActiveByType(BannerImage.TYPE_HOME, null);
    }

    public boolean addHomeBanner(BannerImage banner) {
        banner.setId(null);
        banner.setType(BannerImage.TYPE_HOME);
        banner.setCategoryId(null);
        normalizeDefaults(banner);
        return bannerImageMapper.insert(banner) > 0;
    }

    public boolean updateHomeBanner(Integer id, BannerImage banner) {
        BannerImage existing = bannerImageMapper.selectById(id);
        if (existing == null || !BannerImage.TYPE_HOME.equals(existing.getType())) {
            return false;
        }
        banner.setId(id);
        banner.setType(BannerImage.TYPE_HOME);
        banner.setCategoryId(null);
        normalizeDefaults(banner);
        return bannerImageMapper.updateById(banner) > 0;
    }

    public boolean deleteHomeBanner(Integer id) {
        BannerImage existing = bannerImageMapper.selectById(id);
        if (existing == null || !BannerImage.TYPE_HOME.equals(existing.getType())) {
            return false;
        }
        return bannerImageMapper.deleteById(id) > 0;
    }

    public List<BannerImage> getCategoryBanners(Integer categoryId) {
        return listActiveByType(BannerImage.TYPE_CATEGORY, categoryId);
    }

    public boolean addCategoryBanner(Integer categoryId, BannerImage banner) {
        banner.setId(null);
        banner.setType(BannerImage.TYPE_CATEGORY);
        banner.setCategoryId(categoryId);
        normalizeDefaults(banner);
        return bannerImageMapper.insert(banner) > 0;
    }

    public boolean updateCategoryBanner(Integer categoryId, Integer id, BannerImage banner) {
        BannerImage existing = bannerImageMapper.selectById(id);
        if (existing == null
                || !BannerImage.TYPE_CATEGORY.equals(existing.getType())
                || !categoryId.equals(existing.getCategoryId())) {
            return false;
        }
        banner.setId(id);
        banner.setType(BannerImage.TYPE_CATEGORY);
        banner.setCategoryId(categoryId);
        normalizeDefaults(banner);
        return bannerImageMapper.updateById(banner) > 0;
    }

    public boolean deleteCategoryBanner(Integer categoryId, Integer id) {
        BannerImage existing = bannerImageMapper.selectById(id);
        if (existing == null
                || !BannerImage.TYPE_CATEGORY.equals(existing.getType())
                || !categoryId.equals(existing.getCategoryId())) {
            return false;
        }
        return bannerImageMapper.deleteById(id) > 0;
    }

    public BannerImage getBackgroundImage() {
        return getActiveSingle(BannerImage.TYPE_BACKGROUND);
    }

    public BannerImage getUserProfileBackground() {
        return getActiveSingle(BannerImage.TYPE_USER_PROFILE);
    }

    public boolean saveBackgroundImage(BannerImage banner) {
        return replaceSingle(BannerImage.TYPE_BACKGROUND, banner);
    }

    public boolean deleteBackgroundImage() {
        return deleteByType(BannerImage.TYPE_BACKGROUND);
    }

    public boolean saveUserProfileBackground(BannerImage banner) {
        return replaceSingle(BannerImage.TYPE_USER_PROFILE, banner);
    }

    public boolean deleteUserProfileBackground() {
        return deleteByType(BannerImage.TYPE_USER_PROFILE);
    }

    private List<BannerImage> listActiveByType(Integer type, Integer categoryId) {
        Date now = new Date();
        LambdaQueryWrapper<BannerImage> query = new LambdaQueryWrapper<BannerImage>()
                .eq(BannerImage::getType, type)
                .eq(BannerImage::getStatus, BannerImage.STATUS_ENABLED)
                .and(wrapper -> wrapper.isNull(BannerImage::getStartTime).or().le(BannerImage::getStartTime, now))
                .and(wrapper -> wrapper.isNull(BannerImage::getEndTime).or().ge(BannerImage::getEndTime, now))
                .orderByAsc(BannerImage::getSortOrder)
                .orderByDesc(BannerImage::getId);
        if (categoryId != null) {
            query.eq(BannerImage::getCategoryId, categoryId);
        }
        return bannerImageMapper.selectList(query);
    }

    private BannerImage getActiveSingle(Integer type) {
        List<BannerImage> banners = listActiveByType(type, null);
        return banners.isEmpty() ? null : banners.get(0);
    }

    private boolean replaceSingle(Integer type, BannerImage banner) {
        deleteByType(type);
        banner.setId(null);
        banner.setType(type);
        banner.setCategoryId(null);
        normalizeDefaults(banner);
        return bannerImageMapper.insert(banner) > 0;
    }

    private boolean deleteByType(Integer type) {
        LambdaQueryWrapper<BannerImage> query = new LambdaQueryWrapper<BannerImage>()
                .eq(BannerImage::getType, type);
        bannerImageMapper.delete(query);
        return true;
    }

    private void normalizeDefaults(BannerImage banner) {
        if (banner.getStatus() == null) {
            banner.setStatus(BannerImage.STATUS_ENABLED);
        }
        if (banner.getSortOrder() == null) {
            banner.setSortOrder(0);
        }
    }
}
