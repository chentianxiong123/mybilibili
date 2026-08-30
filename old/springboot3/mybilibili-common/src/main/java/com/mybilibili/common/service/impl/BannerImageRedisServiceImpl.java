package com.mybilibili.common.service.impl;

import com.alibaba.fastjson2.JSON;
import com.alibaba.fastjson2.TypeReference;
import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.common.service.BannerImageRedisService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.Date;
import java.util.List;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.Collectors;

@Service
public class BannerImageRedisServiceImpl implements BannerImageRedisService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String KEY_HOME = "banner:home";
    private static final String KEY_CATEGORY = "banner:category:";
    private static final String KEY_BACKGROUND = "banner:background";
    private static final String KEY_USER_PROFILE_BG = "banner:user:profile:background";

    private final AtomicInteger idGenerator = new AtomicInteger(1);

    // ==================== 首页轮播图 ====================

    @Override
    public List<BannerImage> getHomeBanners() {
        String json = redisTemplate.opsForValue().get(KEY_HOME);
        if (json == null) {
            return new ArrayList<>();
        }
        List<BannerImage> banners = JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
        // 过滤已启用且在有效期内的
        return filterValidBanners(banners);
    }

    @Override
    public void saveHomeBanners(List<BannerImage> banners) {
        String json = JSON.toJSONString(banners);
        redisTemplate.opsForValue().set(KEY_HOME, json);
    }

    @Override
    public void addHomeBanner(BannerImage banner) {
        List<BannerImage> banners = getAllHomeBanners();
        banner.setId(idGenerator.getAndIncrement());
        banner.setStatus(1);
        banners.add(banner);
        // 按排序号排序
        banners.sort(Comparator.comparingInt(BannerImage::getSortOrder));
        saveHomeBanners(banners);
    }

    @Override
    public void updateHomeBanner(Integer id, BannerImage banner) {
        List<BannerImage> banners = getAllHomeBanners();
        for (int i = 0; i < banners.size(); i++) {
            if (banners.get(i).getId().equals(id)) {
                banner.setId(id);
                banners.set(i, banner);
                break;
            }
        }
        banners.sort(Comparator.comparingInt(BannerImage::getSortOrder));
        saveHomeBanners(banners);
    }

    @Override
    public void deleteHomeBanner(Integer id) {
        List<BannerImage> banners = getAllHomeBanners();
        banners.removeIf(b -> b.getId().equals(id));
        saveHomeBanners(banners);
    }

    // ==================== 分类轮播图 ====================

    @Override
    public List<BannerImage> getCategoryBanners(Integer categoryId) {
        String json = redisTemplate.opsForValue().get(KEY_CATEGORY + categoryId);
        if (json == null) {
            return new ArrayList<>();
        }
        List<BannerImage> banners = JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
        return filterValidBanners(banners);
    }

    @Override
    public void saveCategoryBanners(Integer categoryId, List<BannerImage> banners) {
        String json = JSON.toJSONString(banners);
        redisTemplate.opsForValue().set(KEY_CATEGORY + categoryId, json);
    }

    @Override
    public void addCategoryBanner(Integer categoryId, BannerImage banner) {
        List<BannerImage> banners = getAllCategoryBannersById(categoryId);
        banner.setId(idGenerator.getAndIncrement());
        banner.setStatus(1);
        banners.add(banner);
        banners.sort(Comparator.comparingInt(BannerImage::getSortOrder));
        saveCategoryBanners(categoryId, banners);
    }

    @Override
    public void updateCategoryBanner(Integer categoryId, Integer id, BannerImage banner) {
        List<BannerImage> banners = getAllCategoryBannersById(categoryId);
        for (int i = 0; i < banners.size(); i++) {
            if (banners.get(i).getId().equals(id)) {
                banner.setId(id);
                banners.set(i, banner);
                break;
            }
        }
        banners.sort(Comparator.comparingInt(BannerImage::getSortOrder));
        saveCategoryBanners(categoryId, banners);
    }

    @Override
    public void deleteCategoryBanner(Integer categoryId, Integer id) {
        List<BannerImage> banners = getAllCategoryBannersById(categoryId);
        banners.removeIf(b -> b.getId().equals(id));
        saveCategoryBanners(categoryId, banners);
    }

    // ==================== 顶部背景图 ====================

    @Override
    public BannerImage getBackgroundImage() {
        String json = redisTemplate.opsForValue().get(KEY_BACKGROUND);
        if (json == null) {
            return null;
        }
        BannerImage banner = JSON.parseObject(json, BannerImage.class);
        // 检查是否有效
        if (banner.getStatus() != null && banner.getStatus() == 1
                && isValidTime(banner.getStartTime(), banner.getEndTime())) {
            return banner;
        }
        return null;
    }

    @Override
    public void saveBackgroundImage(BannerImage banner) {
        if (banner.getId() == null) {
            banner.setId(idGenerator.getAndIncrement());
        }
        String json = JSON.toJSONString(banner);
        redisTemplate.opsForValue().set(KEY_BACKGROUND, json);
    }

    @Override
    public void deleteBackgroundImage() {
        redisTemplate.delete(KEY_BACKGROUND);
    }

    // ==================== 用户主页背景图 ====================

    @Override
    public BannerImage getUserProfileBackground() {
        String json = redisTemplate.opsForValue().get(KEY_USER_PROFILE_BG);
        if (json == null) {
            return null;
        }
        BannerImage banner = JSON.parseObject(json, BannerImage.class);
        // 检查是否有效
        if (banner.getStatus() != null && banner.getStatus() == 1
                && isValidTime(banner.getStartTime(), banner.getEndTime())) {
            return banner;
        }
        return null;
    }

    @Override
    public void saveUserProfileBackground(BannerImage banner) {
        if (banner.getId() == null) {
            banner.setId(idGenerator.getAndIncrement());
        }
        String json = JSON.toJSONString(banner);
        redisTemplate.opsForValue().set(KEY_USER_PROFILE_BG, json);
    }

    @Override
    public void deleteUserProfileBackground() {
        redisTemplate.delete(KEY_USER_PROFILE_BG);
    }

    // ==================== 通用方法 ====================

    @Override
    public List<BannerImage> getAllCategoryBanners() {
        // 获取所有分类轮播图的key
        List<String> keys = new ArrayList<>(redisTemplate.keys(KEY_CATEGORY + "*"));
        List<BannerImage> allBanners = new ArrayList<>();
        for (String key : keys) {
            String json = redisTemplate.opsForValue().get(key);
            if (json != null) {
                List<BannerImage> banners = JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
                allBanners.addAll(banners);
            }
        }
        return allBanners;
    }

    @Override
    public void clearAll() {
        redisTemplate.delete(KEY_HOME);
        redisTemplate.delete(KEY_BACKGROUND);
        redisTemplate.keys(KEY_CATEGORY + "*").forEach(redisTemplate::delete);
    }

    // ==================== 私有方法 ====================

    /**
     * 获取所有首页轮播图（不过滤状态）
     */
    private List<BannerImage> getAllHomeBanners() {
        String json = redisTemplate.opsForValue().get(KEY_HOME);
        if (json == null) {
            return new ArrayList<>();
        }
        return JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
    }

    /**
     * 获取指定分类的所有轮播图（不过滤状态）
     */
    private List<BannerImage> getAllCategoryBannersById(Integer categoryId) {
        String json = redisTemplate.opsForValue().get(KEY_CATEGORY + categoryId);
        if (json == null) {
            return new ArrayList<>();
        }
        return JSON.parseObject(json, new TypeReference<List<BannerImage>>() {});
    }

    /**
     * 过滤有效的轮播图（已启用且在有效期内）
     */
    private List<BannerImage> filterValidBanners(List<BannerImage> banners) {
        Date now = new Date();
        return banners.stream()
                .filter(b -> b.getStatus() != null && b.getStatus() == 1)
                .filter(b -> isValidTime(b.getStartTime(), b.getEndTime()))
                .sorted(Comparator.comparingInt(BannerImage::getSortOrder))
                .collect(Collectors.toList());
    }

    /**
     * 检查当前时间是否在有效期内
     */
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
