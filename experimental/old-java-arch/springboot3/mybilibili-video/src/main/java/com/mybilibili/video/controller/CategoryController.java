package com.mybilibili.video.controller;

import com.mybilibili.common.entity.Category;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.CategoryService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/category")
@Tag(name = "分区接口", description = "获取分区列表")
public class CategoryController {

    @Autowired
    private CategoryService categoryService;

    @GetMapping
    @Operation(summary = "获取全部分区", description = "获取所有分区列表")
    public Result<List<Category>> getCategoryList() {
        List<Category> categories = categoryService.getAllCategories();
        return Result.success(categories);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取单个分区", description = "根据ID获取分区详情")
    public Result<Category> getCategoryById(@PathVariable Integer id) {
        Category category = categoryService.getCategoryById(id);
        if (category == null) {
            return Result.error(404, "分区不存在");
        }
        return Result.success(category);
    }

    @PostMapping
    @Operation(summary = "添加分区（管理员）", description = "创建新的分区")
    public Result<Category> addCategory(@RequestBody Category category) {
        Category newCategory = categoryService.addCategory(category);
        return Result.success("添加成功", newCategory);
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新分区（管理员）", description = "更新指定分区的信息")
    public Result<Category> updateCategory(
            @PathVariable Integer id,
            @RequestBody Category category) {
        Category updated = categoryService.updateCategory(id, category);
        if (updated == null) {
            return Result.error(404, "分区不存在");
        }
        return Result.success("更新成功", updated);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除分区（管理员）", description = "删除指定的分区")
    public Result<Void> deleteCategory(@PathVariable Integer id) {
        boolean success = categoryService.deleteCategory(id);
        if (success) {
            return Result.success("删除成功", null);
        } else {
            return Result.error(404, "分区不存在或删除失败");
        }
    }
}
