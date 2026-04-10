package com.mybilibili.video.service;

import com.alibaba.fastjson.JSON;
import com.alibaba.fastjson.TypeReference;
import com.mybilibili.common.entity.BannerImage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.Date;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class BannerImageRedisService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String KEY_HOME = "banner:home";
    private static final String KEY_CATEGORY = "banner:category:";
    private static final String KEY_BACKGROUND = "banner:background";
    private static final String KEY_USER_PROFILE_BG = "banner:user:profile:background";

    public List<BannerImage> getHomeBanners() {
        String json = redisTemplate.opsForValue().get(KEY_HOME);
        if (json == null) {
            return new ArrayList<>();
        }
        List<BannerImage> banners = JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
        return filterValidBanners(banners);
    }

    public boolean addHomeBanner(BannerImage banner) {
        List<BannerImage> banners = getAllHomeBanners();
        banners.add(banner);
        redisTemplate.opsForValue().set(KEY_HOME, JSON.toJSONString(banners));
        return true;
    }

    public boolean updateHomeBanner(Integer id, BannerImage banner) {
        List<BannerImage> banners = getAllHomeBanners();
        for (int i = 0; i < banners.size(); i++) {
            if (banners.get(i).getId().equals(id)) {
                banner.setId(id);
                banners.set(i, banner);
                redisTemplate.opsForValue().set(KEY_HOME, JSON.toJSONString(banners));
                return true;
            }
        }
        return false;
    }

    public boolean deleteHomeBanner(Integer id) {
        List<BannerImage> banners = getAllHomeBanners();
        boolean removed = banners.removeIf(b -> b.getId().equals(id));
        if (removed) {
            redisTemplate.opsForValue().set(KEY_HOME, JSON.toJSONString(banners));
        }
        return removed;
    }

    private List<BannerImage> getAllHomeBanners() {
        String json = redisTemplate.opsForValue().get(KEY_HOME);
        if (json == null) {
            return new ArrayList<>();
        }
        return JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
    }

    public List<BannerImage> getCategoryBanners(Integer categoryId) {
        String json = redisTemplate.opsForValue().get(KEY_CATEGORY + categoryId);
        if (json == null) {
            return new ArrayList<>();
        }
        List<BannerImage> banners = JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
        return filterValidBanners(banners);
    }

    public boolean addCategoryBanner(Integer categoryId, BannerImage banner) {
        List<BannerImage> banners = getAllCategoryBanners(categoryId);
        banners.add(banner);
        redisTemplate.opsForValue().set(KEY_CATEGORY + categoryId, JSON.toJSONString(banners));
        return true;
    }

    public boolean updateCategoryBanner(Integer categoryId, Integer id, BannerImage banner) {
        List<BannerImage> banners = getAllCategoryBanners(categoryId);
        for (int i = 0; i < banners.size(); i++) {
            if (banners.get(i).getId().equals(id)) {
                banner.setId(id);
                banners.set(i, banner);
                redisTemplate.opsForValue().set(KEY_CATEGORY + categoryId, JSON.toJSONString(banners));
                return true;
            }
        }
        return false;
    }

    public boolean deleteCategoryBanner(Integer categoryId, Integer id) {
        List<BannerImage> banners = getAllCategoryBanners(categoryId);
        boolean removed = banners.removeIf(b -> b.getId().equals(id));
        if (removed) {
            redisTemplate.opsForValue().set(KEY_CATEGORY + categoryId, JSON.toJSONString(banners));
        }
        return removed;
    }

    private List<BannerImage> getAllCategoryBanners(Integer categoryId) {
        String json = redisTemplate.opsForValue().get(KEY_CATEGORY + categoryId);
        if (json == null) {
            return new ArrayList<>();
        }
        return JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
    }

    public BannerImage getBackgroundImage() {
        String json = redisTemplate.opsForValue().get(KEY_BACKGROUND);
        if (json == null) {
            return null;
        }
        BannerImage banner = JSON.parseObject(json, BannerImage.class);
        if (banner.getStatus() != null && banner.getStatus() == 1
                && isValidTime(banner.getStartTime(), banner.getEndTime())) {
            return banner;
        }
        return null;
    }

    public BannerImage getUserProfileBackground() {
        String json = redisTemplate.opsForValue().get(KEY_USER_PROFILE_BG);
        if (json == null) {
            return null;
        }
        BannerImage banner = JSON.parseObject(json, BannerImage.class);
        if (banner.getStatus() != null && banner.getStatus() == 1
                && isValidTime(banner.getStartTime(), banner.getEndTime())) {
            return banner;
        }
        return null;
    }

    public boolean saveBackgroundImage(BannerImage banner) {
        redisTemplate.opsForValue().set(KEY_BACKGROUND, JSON.toJSONString(banner));
        return true;
    }

    public boolean deleteBackgroundImage() {
        redisTemplate.delete(KEY_BACKGROUND);
        return true;
    }

    public boolean saveUserProfileBackground(BannerImage banner) {
        redisTemplate.opsForValue().set(KEY_USER_PROFILE_BG, JSON.toJSONString(banner));
        return true;
    }

    public boolean deleteUserProfileBackground() {
        redisTemplate.delete(KEY_USER_PROFILE_BG);
        return true;
    }

    private List<BannerImage> filterValidBanners(List<BannerImage> banners) {
        Date now = new Date();
        return banners.stream()
                .filter(b -> b.getStatus() != null && b.getStatus() == 1)
                .filter(b -> isValidTime(b.getStartTime(), b.getEndTime()))
                .sorted(Comparator.comparingInt(BannerImage::getSortOrder))
                .collect(Collectors.toList());
    }

    private boolean isValidTime(Date startTime, Date endTime) {
        Date now = new Date();
        if (startTime != null && now.before(startTime)) {
            return false;
        }
        if (endTime != null && now.after(endTime)) {
            return false;
        }
        return true;
    }
}
