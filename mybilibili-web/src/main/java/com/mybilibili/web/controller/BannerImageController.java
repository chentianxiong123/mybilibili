package com.mybilibili.web.controller;

import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.common.service.BannerImageRedisService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/banner-images")
@Tag(name = "轮播图查询接口", description = "前端展示用的轮播图查询")
public class BannerImageController {

    @Autowired
    private BannerImageRedisService bannerImageService;

    @GetMapping("/home")
    @Operation(summary = "获取首页轮播图", description = "获取启用的首页轮播图列表")
    public Result<List<BannerImage>> getHomeBanners() {
        return Result.success(bannerImageService.getHomeBanners());
    }

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "获取分类轮播图", description = "获取指定分类的轮播图列表")
    public Result<List<BannerImage>> getCategoryBanners(@PathVariable Integer categoryId) {
        return Result.success(bannerImageService.getCategoryBanners(categoryId));
    }

    @GetMapping("/background")
    @Operation(summary = "获取顶部背景图", description = "获取当前启用的顶部背景图")
    public Result<BannerImage> getBackgroundImage() {
        return Result.success(bannerImageService.getBackgroundImage());
    }

    @GetMapping("/user-profile")
    @Operation(summary = "获取用户主页背景图", description = "获取用户主页背景图")
    public Result<BannerImage> getUserProfileBackground() {
        return Result.success(bannerImageService.getUserProfileBackground());
    }
}
