package com.mybilibili.user.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.user.service.CaptchaService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.Map;
import java.util.UUID;

@RestController
@RequestMapping("/captcha")
@Tag(name = "验证码")
public class CaptchaController {

    @Autowired
    private CaptchaService captchaService;

    @PostMapping("/new")
    @Operation(summary = "生成验证码，返回 captchaId 和算术题")
    public Result<Map<String, String>> newCaptcha() {
        String captchaId = UUID.randomUUID().toString();
        String question = captchaService.generate(captchaId);
        return Result.success(Map.of(
                "captchaId", captchaId,
                "question", question
        ));
    }

    @PostMapping("/verify")
    @Operation(summary = "校验验证码")
    public Result<Boolean> verify(@RequestBody Map<String, String> request) {
        String captchaId = request.get("captchaId");
        String answer = request.get("answer");
        if (captchaId == null || answer == null) {
            return Result.error("参数不完整");
        }
        boolean ok = captchaService.verify(captchaId, answer);
        return Result.success(ok);
    }
}