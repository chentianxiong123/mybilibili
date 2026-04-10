package com.mybilibili.content.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.common.entity.Category;
import com.mybilibili.content.service.CategoryService;
import io.swagger.v3.oas.annotations.Operation;
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
}
