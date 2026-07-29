package com.mybilibili.search.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.search.service.HotSearchService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/search")
@Tag(name = "热搜榜", description = "热搜榜相关接口")
public class HotSearchController {

    @Autowired
    private HotSearchService hotSearchService;

    @GetMapping("/hot")
    @Operation(summary = "获取热搜榜", description = "获取热度最高的10个搜索关键词")
    public Result<List<HotSearchService.HotSearchVO>> getHotSearchTop10() {
        try {
            List<HotSearchService.HotSearchVO> hotSearchList = hotSearchService.getHotSearchTop10();
            return Result.success("获取成功", hotSearchList);
        } catch (Exception e) {
            return Result.error("获取热搜榜失败: " + e.getMessage());
        }
    }
}