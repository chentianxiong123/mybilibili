package com.mybilibili.user.service.impl;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.user.service.EmailCodeService;
import jakarta.mail.internet.MimeMessage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.mail.javamail.JavaMailSender;
import org.springframework.mail.javamail.MimeMessageHelper;
import org.springframework.stereotype.Service;

import java.util.Random;
import java.util.concurrent.TimeUnit;

@Service
public class EmailCodeServiceImpl implements EmailCodeService {

    @Autowired
    private JavaMailSender mailSender;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Value("${spring.mail.username:redacted@example.com}")
    private String fromEmail;

    private static final String PREFIX = "email:code:";
    private static final long EXPIRE_MINUTES = 5;
    private static final int CODE_LENGTH = 6;

    @Override
    public String sendCode(String email) {
        // 检查发送频率：同一邮箱 60 秒内只能发一次
        String rateKey = "email:send:rate:" + email;
        Boolean limited = redisTemplate.hasKey(rateKey);
        if (Boolean.TRUE.equals(limited)) {
            throw new BusinessException("发送过于频繁，请稍后再试");
        }

        String code = String.format("%0" + CODE_LENGTH + "d", new Random().nextInt((int) Math.pow(10, CODE_LENGTH)));

        try {
            MimeMessage message = mailSender.createMimeMessage();
            MimeMessageHelper helper = new MimeMessageHelper(message, true, "UTF-8");
            helper.setFrom(fromEmail);
            helper.setTo(email);
            helper.setSubject("【哔哩哔哩】您的注册验证码");
            helper.setText(buildHtmlContent(code), true);

            mailSender.send(message);
        } catch (Exception e) {
            throw new BusinessException("邮件发送失败: " + e.getMessage());
        }

        // 存 Redis
        redisTemplate.opsForValue().set(PREFIX + email, code, EXPIRE_MINUTES, TimeUnit.MINUTES);
        // 发送频率限制 60 秒
        redisTemplate.opsForValue().set(rateKey, "1", 60, TimeUnit.SECONDS);

        return code;
    }

    @Override
    public boolean verify(String email, String code) {
        if (email == null || code == null) return false;
        String key = PREFIX + email;
        String saved = redisTemplate.opsForValue().get(key);
        if (saved == null) return false;
        boolean ok = saved.equals(code.trim());
        if (ok) redisTemplate.delete(key);
        return ok;
    }

    private String buildHtmlContent(String code) {
        return "<div style='font-family:Arial,sans-serif;max-width:480px;padding:32px;border:1px solid #e5e5e5;border-radius:8px;'>" +
                "<h2 style='color:#00a1d6;margin:0 0 24px;'>哔哩哔哩</h2>" +
                "<p style='color:#333;font-size:16px;line-height:1.6;'>您好！</p>" +
                "<p style='color:#333;font-size:16px;line-height:1.6;'>您的注册验证码是：</p>" +
                "<p style='color:#00a1d6;font-size:32px;font-weight:bold;letter-spacing:8px;margin:24px 0;'>" + code + "</p>" +
                "<p style='color:#888;font-size:14px;line-height:1.6;'>验证码 <strong>5 分钟</strong> 内有效，请尽快完成验证。</p>" +
                "<p style='color:#888;font-size:14px;line-height:1.6;'>如非本人操作，请忽略此邮件。</p>" +
                "<hr style='border:none;border-top:1px solid #e5e5e5;margin:24px 0;'>" +
                "<p style='color:#ccc;font-size:12px;line-height:1.6;'>此邮件由系统自动发送，请勿回复。</p>" +
                "</div>";
    }
}