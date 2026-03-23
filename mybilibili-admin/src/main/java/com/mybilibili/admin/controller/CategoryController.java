package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.CategoryService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/categories")
@Tag(name = "分区管理接口", description = "分区管理相关操作")
public class CategoryController {

    @Autowired
    private CategoryService categoryService;

    @GetMapping
    @Operation(summary = "获取分区列表", description = "获取分区列表，支持分页和搜索")
    public Object getCategoryList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword) {
        return categoryService.getCategoryList(page, size, keyword);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取分区详情", description = "根据ID获取分区详情")
    public Object getCategoryById(@PathVariable Integer id) {
        return categoryService.getCategoryById(id);
    }

    @PostMapping
    @Operation(summary = "添加分区", description = "添加新分区")
    public Object addCategory(@RequestParam String name) {
        return categoryService.addCategory(name);
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新分区名称", description = "更新分区名称")
    public Object updateCategory(@PathVariable Integer id, @RequestParam String name) {
        return categoryService.updateCategory(id, name);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除分区", description = "删除分区")
    public Object deleteCategory(@PathVariable Integer id) {
        return categoryService.deleteCategory(id);
    }
}