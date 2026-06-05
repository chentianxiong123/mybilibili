package com.mybilibili.user.service.impl;

import com.mybilibili.user.service.CaptchaService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.Random;
import java.util.concurrent.TimeUnit;

@Service
public class CaptchaServiceImpl implements CaptchaService {

    @Autowired
    private StringRedisTemplate redisTemplate;

    private static final String PREFIX = "captcha:";
    private static final long EXPIRE_MINUTES = 5;

    private final Random random = new Random();

    @Override
    public String generate(String captchaId) {
        int a = random.nextInt(10) + 1;   // 1-10
        int b = random.nextInt(10) + 1;   // 1-10
        int op = random.nextInt(3);       // 0=+, 1=-, 2=*

        String question;
        int answer;
        switch (op) {
            case 0: // 加法
                question = a + " + " + b + " = ?";
                answer = a + b;
                break;
            case 1: // 减法（结果非负）
                if (a < b) { int t = a; a = b; b = t; }
                question = a + " - " + b + " = ?";
                answer = a - b;
                break;
            default: // 乘法（结果 <= 50）
                a = random.nextInt(9) + 1;
                b = random.nextInt(9) + 1;
                question = a + " × " + b + " = ?";
                answer = a * b;
                break;
        }

        redisTemplate.opsForValue().set(PREFIX + captchaId, String.valueOf(answer), EXPIRE_MINUTES, TimeUnit.MINUTES);
        return question;
    }

    @Override
    public boolean verify(String captchaId, String answer) {
        if (captchaId == null || answer == null) return false;
        String key = PREFIX + captchaId;
        String correct = redisTemplate.opsForValue().get(key);
        if (correct == null) return false;
        boolean ok = correct.equals(answer.trim());
        if (ok) {
            redisTemplate.delete(key); // 一次性，验证后删除
        }
        return ok;
    }
}