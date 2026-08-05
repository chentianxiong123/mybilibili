package com.mybilibili.user.service;

public interface EmailCodeService {
    /**
     * 发送邮箱验证码，存 Redis，5分钟有效
     * @param email 目标邮箱
     * @return 邮件中的验证码（纯数字6位）
     */
    String sendCode(String email);

    /**
     * 校验验证码
     */
    boolean verify(String email, String code);
}