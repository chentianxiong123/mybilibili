package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.ProhibitedWordMapper;
import com.mybilibili.admin.service.ProhibitedWordService;
import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class ProhibitedWordServiceImpl implements ProhibitedWordService {

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    @Override
    public Result<?> getWordList(Integer page, Integer size, String keyword) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;
            if (keyword == null) keyword = "";

            int offset = (page - 1) * size;
            List<ProhibitedWord> words = prohibitedWordMapper.selectByKeyword(offset, size, keyword);
            int total = prohibitedWordMapper.countByKeyword(keyword);

            Map<String, Object> data = new HashMap<>();
            data.put("list", words);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取违禁词列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getWordById(Integer id) {
        try {
            ProhibitedWord word = prohibitedWordMapper.selectById(id);
            if (word == null) {
                return Result.error("违禁词不存在");
            }
            return Result.success("获取违禁词成功", word);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> addWord(ProhibitedWord word) {
        try {
            if (word.getWord() == null || word.getWord().trim().isEmpty()) {
                return Result.error("违禁词不能为空");
            }

            // 检查是否已存在
            ProhibitedWord existing = prohibitedWordMapper.selectByWord(word.getWord().trim());
            if (existing != null) {
                return Result.error("该违禁词已存在");
            }

            // 设置默认值
            if (word.getMatchType() == null) {
                word.setMatchType("CONTAINS");
            }
            if (word.getIsEnabled() == null) {
                word.setIsEnabled(1);
            }

            word.setWord(word.getWord().trim());

            int result = prohibitedWordMapper.insert(word);
            if (result > 0) {
                return Result.success("添加违禁词成功", null);
            } else {
                return Result.error("添加违禁词失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> updateWord(Integer id, ProhibitedWord word) {
        try {
            ProhibitedWord existing = prohibitedWordMapper.selectById(id);
            if (existing == null) {
                return Result.error("违禁词不存在");
            }

            if (word.getWord() != null && !word.getWord().trim().isEmpty()) {
                // 检查新词是否与其他词冲突
                ProhibitedWord conflict = prohibitedWordMapper.selectByWord(word.getWord().trim());
                if (conflict != null && !conflict.getId().equals(id)) {
                    return Result.error("该违禁词已存在");
                }
                word.setWord(word.getWord().trim());
            }

            word.setId(id);
            int result = prohibitedWordMapper.update(word);
            if (result > 0) {
                return Result.success("更新违禁词成功", null);
            } else {
                return Result.error("更新违禁词失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> deleteWord(Integer id) {
        try {
            ProhibitedWord existing = prohibitedWordMapper.selectById(id);
            if (existing == null) {
                return Result.error("违禁词不存在");
            }

            int result = prohibitedWordMapper.deleteById(id);
            if (result > 0) {
                return Result.success("删除违禁词成功", null);
            } else {
                return Result.error("删除违禁词失败");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> batchImport(List<String> words, String matchType, String category) {
        try {
            if (words == null || words.isEmpty()) {
                return Result.error("违禁词列表不能为空");
            }

            int successCount = 0;
            int failCount = 0;

            for (String wordStr : words) {
                if (wordStr == null || wordStr.trim().isEmpty()) {
                    continue;
                }

                wordStr = wordStr.trim();

                // 检查是否已存在
                ProhibitedWord existing = prohibitedWordMapper.selectByWord(wordStr);
                if (existing != null) {
                    failCount++;
                    continue;
                }

                ProhibitedWord word = new ProhibitedWord();
                word.setWord(wordStr);
                word.setMatchType(matchType != null ? matchType : "CONTAINS");
                word.setCategory(category);
                word.setIsEnabled(1);

                try {
                    prohibitedWordMapper.insert(word);
                    successCount++;
                } catch (Exception e) {
                    failCount++;
                }
            }

            Map<String, Object> result = new HashMap<>();
            result.put("successCount", successCount);
            result.put("failCount", failCount);

            return Result.success("批量导入完成", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public List<ProhibitedWord> getAllEnabledWords() {
        return prohibitedWordMapper.selectAllEnabled();
    }
}
