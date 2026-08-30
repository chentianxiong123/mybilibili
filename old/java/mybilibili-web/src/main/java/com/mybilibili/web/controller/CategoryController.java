package com.mybilibili.web.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.web.mapper.CategoryMapper;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/category")
@Tag(name = "分类相关接口", description = "获取分类列表等操作")
public class CategoryController {

    @Autowired
    private CategoryMapper categoryMapper;

    @GetMapping
    @Operation(summary = "获取分类列表", description = "获取所有分类列表")
    public Result<?> getCategoryList() {
        try {
            return Result.success("获取成功", categoryMapper.selectAll());
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}