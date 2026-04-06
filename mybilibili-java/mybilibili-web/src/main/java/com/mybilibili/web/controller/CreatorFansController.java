package com.mybilibili.web.controller;

import com.mybilibili.common.utils.JwtUtils;
import com.mybilibili.common.vo.FanVO;
import com.mybilibili.common.vo.FansStatsVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.service.CreatorFansService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 创作者粉丝管理控制器
 */
@RestController
@RequestMapping("/creator/fans")
@Tag(name = "创作者粉丝管理", description = "创作者粉丝列表、统计等接口")
public class CreatorFansController {

    @Autowired
    private CreatorFansService creatorFansService;

    @GetMapping
    @Operation(summary = "获取粉丝列表", description = "获取创作者的粉丝列表，支持分页和互关筛选")
    @SecurityRequirement(name = "JWT")
    public Result<Map<String, Object>> getFansList(
            @Parameter(description = "页码，默认1") @RequestParam(value = "page", defaultValue = "1") Integer page,
            @Parameter(description = "每页大小，默认20") @RequestParam(value = "size", defaultValue = "20") Integer size,
            @Parameter(description = "是否只显示互关粉丝") @RequestParam(value = "mutual", required = false) Boolean mutual,
            HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            List<FanVO> fans = creatorFansService.getFansList(userId, page, size, mutual);
            int total = creatorFansService.getFansCount(userId, mutual);

            Map<String, Object> result = new HashMap<>();
            result.put("list", fans);
            result.put("page", page);
            result.put("size", size);
            result.put("total", total);

            return Result.success("获取成功", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @GetMapping("/stats")
    @Operation(summary = "获取粉丝统计数据", description = "获取创作者的粉丝统计数据")
    @SecurityRequirement(name = "JWT")
    public Result<FansStatsVO> getFansStats(HttpServletRequest request) {
        try {
            Integer userId = JwtUtils.getUserIdFromRequest(request);
            FansStatsVO stats = creatorFansService.getFansStats(userId);
            return Result.success("获取成功", stats);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
