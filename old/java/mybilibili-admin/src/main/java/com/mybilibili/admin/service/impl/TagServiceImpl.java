package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.TagMapper;
import com.mybilibili.admin.service.TagService;
import com.mybilibili.common.entity.Tag;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class TagServiceImpl implements TagService {

    @Autowired
    private TagMapper tagMapper;

    @Override
    public Result<?> getTagList(Integer page, Integer size, String keyword) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;
            if (keyword == null) keyword = "";

            int offset = (page - 1) * size;
            List<Tag> tags = tagMapper.selectTagsByKeyword(offset, size, keyword);
            int total = tagMapper.countTagsByKeyword(keyword);

            Map<String, Object> data = new HashMap<>();
            data.put("list", tags);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取标签列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getTagById(Integer id) {
        try {
            Tag tag = tagMapper.selectById(id);
            if (tag == null) {
                return Result.error("标签不存在");
            }
            return Result.success("获取标签详情成功", tag);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> addTag(String name, String description) {
        try {
            int result = tagMapper.insert(name, description);
            if (result > 0) {
                return Result.success("添加标签成功", null);
            } else {
                return Result.error("添加标签失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> updateTag(Integer id, String name, String description) {
        try {
            Tag tag = tagMapper.selectById(id);
            if (tag == null) {
                return Result.error("标签不存在");
            }

            int result = tagMapper.update(id, name, description);
            if (result > 0) {
                return Result.success("更新标签成功", null);
            } else {
                return Result.error("更新标签失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> deleteTag(Integer id) {
        try {
            Tag tag = tagMapper.selectById(id);
            if (tag == null) {
                return Result.error("标签不存在");
            }

            int result = tagMapper.delete(id);
            if (result > 0) {
                return Result.success("删除标签成功", null);
            } else {
                return Result.error("删除标签失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}