package com.mybilibili.video.controller;

import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.common.utils.UploadFilePathUtils;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.BannerImageRedisService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/banner-images")
@Tag(name = "轮播图接口", description = "首页轮播图、背景图接口")
public class BannerImageController {

    @Autowired
    private BannerImageRedisService bannerImageService;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @GetMapping("/home")
    @Operation(summary = "获取首页轮播图", description = "获取首页轮播图列表")
    public Result<List<BannerImage>> getHomeBanners() {
        List<BannerImage> banners = bannerImageService.getHomeBanners();
        return Result.success(banners);
    }

    @PostMapping("/home")
    @Operation(summary = "添加首页轮播图", description = "添加首页轮播图")
    public Result<?> addHomeBanner(@RequestBody BannerImage banner) {
        bannerImageService.addHomeBanner(banner);
        return Result.success("添加成功", null);
    }

    @PutMapping("/home/{id}")
    @Operation(summary = "更新首页轮播图", description = "更新首页轮播图")
    public Result<?> updateHomeBanner(@PathVariable Integer id, @RequestBody BannerImage banner) {
        boolean success = bannerImageService.updateHomeBanner(id, banner);
        if (success) {
            return Result.success("更新成功", null);
        }
        return Result.error("更新失败，轮播图不存在");
    }

    @DeleteMapping("/home/{id}")
    @Operation(summary = "删除首页轮播图", description = "删除首页轮播图")
    public Result<?> deleteHomeBanner(@PathVariable Integer id) {
        boolean success = bannerImageService.deleteHomeBanner(id);
        if (success) {
            return Result.success("删除成功", null);
        }
        return Result.error("删除失败，轮播图不存在");
    }

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "获取分类轮播图", description = "获取指定分类的轮播图")
    public Result<List<BannerImage>> getCategoryBanners(@PathVariable Integer categoryId) {
        List<BannerImage> banners = bannerImageService.getCategoryBanners(categoryId);
        return Result.success(banners);
    }

    @PostMapping("/category/{categoryId}")
    @Operation(summary = "添加分类轮播图", description = "添加分类轮播图")
    public Result<?> addCategoryBanner(@PathVariable Integer categoryId, @RequestBody BannerImage banner) {
        bannerImageService.addCategoryBanner(categoryId, banner);
        return Result.success("添加成功", null);
    }

    @PutMapping("/category/{categoryId}/{id}")
    @Operation(summary = "更新分类轮播图", description = "更新分类轮播图")
    public Result<?> updateCategoryBanner(@PathVariable Integer categoryId, @PathVariable Integer id, @RequestBody BannerImage banner) {
        boolean success = bannerImageService.updateCategoryBanner(categoryId, id, banner);
        if (success) {
            return Result.success("更新成功", null);
        }
        return Result.error("更新失败，轮播图不存在");
    }

    @DeleteMapping("/category/{categoryId}/{id}")
    @Operation(summary = "删除分类轮播图", description = "删除分类轮播图")
    public Result<?> deleteCategoryBanner(@PathVariable Integer categoryId, @PathVariable Integer id) {
        boolean success = bannerImageService.deleteCategoryBanner(categoryId, id);
        if (success) {
            return Result.success("删除成功", null);
        }
        return Result.error("删除失败，轮播图不存在");
    }

    @GetMapping("/background")
    @Operation(summary = "获取顶部背景图", description = "获取顶部背景图")
    public Result<BannerImage> getBackgroundImage() {
        BannerImage banner = bannerImageService.getBackgroundImage();
        return Result.success(banner);
    }

    @PostMapping("/background")
    @Operation(summary = "设置顶部背景图", description = "设置顶部背景图")
    public Result<?> saveBackgroundImage(@RequestBody BannerImage banner) {
        bannerImageService.saveBackgroundImage(banner);
        return Result.success("保存成功", null);
    }

    @DeleteMapping("/background")
    @Operation(summary = "删除顶部背景图", description = "删除顶部背景图")
    public Result<?> deleteBackgroundImage() {
        bannerImageService.deleteBackgroundImage();
        return Result.success("删除成功", null);
    }

    @GetMapping("/user-profile")
    @Operation(summary = "获取用户主页背景图", description = "获取用户主页背景图")
    public Result<BannerImage> getUserProfileBackground() {
        BannerImage banner = bannerImageService.getUserProfileBackground();
        return Result.success(banner);
    }

    @PostMapping("/user-profile")
    @Operation(summary = "设置用户主页背景图", description = "设置用户主页背景图")
    public Result<?> saveUserProfileBackground(@RequestBody BannerImage banner) {
        bannerImageService.saveUserProfileBackground(banner);
        return Result.success("保存成功", null);
    }

    @DeleteMapping("/user-profile")
    @Operation(summary = "删除用户主页背景图", description = "删除用户主页背景图")
    public Result<?> deleteUserProfileBackground() {
        bannerImageService.deleteUserProfileBackground();
        return Result.success("删除成功", null);
    }

    @PostMapping("/upload")
    @Operation(summary = "上传轮播图图片", description = "上传轮播图图片到服务器")
    public Result<Map<String, String>> uploadImage(@RequestParam("file") MultipartFile file) {
        try {
            if (file.isEmpty()) {
                return Result.error("文件不能为空");
            }

            if (!uploadFilePathUtils.isValidImageType(file.getContentType())) {
                return Result.error("图片类型不支持");
            }

            if (!uploadFilePathUtils.isValidImageExtension(file.getOriginalFilename())) {
                return Result.error("图片格式不支持");
            }

            uploadFilePathUtils.createImagesDirectory();

            String fileName = uploadFilePathUtils.generateImageFileName(file.getOriginalFilename());
            String filePath = uploadFilePathUtils.getImagePath(fileName);

            java.io.File destFile = new java.io.File(filePath);
            file.transferTo(destFile);

            String imageUrl = uploadFilePathUtils.getImageUrl(fileName);

            Map<String, String> result = new HashMap<>();
            result.put("url", imageUrl);

            return Result.success("上传成功", result);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error("上传失败: " + e.getMessage());
        }
    }
}
