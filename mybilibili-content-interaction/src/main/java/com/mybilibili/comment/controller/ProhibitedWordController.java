package com.mybilibili.comment.controller;

import com.mybilibili.comment.service.ProhibitedWordService;
import com.mybilibili.common.entity.ProhibitedWord;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/admin/prohibited-words")
@Tag(name = "违禁词管理接口", description = "违禁词管理相关操作")
public class ProhibitedWordController {

    @Autowired
    private ProhibitedWordService prohibitedWordService;

    @GetMapping
    @Operation(summary = "获取违禁词列表", description = "获取违禁词列表，支持分页和搜索")
    public Result<?> getList(
            @Parameter(description = "页码") @RequestParam(defaultValue = "1") Integer page,
            @Parameter(description = "每页数量") @RequestParam(defaultValue = "10") Integer size,
            @Parameter(description = "搜索关键词") @RequestParam(required = false) String keyword,
            @Parameter(description = "分类") @RequestParam(required = false) String category,
            @Parameter(description = "状态") @RequestParam(required = false) Integer isEnabled) {
        return prohibitedWordService.getList(page, size, keyword, category, isEnabled);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取违禁词详情", description = "根据ID获取违禁词详情")
    public Result<?> getById(@PathVariable Integer id) {
        return prohibitedWordService.getById(id);
    }

    @PostMapping
    @Operation(summary = "添加违禁词", description = "添加新的违禁词")
    public Result<?> add(@RequestBody ProhibitedWord prohibitedWord) {
        return prohibitedWordService.add(prohibitedWord);
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新违禁词", description = "更新违禁词信息")
    public Result<?> update(@PathVariable Integer id, @RequestBody ProhibitedWord prohibitedWord) {
        return prohibitedWordService.update(id, prohibitedWord);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除违禁词", description = "删除违禁词")
    public Result<?> delete(@PathVariable Integer id) {
        return prohibitedWordService.delete(id);
    }

    @PostMapping("/batch-import")
    @Operation(summary = "批量导入违禁词", description = "批量导入违禁词")
    public Result<?> batchImport(@RequestBody List<ProhibitedWord> words) {
        return prohibitedWordService.batchImport(words);
    }
}