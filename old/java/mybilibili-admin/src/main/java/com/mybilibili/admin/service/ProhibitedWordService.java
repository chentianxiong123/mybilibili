package com.mybilibili.admin.service;

import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.vo.Result;

import java.util.List;

public interface ProhibitedWordService {
    Result<?> getWordList(Integer page, Integer size, String keyword);
    Result<?> getWordById(Integer id);
    Result<?> addWord(ProhibitedWord word);
    Result<?> updateWord(Integer id, ProhibitedWord word);
    Result<?> deleteWord(Integer id);
    Result<?> batchImport(List<String> words, String matchType, String category);
    List<ProhibitedWord> getAllEnabledWords();
}
