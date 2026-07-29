package com.mybilibili.user.service;

public interface CaptchaService {
    /**
     * 生成验证码，保存答案到 Redis
     * @param captchaId 唯一ID
     * @return 显示给用户的算术题文字（如 "3 + 5 = ?"）
     */
    String generate(String captchaId);

    /**
     * 校验答案
     */
    boolean verify(String captchaId, String answer);
}