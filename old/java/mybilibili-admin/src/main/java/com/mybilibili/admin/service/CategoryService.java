package com.mybilibili.admin.service;

import com.mybilibili.common.vo.Result;

import java.util.Map;

public interface CategoryService {
    Result<?> getCategoryList(Integer page, Integer size, String keyword);
    Result<?> getCategoryById(Integer id);
    Result<?> addCategory(String name);
    Result<?> updateCategory(Integer id, String name);
    Result<?> deleteCategory(Integer id);
}