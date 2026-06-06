package com.mybilibili.video.service;

import com.mybilibili.common.entity.Category;
import com.mybilibili.video.mapper.CategoryMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class CategoryService {

    @Autowired
    private CategoryMapper categoryMapper;

    public List<Category> getAllCategories() {
        return categoryMapper.selectList(null);
    }

    public Category getCategoryById(Integer id) {
        return categoryMapper.selectById(id);
    }

    public Category addCategory(Category category) {
        categoryMapper.insert(category);
        return category;
    }

    public Category updateCategory(Integer id, Category category) {
        Category existing = categoryMapper.selectById(id);
        if (existing == null) {
            return null;
        }
        category.setId(id);
        categoryMapper.updateById(category);
        return categoryMapper.selectById(id);
    }

    public boolean deleteCategory(Integer id) {
        return categoryMapper.deleteById(id) > 0;
    }
}
