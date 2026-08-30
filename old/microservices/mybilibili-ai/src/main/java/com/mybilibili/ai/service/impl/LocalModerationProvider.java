package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.service.ModerationProvider;
import com.mybilibili.ai.service.ModerationProvider.ModerateRequest;
import com.mybilibili.ai.service.ModerationProvider.ModerationResult;
import org.springframework.stereotype.Component;

import java.util.List;

/**
 * 本地违禁词审核提供者。
 * 基于违禁词库和正则规则进行本地审核。
 */
@Component
public class LocalModerationProvider implements ModerationProvider {

    @Override
    public String getName() {
        return "local-moderation";
    }

    @Override
    public boolean isAvailable() {
        return true;
    }

    @Override
    public Object invoke(ModerateRequest request) {
        return moderate(request);
    }

    public ModerationResult moderate(ModerateRequest request) {
        String text = request.getText();
        if (text == null || text.isEmpty()) {
            return new ModerationResult(true, null);
        }

        // TODO: 从 Redis 或数据库读取违禁词库进行匹配
        // 这里暂时返回通过，实际应调用违禁词检测服务
        return new ModerationResult(true, null);
    }
}