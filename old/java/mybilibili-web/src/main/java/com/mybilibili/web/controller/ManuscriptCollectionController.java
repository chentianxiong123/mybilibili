package com.mybilibili.web.controller;

import com.mybilibili.common.entity.ManuscriptCollection;
import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.ManuscriptCollectionVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.ManuscriptCollectionService;
import com.mybilibili.web.utils.UploadFilePathUtils;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import javax.servlet.http.HttpServletRequest;
import java.io.File;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/collection")
@Tag(name = "合集接口", description = "合集创建、查询、更新、删除及稿件管理")
public class ManuscriptCollectionController {

    @Autowired
    private ManuscriptCollectionService collectionService;

    @Autowired
    private UploadFilePathUtils uploadFilePathUtils;

    @PostMapping
    @Operation(summary = "创建合集", description = "创建新的稿件合集")
    @SecurityRequirement(name = "JWT")
    public Result<ManuscriptCollectionVO> createCollection(
            @RequestParam("name") String name,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam(value = "cover", required = false) MultipartFile cover,
            @RequestParam(value = "isPublic", defaultValue = "true") String isPublicStr,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            
            // 解析isPublic参数
            boolean isPublic = !"false".equalsIgnoreCase(isPublicStr);
            
            // 构建合集对象
            ManuscriptCollection collection = new ManuscriptCollection();
            collection.setTitle(name);
            collection.setDescription(description);
            collection.setStatus(isPublic ? ManuscriptCollection.STATUS_PUBLIC : ManuscriptCollection.STATUS_PRIVATE);
            
            // 处理封面文件上传（如果有）
            if (cover != null && !cover.isEmpty()) {
                // 校验图片类型
                if (!uploadFilePathUtils.isValidImageType(cover.getContentType())) {
                    return Result.error("封面图片格式不支持，仅支持 jpg、jpeg、png、gif、webp 格式");
                }
                
                // 生成文件名并保存
                String fileName = uploadFilePathUtils.generateImageFileName(cover.getOriginalFilename());
                String imagePath = uploadFilePathUtils.getImagePath(fileName);
                
                // 确保目录存在
                uploadFilePathUtils.createImagesDirectory();
                
                // 保存文件
                File destFile = new File(imagePath);
                cover.transferTo(destFile);
                
                // 设置封面URL
                collection.setCoverUrl(uploadFilePathUtils.getImageUrl(fileName));
            }
            
            ManuscriptCollectionVO vo = collectionService.createCollection(collection, userId);
            return Result.success("创建成功", vo);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取合集详情", description = "根据合集ID获取详情，包含稿件列表")
    public Result<ManuscriptCollectionVO> getCollectionById(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            // 尝试从JWT中获取当前用户ID
            Integer currentUserId = null;
            try {
                String authHeader = request.getHeader("Authorization");
                if (authHeader != null && authHeader.startsWith("Bearer ")) {
                    String token = authHeader.substring(7);
                    currentUserId = JwtUtils.getUserIdFromToken(token);
                }
            } catch (Exception e) {
                // 未登录或token无效
            }

            ManuscriptCollectionVO vo = collectionService.getCollectionById(id, currentUserId);
            if (vo == null) {
                return Result.error("合集不存在");
            }
            return Result.success("获取成功", vo);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}")
    @Operation(summary = "获取用户合集列表", description = "根据用户ID获取其所有合集")
    public Result<List<ManuscriptCollectionVO>> getCollectionsByUserId(@PathVariable Integer userId) {
        try {
            List<ManuscriptCollectionVO> collections = collectionService.getCollectionsByUserId(userId);
            return Result.success("获取成功", collections);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/user/{userId}/status/{status}")
    @Operation(summary = "获取用户合集列表（按状态筛选）", description = "根据用户ID和状态获取合集列表")
    public Result<List<ManuscriptCollectionVO>> getCollectionsByUserIdAndStatus(
            @PathVariable Integer userId,
            @PathVariable Integer status) {
        try {
            List<ManuscriptCollectionVO> collections = collectionService.getCollectionsByUserIdAndStatus(userId, status);
            return Result.success("获取成功", collections);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新合集", description = "更新合集信息")
    @SecurityRequirement(name = "JWT")
    public Result<ManuscriptCollectionVO> updateCollection(
            @PathVariable Integer id,
            @RequestParam(value = "name", required = false) String name,
            @RequestParam(value = "description", required = false) String description,
            @RequestParam(value = "cover", required = false) MultipartFile cover,
            @RequestParam(value = "isPublic", required = false) String isPublicStr,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            
            // 构建合集对象
            ManuscriptCollection collection = new ManuscriptCollection();
            collection.setTitle(name);
            collection.setDescription(description);
            if (isPublicStr != null && !isPublicStr.isEmpty()) {
                boolean isPublic = !"false".equalsIgnoreCase(isPublicStr);
                collection.setStatus(isPublic ? ManuscriptCollection.STATUS_PUBLIC : ManuscriptCollection.STATUS_PRIVATE);
            }
            
            // 处理封面文件上传（如果有）
            if (cover != null && !cover.isEmpty()) {
                // 校验图片类型
                if (!uploadFilePathUtils.isValidImageType(cover.getContentType())) {
                    return Result.error("封面图片格式不支持，仅支持 jpg、jpeg、png、gif、webp 格式");
                }
                
                // 生成文件名并保存
                String fileName = uploadFilePathUtils.generateImageFileName(cover.getOriginalFilename());
                String imagePath = uploadFilePathUtils.getImagePath(fileName);
                
                // 确保目录存在
                uploadFilePathUtils.createImagesDirectory();
                
                // 保存文件
                File destFile = new File(imagePath);
                cover.transferTo(destFile);
                
                // 设置封面URL
                collection.setCoverUrl(uploadFilePathUtils.getImageUrl(fileName));
            }
            
            ManuscriptCollectionVO vo = collectionService.updateCollection(id, collection, userId);
            return Result.success("更新成功", vo);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除合集", description = "删除合集及其中的所有稿件关联")
    @SecurityRequirement(name = "JWT")
    public Result<?> deleteCollection(
            @PathVariable Integer id,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            collectionService.deleteCollection(id, userId);
            return Result.success("删除成功", null);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{collectionId}/manuscript/{manuscriptId}")
    @Operation(summary = "添加稿件到合集", description = "将稿件添加到指定合集")
    @SecurityRequirement(name = "JWT")
    public Result<?> addManuscriptToCollection(
            @PathVariable Integer collectionId,
            @PathVariable Integer manuscriptId,
            @RequestParam(defaultValue = "0") Integer order,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean success = collectionService.addManuscriptToCollection(collectionId, manuscriptId, userId);
            if (success) {
                return Result.success("添加成功", null);
            } else {
                return Result.error("添加失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{collectionId}/manuscript/{manuscriptId}")
    @Operation(summary = "从合集中移除稿件", description = "将稿件从合集中移除")
    @SecurityRequirement(name = "JWT")
    public Result<?> removeManuscriptFromCollection(
            @PathVariable Integer collectionId,
            @PathVariable Integer manuscriptId,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean success = collectionService.removeManuscriptFromCollection(collectionId, manuscriptId, userId);
            if (success) {
                return Result.success("移除成功", null);
            } else {
                return Result.error("移除失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{collectionId}/manuscript/{manuscriptId}/order")
    @Operation(summary = "调整稿件顺序", description = "调整稿件在合集中的顺序")
    @SecurityRequirement(name = "JWT")
    public Result<?> updateManuscriptOrder(
            @PathVariable Integer collectionId,
            @PathVariable Integer manuscriptId,
            @RequestParam Integer order,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            boolean success = collectionService.updateManuscriptOrder(collectionId, manuscriptId, order, userId);
            if (success) {
                return Result.success("调整成功", null);
            } else {
                return Result.error("调整失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{collectionId}/manuscripts")
    @Operation(summary = "获取合集中的稿件列表", description = "根据合集ID获取其中的稿件列表")
    public Result<List<ManuscriptCollectionVO.ManuscriptItemVO>> getCollectionManuscripts(
            @PathVariable Integer collectionId,
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "20") Integer size) {
        try {
            ManuscriptCollectionVO vo = collectionService.getCollectionById(collectionId);
            if (vo == null) {
                return Result.error("合集不存在");
            }
            return Result.success("获取成功", vo.getManuscripts());
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{collectionId}/manuscripts/order")
    @Operation(summary = "批量调整稿件顺序", description = "批量调整合集中所有稿件的顺序")
    @SecurityRequirement(name = "JWT")
    public Result<?> batchUpdateManuscriptOrder(
            @PathVariable Integer collectionId,
            @RequestBody Map<String, List<Integer>> requestBody,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromToken(request.getHeader("Authorization"));
            List<Integer> manuscriptIds = requestBody.get("manuscriptIds");
            if (manuscriptIds == null || manuscriptIds.isEmpty()) {
                return Result.error("稿件ID列表不能为空");
            }
            boolean success = collectionService.batchUpdateManuscriptOrder(collectionId, manuscriptIds, userId);
            if (success) {
                return Result.success("批量调整成功", null);
            } else {
                return Result.error("批量调整失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
