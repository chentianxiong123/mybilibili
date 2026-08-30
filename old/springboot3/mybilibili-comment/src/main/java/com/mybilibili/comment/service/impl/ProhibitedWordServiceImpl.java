package com.mybilibili.comment.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.comment.mapper.ProhibitedWordMapper;
import com.mybilibili.comment.service.ProhibitedWordService;
import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class ProhibitedWordServiceImpl implements ProhibitedWordService {

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    @Override
    public Result<?> getList(Integer page, Integer size, String keyword, String category, Integer isEnabled) {
        int offset = (page - 1) * size;
        List<ProhibitedWord> list = prohibitedWordMapper.selectList(keyword, category, isEnabled, offset, size);
        int total = prohibitedWordMapper.countList(keyword, category, isEnabled);

        Map<String, Object> data = new HashMap<>();
        data.put("list", list);
        data.put("total", total);

        return Result.success(data);
    }

    @Override
    public Result<?> getById(Integer id) {
        ProhibitedWord prohibitedWord = prohibitedWordMapper.selectById(id);
        if (prohibitedWord == null) {
            return Result.error("违禁词不存在");
        }
        return Result.success(prohibitedWord);
    }

    @Override
    public Result<?> add(ProhibitedWord prohibitedWord) {
        if (!StringUtils.hasText(prohibitedWord.getWord())) {
            return Result.error("违禁词内容不能为空");
        }

        ProhibitedWord existing = prohibitedWordMapper.selectByWord(prohibitedWord.getWord());
        if (existing != null) {
            return Result.error("违禁词已存在");
        }

        prohibitedWord.setCreatedAt(new Date());
        prohibitedWord.setUpdatedAt(new Date());
        if (prohibitedWord.getIsEnabled() == null) {
            prohibitedWord.setIsEnabled(1);
        }

        prohibitedWordMapper.insert(prohibitedWord);
        return Result.success("添加成功", null);
    }

    @Override
    public Result<?> update(Integer id, ProhibitedWord prohibitedWord) {
        ProhibitedWord existing = prohibitedWordMapper.selectById(id);
        if (existing == null) {
            return Result.error("违禁词不存在");
        }

        if (StringUtils.hasText(prohibitedWord.getWord())) {
            existing.setWord(prohibitedWord.getWord());
        }
        if (StringUtils.hasText(prohibitedWord.getMatchType())) {
            existing.setMatchType(prohibitedWord.getMatchType());
        }
        if (StringUtils.hasText(prohibitedWord.getCategory())) {
            existing.setCategory(prohibitedWord.getCategory());
        }
        if (prohibitedWord.getIsEnabled() != null) {
            existing.setIsEnabled(prohibitedWord.getIsEnabled());
        }
        existing.setUpdatedAt(new Date());

        prohibitedWordMapper.updateById(existing);
        return Result.success("更新成功", null);
    }

    @Override
    public Result<?> delete(Integer id) {
        prohibitedWordMapper.deleteById(id);
        return Result.success("删除成功", null);
    }

    @Override
    public Result<?> batchImport(List<ProhibitedWord> words) {
        int successCount = 0;
        int failCount = 0;

        for (ProhibitedWord word : words) {
            try {
                ProhibitedWord existing = prohibitedWordMapper.selectByWord(word.getWord());
                if (existing == null) {
                    word.setCreatedAt(new Date());
                    word.setUpdatedAt(new Date());
                    if (word.getIsEnabled() == null) {
                        word.setIsEnabled(1);
                    }
                    prohibitedWordMapper.insert(word);
                    successCount++;
                } else {
                    failCount++;
                }
            } catch (Exception e) {
                failCount++;
            }
        }

        Map<String, Object> data = new HashMap<>();
        data.put("successCount", successCount);
        data.put("failCount", failCount);

        return Result.success(data);
    }
}