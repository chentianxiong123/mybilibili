package com.mybilibili.web.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.web.utils.UploadFilePathUtils;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.util.HashMap;
import java.util.Map;

/**
 * 图片上传控制器
 * 用于处理评论、动态、消息等场景的图片上传
 * 图片存储采用扁平化结构：/uploads/images/{timestamp}_{uuid}.{ext}
 */
@RestController
@RequestMapping("/image")
@Tag(name = "图片上传接口", description = "通用图片上传，用于评论、动态、消息等场景")
public class ImageController {

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    // 最大文件大小：5MB
    private static final long MAX_FILE_SIZE = 5 * 1024 * 1024;

    /**
     * 单图上传接口
     *
     * @param file 图片文件
     * @return 图片访问URL
     */
    @PostMapping("/upload")
    @Operation(summary = "上传单张图片", description = "上传通用图片，支持jpg/jpeg/png/gif/webp格式，最大5MB")
    @SecurityRequirement(name = "JWT")
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

    /**
     * 多图上传接口
     *
     * @param files 图片文件数组
     * @return 图片访问URL列表
     */
    @PostMapping("/upload/batch")
    @Operation(summary = "批量上传图片", description = "批量上传通用图片，最多9张，每张最大5MB")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> uploadImages(@RequestParam("files") MultipartFile[] files) {
        try {
            // 校验文件数组
            if (files == null || files.length == 0) {
                return Result.error("请选择要上传的图片");
            }

            // 限制最多9张
            if (files.length > 9) {
                return Result.error("一次最多上传9张图片");
            }

            // 确保图片目录存在
            uploadFilePathUtils.createImagesDirectory();

            Map<String, Object> result = new HashMap<>();
            java.util.List<Map<String, String>> successList = new java.util.ArrayList<>();
            java.util.List<Map<String, String>> failList = new java.util.ArrayList<>();

            for (int i = 0; i < files.length; i++) {
                MultipartFile file = files[i];
                Map<String, String> itemResult = new HashMap<>();
                itemResult.put("index", String.valueOf(i));
                itemResult.put("originalName", file.getOriginalFilename());

                try {
                    // 校验文件是否为空
                    if (file.isEmpty()) {
                        itemResult.put("error", "文件为空");
                        failList.add(itemResult);
                        continue;
                    }

                    // 校验文件大小
                    if (file.getSize() > MAX_FILE_SIZE) {
                        itemResult.put("error", "图片大小超过5MB限制");
                        failList.add(itemResult);
                        continue;
                    }

                    // 校验文件类型
                    String contentType = file.getContentType();
                    if (!uploadFilePathUtils.isValidImageType(contentType)) {
                        itemResult.put("error", "不支持的图片格式");
                        failList.add(itemResult);
                        continue;
                    }

                    // 校验文件扩展名
                    String originalFilename = file.getOriginalFilename();
                    if (!uploadFilePathUtils.isValidImageExtension(originalFilename)) {
                        itemResult.put("error", "不支持的图片格式");
                        failList.add(itemResult);
                        continue;
                    }

                    // 生成唯一文件名
                    String fileName = uploadFilePathUtils.generateImageFileName(originalFilename);

                    // 获取完整保存路径
                    String filePath = uploadFilePathUtils.getImagePath(fileName);

                    // 保存文件
                    File destFile = new File(filePath);
                    file.transferTo(destFile);

                    // 获取访问URL
                    String imageUrl = uploadFilePathUtils.getImageUrl(fileName);

                    itemResult.put("url", imageUrl);
                    itemResult.put("fileName", fileName);
                    successList.add(itemResult);

                } catch (Exception e) {
                    itemResult.put("error", "上传失败：" + e.getMessage());
                    failList.add(itemResult);
                }
            }

            result.put("success", successList);
            result.put("fail", failList);
            result.put("successCount", successList.size());
            result.put("failCount", failList.size());

            if (successList.isEmpty()) {
                return Result.error(500, "所有图片上传失败");
            }

            return Result.success("上传完成", result);

        } catch (Exception e) {
            return Result.error("批量上传失败：" + e.getMessage());
        }
    }
}
