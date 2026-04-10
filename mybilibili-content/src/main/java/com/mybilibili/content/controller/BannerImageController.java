package com.mybilibili.content.controller;

import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.common.vo.Result;
import com.mybilibili.content.service.BannerImageService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/banner-images")
@Tag(name = "轮播图接口", description = "首页轮播图、背景图接口")
public class BannerImageController {

    @Autowired
    private BannerImageService bannerImageService;

    @GetMapping("/home")
    @Operation(summary = "获取首页轮播图", description = "获取首页轮播图列表")
    public Result<List<BannerImage>> getHomeBanners() {
        List<BannerImage> banners = bannerImageService.getHomeBanners();
        return Result.success(banners);
    }

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "获取分类轮播图", description = "获取指定分类的轮播图")
    public Result<List<BannerImage>> getCategoryBanners(@PathVariable Integer categoryId) {
        List<BannerImage> banners = bannerImageService.getCategoryBanners(categoryId);
        return Result.success(banners);
    }

    @GetMapping("/background")
    @Operation(summary = "获取顶部背景图", description = "获取顶部背景图")
    public Result<BannerImage> getBackgroundImage() {
        BannerImage banner = bannerImageService.getBackgroundImage();
        return Result.success(banner);
    }

    @GetMapping("/user-profile")
    @Operation(summary = "获取用户主页背景图", description = "获取用户主页背景图")
    public Result<BannerImage> getUserProfileBackground() {
        BannerImage banner = bannerImageService.getUserProfileBackground();
        return Result.success(banner);
    }
}
