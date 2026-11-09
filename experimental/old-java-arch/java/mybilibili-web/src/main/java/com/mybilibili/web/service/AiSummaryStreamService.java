package com.mybilibili.web.service;

import org.springframework.stereotype.Service;

/**
 * AI摘要流式输出服务
 * 用于处理AI摘要的流式输出逻辑
 */
@Service
public class AiSummaryStreamService {

    /**
     * 计算打字延迟时间
     * 根据内容类型返回不同的延迟，模拟真实的AI打字效果
     *
     * @param content 要发送的内容
     * @return 延迟时间（毫秒）
     */
    public int calculateTypeDelay(String content) {
        if (content == null || content.isEmpty()) {
            return 30;
        }

        // 标点符号后面稍微停顿一下
        char lastChar = content.charAt(content.length() - 1);
        if (lastChar == '。' || lastChar == '！' || lastChar == '？' || lastChar == '\n') {
            return 80 + (int)(Math.random() * 70);
        }

        // 逗号、分号后面短暂停顿
        if (lastChar == '，' || lastChar == '；' || lastChar == '：') {
            return 50 + (int)(Math.random() * 40);
        }

        // 普通字符快速输出
        return 25 + (int)(Math.random() * 35);
    }

    /**
     * 计算思考停顿时间
     * 在特定位置模拟AI思考停顿
     *
     * @param position 当前位置
     * @param totalLength 总长度
     * @return 是否需要停顿
     */
    public boolean shouldPause(int position, int totalLength) {
        // 在段落结束时停顿
        if (position % 100 == 0 && position > 0) {
            return Math.random() > 0.6;
        }

        // 在内容中间位置停顿
        if (position == totalLength / 2) {
            return true;
        }

        return false;
    }

    /**
     * 获取思考停顿时间
     *
     * @return 停顿时间（毫秒）
     */
    public int getPauseDuration() {
        return 150 + (int)(Math.random() * 200);
    }
}
