package com.mybilibili.comment.service;

import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface ProhibitedWordService {

    Result<?> getList(Integer page, Integer size, String keyword, String category, Integer isEnabled);

    Result<?> getById(Integer id);

    Result<?> add(ProhibitedWord prohibitedWord);

    Result<?> update(Integer id, ProhibitedWord prohibitedWord);

    Result<?> delete(Integer id);

    Result<?> batchImport(List<ProhibitedWord> words);
}