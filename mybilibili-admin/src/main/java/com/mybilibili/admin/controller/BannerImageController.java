package com.mybilibili.admin.controller;

import com.mybilibili.admin.dto.BannerImageDTO;
import com.mybilibili.admin.utils.UploadFilePathUtils;
import com.mybilibili.common.entity.BannerImage;
import com.mybilibili.common.service.BannerImageRedisService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/banner-images")
@Tag(name = "图片管理接口", description = "轮播图和背景图管理")
public class BannerImageController {

    @Autowired
    private BannerImageRedisService bannerImageService;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    // 最大文件大小：5MB
    private static final long MAX_FILE_SIZE = 5 * 1024 * 1024;

    // ==================== 首页轮播图管理 ====================

    @GetMapping("/home")
    @Operation(summary = "获取首页轮播图列表", description = "获取所有首页轮播图（管理端显示全部）")
    public Result<List<BannerImage>> getHomeBanners() {
        List<BannerImage> banners = bannerImageService.getHomeBanners();
        return Result.success(banners);
    }

    @PostMapping("/home")
    @Operation(summary = "添加首页轮播图", description = "添加首页轮播图")
    public Result<?> addHomeBanner(@RequestBody BannerImageDTO dto) {
        BannerImage banner = convertToEntity(dto);
        bannerImageService.addHomeBanner(banner);
        return Result.success("添加成功");
    }

    @PutMapping("/home/{id}")
    @Operation(summary = "更新首页轮播图", description = "更新首页轮播图")
    public Result<?> updateHomeBanner(@PathVariable Integer id, @RequestBody BannerImageDTO dto) {
        BannerImage banner = convertToEntity(dto);
        bannerImageService.updateHomeBanner(id, banner);
        return Result.success("更新成功");
    }

    @DeleteMapping("/home/{id}")
    @Operation(summary = "删除首页轮播图", description = "删除首页轮播图")
    public Result<?> deleteHomeBanner(@PathVariable Integer id) {
        bannerImageService.deleteHomeBanner(id);
        return Result.success("删除成功");
    }

    // ==================== 分类轮播图管理 ====================

    @GetMapping("/category/{categoryId}")
    @Operation(summary = "获取分类轮播图列表", description = "获取指定分类的轮播图")
    public Result<List<BannerImage>> getCategoryBanners(@PathVariable Integer categoryId) {
        List<BannerImage> banners = bannerImageService.getCategoryBanners(categoryId);
        return Result.success(banners);
    }

    @PostMapping("/category/{categoryId}")
    @Operation(summary = "添加分类轮播图", description = "添加分类轮播图")
    public Result<?> addCategoryBanner(@PathVariable Integer categoryId, @RequestBody BannerImageDTO dto) {
        BannerImage banner = convertToEntity(dto);
        bannerImageService.addCategoryBanner(categoryId, banner);
        return Result.success("添加成功");
    }

    @PutMapping("/category/{categoryId}/{id}")
    @Operation(summary = "更新分类轮播图", description = "更新分类轮播图")
    public Result<?> updateCategoryBanner(@PathVariable Integer categoryId, @PathVariable Integer id,
                                          @RequestBody BannerImageDTO dto) {
        BannerImage banner = convertToEntity(dto);
        bannerImageService.updateCategoryBanner(categoryId, id, banner);
        return Result.success("更新成功");
    }

    @DeleteMapping("/category/{categoryId}/{id}")
    @Operation(summary = "删除分类轮播图", description = "删除分类轮播图")
    public Result<?> deleteCategoryBanner(@PathVariable Integer categoryId, @PathVariable Integer id) {
        bannerImageService.deleteCategoryBanner(categoryId, id);
        return Result.success("删除成功");
    }

    // ==================== 顶部背景图管理 ====================

    @GetMapping("/background")
    @Operation(summary = "获取顶部背景图", description = "获取顶部背景图")
    public Result<BannerImage> getBackgroundImage() {
        BannerImage banner = bannerImageService.getBackgroundImage();
        return Result.success(banner);
    }

