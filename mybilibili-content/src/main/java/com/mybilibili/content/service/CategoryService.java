package com.mybilibili.content.service;

import com.mybilibili.common.entity.Category;
import com.mybilibili.content.mapper.CategoryMapper;
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
}
