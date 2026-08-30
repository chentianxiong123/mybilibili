package com.mybilibili.admin.controller;

import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/test")
@Tag(name = "测试接口", description = "管理员端测试接口")
public class TestController {

    @GetMapping
    @Operation(summary = "测试接口", description = "测试管理员端服务是否正常")
    public String test(){
        return "hello world";
    }
}
