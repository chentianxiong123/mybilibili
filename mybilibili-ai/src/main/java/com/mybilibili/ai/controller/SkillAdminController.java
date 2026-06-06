package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.AiSkill;
import com.mybilibili.ai.service.AiSkillService;
import com.mybilibili.ai.service.SkillRoutingService;
import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/skills")
@Tag(name = "AI技能管理")
public class SkillAdminController {

    @Autowired
    private AiSkillService aiSkillService;

    @Autowired
    private SkillRoutingService skillRoutingService;

    @GetMapping
    @Operation(summary = "获取所有AI技能")
    public Result<List<AiSkill>> listAll() {
        return Result.success(aiSkillService.listAll());
    }

    @GetMapping("/type/{type}")
    @Operation(summary = "按类型获取AI技能")
    public Result<List<AiSkill>> listByType(@PathVariable String type) {
        return Result.success(aiSkillService.listByType(type));
    }

    @PostMapping("/customer-service/defaults")
    @Operation(summary = "初始化默认客服技能")
    public Result<List<AiSkill>> createCustomerServiceDefaults() {
        List<AiSkill> created = aiSkillService.createMissingCustomerServiceDefaults();
        return Result.success("已新增 " + created.size() + " 个默认客服技能", created);
    }

    @PostMapping("/customer-service/route-test")
    @Operation(summary = "测试客服技能路由")
    public Result<List<AiSkill>> testCustomerServiceRoute(@RequestBody Map<String, String> request) {
        String question = request.get("question");
        if (question == null || question.trim().isEmpty()) {
            return Result.error("问题不能为空");
        }
        return Result.success(skillRoutingService.routeSkills(question));
    }

    @GetMapping("/{id}")
    @Operation(summary = "获取AI技能详情")
    public Result<AiSkill> getById(@PathVariable Long id) {
        AiSkill skill = aiSkillService.getById(id);
        return skill != null ? Result.success(skill) : Result.error("技能不存在");
    }

    @PostMapping
    @Operation(summary = "创建AI技能")
    public Result<AiSkill> create(@RequestBody AiSkill skill) {
        return Result.success(aiSkillService.create(skill));
    }

    @PutMapping("/{id}")
    @Operation(summary = "更新AI技能")
    public Result<AiSkill> update(@PathVariable Long id, @RequestBody AiSkill skill) {
        return Result.success(aiSkillService.update(id, skill));
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "删除AI技能")
    public Result<?> delete(@PathVariable Long id) {
        aiSkillService.delete(id);
        return Result.success("删除成功");
    }

    @PutMapping("/{id}/toggle")
    @Operation(summary = "启用/禁用AI技能")
    public Result<?> toggleEnabled(@PathVariable Long id) {
        aiSkillService.toggleEnabled(id);
        return Result.success("操作成功");
    }
}
