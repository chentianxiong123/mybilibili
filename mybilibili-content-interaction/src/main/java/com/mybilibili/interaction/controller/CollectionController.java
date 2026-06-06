package com.mybilibili.interaction.controller;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.vo.ManuscriptCollectionVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.interaction.service.CollectionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@Slf4j
@RestController
@RequestMapping("/collection")
@Tag(name = "合集接口", description = "合集相关操作")
public class CollectionController {

    @Autowired
    private CollectionService collectionService;

    @Autowired
    private ObjectMapper objectMapper;

    @GetMapping("/user/{userId}")
    @Operation(summary = "获取用户的合集列表", description = "获取指定用户创建的所有合集")
    public Result<List<ManuscriptCollectionVO>> getCollectionsByUserId(
            @Parameter(description = "用户ID") @PathVariable Integer userId) {
        try {
            List<ManuscriptCollectionVO> collections = collectionService.getCollectionsByUserId(userId);
            return Result.success("获取成功", collections);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{collectionId}")
    @Operation(summary = "获取合集详情", description = "获取指定合集的详细信息和包含的稿件")
    public Result<ManuscriptCollectionVO> getCollectionById(
            @Parameter(description = "合集ID") @PathVariable Integer collectionId,
            @Parameter(description = "当前用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer currentUserId) {
        try {
            ManuscriptCollectionVO collection = collectionService.getCollectionById(collectionId, currentUserId);
            if (collection == null) {
                return Result.error("合集不存在");
            }
            return Result.success("获取成功", collection);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/{collectionId}/manuscripts")
    @Operation(summary = "获取合集中的稿件列表", description = "获取指定合集包含的所有稿件")
    public Result<List<ManuscriptCollectionVO.ManuscriptItemVO>> getCollectionManuscripts(
            @Parameter(description = "合集ID") @PathVariable Integer collectionId) {
        try {
            List<ManuscriptCollectionVO.ManuscriptItemVO> manuscripts = collectionService.getCollectionManuscripts(collectionId);
            return Result.success("获取成功", manuscripts);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping
    @Operation(summary = "创建合集", description = "创建一个新的合集")
    public Result<ManuscriptCollectionVO> createCollection(
            @Parameter(description = "合集名称") @RequestParam String name,
            @Parameter(description = "合集描述") @RequestParam(required = false) String description,
            @Parameter(description = "是否公开") @RequestParam(required = false) String isPublic,
            @Parameter(description = "稿件ID列表(JSON字符串)") @RequestParam(required = false) String manuscriptIds,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            if (userId == null) {
                return Result.error("请先登录");
            }
            if (name == null || name.trim().isEmpty()) {
                return Result.error("请输入合集名称");
            }

            boolean publicStatus = !"false".equals(isPublic);
            log.info("接收到的参数 - name: {}, isPublic: {}, manuscriptIds: {}", name, isPublic, manuscriptIds);

            List<Integer> manuscriptIdList = null;
            if (manuscriptIds != null && !manuscriptIds.isEmpty()) {
                try {
                    manuscriptIdList = objectMapper.readValue(manuscriptIds, new TypeReference<List<Integer>>() {});
                    log.info("解析稿件ID列表: {}", manuscriptIdList);
                } catch (Exception e) {
                    log.error("解析稿件ID列表失败: {}", manuscriptIds, e);
                }
            }

            Integer status = publicStatus ? 1 : 0;
            ManuscriptCollectionVO collection = collectionService.createCollection(name, description, userId, status, manuscriptIdList);
            return Result.success("创建成功", collection);
        } catch (Exception e) {
            log.error("创建合集失败", e);
            return Result.error(e.getMessage());
        }
    }

    @PutMapping("/{collectionId}")
    @Operation(summary = "更新合集", description = "更新合集信息")
    public Result<ManuscriptCollectionVO> updateCollection(
            @Parameter(description = "合集ID") @PathVariable Integer collectionId,
            @Parameter(description = "合集名称") @RequestParam(required = false) String name,
            @Parameter(description = "合集描述") @RequestParam(required = false) String description,
            @Parameter(description = "是否公开") @RequestParam(required = false) Boolean isPublic,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            if (userId == null) {
                return Result.error("请先登录");
            }
            Integer status = (isPublic != null) ? (isPublic ? 1 : 0) : null;
            ManuscriptCollectionVO collection = collectionService.updateCollection(collectionId, name, description, userId, status);
            if (collection == null) {
                return Result.error("合集不存在或无权修改");
            }
            return Result.success("更新成功", collection);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{collectionId}")
    @Operation(summary = "删除合集", description = "删除指定合集")
    public Result<?> deleteCollection(
            @Parameter(description = "合集ID") @PathVariable Integer collectionId,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            if (userId == null) {
                return Result.error("请先登录");
            }
            collectionService.deleteCollection(collectionId, userId);
            return Result.success("删除成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @PostMapping("/{collectionId}/manuscript/{manuscriptId}")
    @Operation(summary = "添加稿件到合集", description = "将指定稿件添加到合集中")
    public Result<?> addManuscriptToCollection(
            @Parameter(description = "合集ID") @PathVariable Integer collectionId,
            @Parameter(description = "稿件ID") @PathVariable Integer manuscriptId,
            @Parameter(description = "排序顺序") @RequestParam(defaultValue = "0") Integer order,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            if (userId == null) {
                return Result.error("请先登录");
            }
            collectionService.addManuscriptToCollection(collectionId, manuscriptId, userId, order);
            return Result.success("添加成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @DeleteMapping("/{collectionId}/manuscript/{manuscriptId}")
    @Operation(summary = "从合集中移除稿件", description = "从合集中移除指定稿件")
    public Result<?> removeManuscriptFromCollection(
            @Parameter(description = "合集ID") @PathVariable Integer collectionId,
            @Parameter(description = "稿件ID") @PathVariable Integer manuscriptId,
            @Parameter(description = "用户ID") @RequestHeader(value = "X-User-Id", required = false) Integer userId) {
        try {
            if (userId == null) {
                return Result.error("请先登录");
            }
            collectionService.removeManuscriptFromCollection(collectionId, manuscriptId, userId);
            return Result.success("移除成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
