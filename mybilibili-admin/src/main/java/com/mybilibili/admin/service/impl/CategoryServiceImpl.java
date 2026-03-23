package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.CategoryMapper;
import com.mybilibili.admin.service.CategoryService;
import com.mybilibili.common.entity.Category;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class CategoryServiceImpl implements CategoryService {

    @Autowired
    private CategoryMapper categoryMapper;

    @Override
    public Result<?> getCategoryList(Integer page, Integer size, String keyword) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;
            if (keyword == null) keyword = "";

            int offset = (page - 1) * size;
            List<Category> categories = categoryMapper.selectCategoriesByKeyword(offset, size, keyword);
            int total = categoryMapper.countCategoriesByKeyword(keyword);

            Map<String, Object> data = new HashMap<>();
            data.put("list", categories);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取分区列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getCategoryById(Integer id) {
        try {
            Category category = categoryMapper.selectById(id);
            if (category == null) {
                return Result.error("分区不存在");
            }
            return Result.success("获取分区详情成功", category);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> addCategory(String name) {
        try {
            int result = categoryMapper.insert(name);
            if (result > 0) {
                return Result.success("添加分区成功", null);
            } else {
                return Result.error("添加分区失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> updateCategory(Integer id, String name) {
        try {
            Category category = categoryMapper.selectById(id);
            if (category == null) {
                return Result.error("分区不存在");
            }

            int result = categoryMapper.update(id, name);
            if (result > 0) {
                return Result.success("更新分区名称成功", null);
            } else {
                return Result.error("更新分区名称失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> deleteCategory(Integer id) {
        try {
            Category category = categoryMapper.selectById(id);
            if (category == null) {
                return Result.error("分区不存在");
            }

            int result = categoryMapper.delete(id);
            if (result > 0) {
                return Result.success("删除分区成功", null);
            } else {
                return Result.error("删除分区失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}