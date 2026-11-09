package com.mybilibili.admin.service;

import com.mybilibili.common.vo.Result;

import java.util.Map;

public interface TagService {
    Result<?> getTagList(Integer page, Integer size, String keyword);
    Result<?> getTagById(Integer id);
    Result<?> addTag(String name, String description);
    Result<?> updateTag(Integer id, String name, String description);
    Result<?> deleteTag(Integer id);
}