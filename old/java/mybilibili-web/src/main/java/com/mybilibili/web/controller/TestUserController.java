package com.mybilibili.web.controller;

import com.mybilibili.common.vo.Result;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/test")
@Tag(name = "测试接口", description = "用于测试服务是否正常运行")
public class TestUserController {

    @GetMapping("/user")
    @Operation(summary = "测试用户接口", description = "测试用户管理功能是否正常")
    public Result<String> testUser() {
        return Result.success("测试用户接口成功");
    }
}
