package com.mybilibili.admin.controller;

import com.mybilibili.admin.service.TagService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/tags")
@Tag(name = "标签管理接口", description = "标签管理相关操作")
public class TagController {

    @Autowired
    private TagService tagService;

    @GetMapping
    @Operation(summary = "获取标签列表", description = "获取标签列表，支持分页和搜索")
    public Object getTagList(
            @RequestParam(defaultValue = "1") Integer page,
            @RequestParam(defaultValue = "10") Integer size,
            @RequestParam(required = false) String keyword) {
        return tagService.getTagList(page, size, keyword);
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取标签详情", description = "根据ID获取标签详情")
    public Object getTagById(@PathVariable Integer id) {
        return tagService.getTagById(id);
    }

    @PostMapping
    @Operation(summary = "添加标签", description = "添加新标签")
    public Object addTag(@RequestParam String name, @RequestParam(required = false) String description) {
        return tagService.addTag(name, description);
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新标签", description = "更新标签信息")
    public Object updateTag(@PathVariable Integer id, @RequestParam String name, @RequestParam(required = false) String description) {
        return tagService.updateTag(id, name, description);
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除标签", description = "删除标签")
    public Object deleteTag(@PathVariable Integer id) {
        return tagService.deleteTag(id);
    }
}