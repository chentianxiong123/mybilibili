package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.ProhibitedWordService;
import com.mybilibili.common.entity.ProhibitedWord;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.util.ArrayList;
import java.util.List;

@RestController
@RequestMapping("/admin/prohibited-words")
@Tag(name = "违禁词管理接口", description = "违禁词词典管理相关操作")
public class ProhibitedWordController {

    @Autowired
    private ProhibitedWordService prohibitedWordService;

    @GetMapping
    @Operation(summary = "获取违禁词列表", description = "获取违禁词列表，支持分页和搜索")
    public Object getWordList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword) {
        return prohibitedWordService.getWordList(page, size, keyword);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取违禁词详情", description = "根据ID获取违禁词详情")
    public Object getWordById(@PathVariable Integer id) {
        return prohibitedWordService.getWordById(id);
    }

    @PostMapping
    @Operation(summary = "添加违禁词", description = "添加新违禁词")
    public Object addWord(@RequestBody ProhibitedWord word) {
        return prohibitedWordService.addWord(word);
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新违禁词", description = "更新违禁词信息")
    public Object updateWord(@PathVariable Integer id, @RequestBody ProhibitedWord word) {
        return prohibitedWordService.updateWord(id, word);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除违禁词", description = "删除违禁词")
    public Object deleteWord(@PathVariable Integer id) {
        return prohibitedWordService.deleteWord(id);
    }

    @PostMapping("/batch-import")
    @Operation(summary = "批量导入违禁词", description = "从文件批量导入违禁词，支持txt文件，每行一个词")
    public Object batchImport(
            @RequestParam MultipartFile file,
            @RequestParam(defaultValue = "CONTAINS") String matchType,
            @RequestParam(required = false) String category) {
        try {
            List<String> words = new ArrayList<>();
            BufferedReader reader = new BufferedReader(new InputStreamReader(file.getInputStream()));
            String line;
            while ((line = reader.readLine()) != null) {
                if (!line.trim().isEmpty()) {
                    words.add(line.trim());
                }
            }
            reader.close();
            return prohibitedWordService.batchImport(words, matchType, category);
        } catch (Exception e) {
            return com.mybilibili.common.vo.Result.error("文件读取失败: " + e.getMessage());
        }
    }
}