    @PostMapping("/background")
    @Operation(summary = "设置顶部背景图", description = "设置顶部背景图")
    public Result<?> saveBackgroundImage(@RequestBody BannerImageDTO dto) {
        BannerImage banner = convertToEntity(dto);
        bannerImageService.saveBackgroundImage(banner);
        return Result.success("设置成功");
    }

    @DeleteMapping("/background")
    @Operation(summary = "删除顶部背景图", description = "删除顶部背景图")
    public Result<?> deleteBackgroundImage() {
        bannerImageService.deleteBackgroundImage();
        return Result.success("删除成功");
    }

    // ==================== 用户主页背景图管理 ====================

    @GetMapping("/user-profile")
    @Operation(summary = "获取用户主页背景图", description = "获取用户主页背景图")
    public Result<BannerImage> getUserProfileBackground() {
        BannerImage banner = bannerImageService.getUserProfileBackground();
        return Result.success(banner);
    }

    @PostMapping("/user-profile")
    @Operation(summary = "设置用户主页背景图", description = "设置用户主页背景图")
    public Result<?> saveUserProfileBackground(@RequestBody BannerImageDTO dto) {
        BannerImage banner = convertToEntity(dto);
        bannerImageService.saveUserProfileBackground(banner);
        return Result.success("设置成功");
    }

    @DeleteMapping("/user-profile")
    @Operation(summary = "删除用户主页背景图", description = "删除用户主页背景图")
    public Result<?> deleteUserProfileBackground() {
        bannerImageService.deleteUserProfileBackground();
        return Result.success("删除成功");
    }

    // ==================== 图片上传 ====================

    @PostMapping("/upload")
    @Operation(summary = "上传图片", description = "上传轮播图或背景图")
    public Result<Map<String, String>> uploadImage(@RequestParam("file") MultipartFile file) {
        try {
            // 校验文件是否为空
            if (file == null || file.isEmpty()) {
                return Result.error("请选择要上传的图片");
            }

            // 校验文件大小
            if (file.getSize() > MAX_FILE_SIZE) {
                return Result.error("图片大小不能超过5MB");
            }

            // 校验文件类型
            String contentType = file.getContentType();
            if (!uploadFilePathUtils.isValidImageType(contentType)) {
                return Result.error("不支持的图片格式，仅支持jpg/jpeg/png/gif/webp");
            }

            // 校验文件扩展名
            String originalFilename = file.getOriginalFilename();
            if (!uploadFilePathUtils.isValidImageExtension(originalFilename)) {
                return Result.error("不支持的图片格式，仅支持jpg/jpeg/png/gif/webp");
            }

            // 确保图片目录存在
            uploadFilePathUtils.createImagesDirectory();

            // 生成唯一文件名
            String fileName = uploadFilePathUtils.generateImageFileName(originalFilename);

            // 获取完整保存路径
            String filePath = uploadFilePathUtils.getImagePath(fileName);

            // 保存文件
            File destFile = new File(filePath);
            file.transferTo(destFile);

            // 获取访问URL
            String imageUrl = uploadFilePathUtils.getImageUrl(fileName);

            // 返回结果
            Map<String, String> result = new HashMap<>();
            result.put("url", imageUrl);
            result.put("fileName", fileName);

            return Result.success("上传成功", result);

        } catch (IOException e) {
            return Result.error("图片上传失败：" + e.getMessage());
        } catch (Exception e) {
            return Result.error("上传失败：" + e.getMessage());
        }
    }

    // ==================== 私有方法 ====================

    private BannerImage convertToEntity(BannerImageDTO dto) {
        BannerImage banner = new BannerImage();
        banner.setTitle(dto.getTitle());
        banner.setImageUrl(dto.getImageUrl());
        banner.setLinkUrl(dto.getLinkUrl());
        banner.setSortOrder(dto.getSortOrder());
        banner.setStatus(dto.getStatus());
        banner.setStartTime(dto.getStartTime());
        banner.setEndTime(dto.getEndTime());
        return banner;
    }
}
