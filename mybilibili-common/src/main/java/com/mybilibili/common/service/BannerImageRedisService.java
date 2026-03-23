package com.mybilibili.common.service;

import com.mybilibili.common.entity.BannerImage;
import java.util.List;

public interface BannerImageRedisService {

    // ==================== 首页轮播图 ====================

    /**
     * 获取首页轮播图列表
     */
    List<BannerImage> getHomeBanners();

    /**
     * 保存首页轮播图列表
     */
    void saveHomeBanners(List<BannerImage> banners);

    /**
     * 添加首页轮播图
     */
    void addHomeBanner(BannerImage banner);

    /**
     * 更新首页轮播图
     */
    void updateHomeBanner(Integer id, BannerImage banner);

    /**
     * 删除首页轮播图
     */
    void deleteHomeBanner(Integer id);

    // ==================== 分类轮播图 ====================

    /**
     * 获取分类轮播图列表
     */
    List<BannerImage> getCategoryBanners(Integer categoryId);

    /**
     * 保存分类轮播图列表
     */
    void saveCategoryBanners(Integer categoryId, List<BannerImage> banners);

    /**
     * 添加分类轮播图
     */
    void addCategoryBanner(Integer categoryId, BannerImage banner);

    /**
     * 更新分类轮播图
     */
    void updateCategoryBanner(Integer categoryId, Integer id, BannerImage banner);

    /**
     * 删除分类轮播图
     */
    void deleteCategoryBanner(Integer categoryId, Integer id);

    // ==================== 顶部背景图 ====================

    /**
     * 获取顶部背景图
     */
    BannerImage getBackgroundImage();

    /**
     * 保存顶部背景图
     */
    void saveBackgroundImage(BannerImage banner);

    /**
     * 删除顶部背景图
     */
    void deleteBackgroundImage();

    // ==================== 用户主页背景图 ====================

    /**
     * 获取用户主页背景图
     */
    BannerImage getUserProfileBackground();

    /**
     * 保存用户主页背景图
     */
    void saveUserProfileBackground(BannerImage banner);

    /**
     * 删除用户主页背景图
     */
    void deleteUserProfileBackground();

    // ==================== 通用方法 ====================

    /**
     * 获取所有分类的轮播图（管理端用）
     */
    List<BannerImage> getAllCategoryBanners();

    /**
     * 清除所有缓存
     */
    void clearAll();
}
