package com.mybilibili.admin.utils;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.regex.Pattern;

/**
 * 字幕文本处理工具类
 * 用于提取SRT字幕中的纯文本内容
 */
public class SubtitleTextUtils {

    // SRT时间戳格式: 00:00:00,000 --> 00:00:00,000
    private static final Pattern TIME_PATTERN = Pattern.compile(
        "\\d{2}:\\d{2}:\\d{2},\\d{3}\\s*-->\\s*\\d{2}:\\d{2}:\\d{2},\\d{3}"
    );

    // 序号行格式: 纯数字
    private static final Pattern INDEX_PATTERN = Pattern.compile("^\\d+$");

    // HTML标签
    private static final Pattern HTML_TAG_PATTERN = Pattern.compile("<[^>]+>");

    // 默认最大文本长度（约4000个token，用于DeepSeek API）
    private static final int DEFAULT_MAX_LENGTH = 12000;

    /**
     * 从SRT文件中提取纯文本内容
     *
     * @param srtFilePath SRT文件路径
     * @return 纯文本内容
     * @throws IOException 读取文件失败时抛出
     */
    public static String extractPlainText(String srtFilePath) throws IOException {
        File file = new File(srtFilePath);
        if (!file.exists()) {
            throw new IOException("字幕文件不存在: " + srtFilePath);
        }

        StringBuilder content = new StringBuilder();
        try (BufferedReader reader = new BufferedReader(new FileReader(file))) {
            String line;
            while ((line = reader.readLine()) != null) {
                String trimmedLine = line.trim();

                // 跳过空行
                if (trimmedLine.isEmpty()) {
                    continue;
                }

                // 跳过序号行
                if (INDEX_PATTERN.matcher(trimmedLine).matches()) {
                    continue;
                }

                // 跳过时间戳行
                if (TIME_PATTERN.matcher(trimmedLine).matches()) {
                    continue;
                }

                // 去除HTML标签
                String cleanLine = HTML_TAG_PATTERN.matcher(trimmedLine).replaceAll("");

                // 添加到内容中
                if (!cleanLine.isEmpty()) {
                    if (content.length() > 0) {
                        content.append(" ");
                    }
                    content.append(cleanLine);
                }
            }
        }

        return content.toString().trim();
    }

    /**
     * 从SRT内容字符串中提取纯文本
     *
     * @param srtContent SRT格式内容
     * @return 纯文本内容
     */
    public static String extractPlainTextFromContent(String srtContent) {
        if (srtContent == null || srtContent.isEmpty()) {
            return "";
        }

        StringBuilder content = new StringBuilder();
        String[] lines = srtContent.split("\\r?\\n");

        for (String line : lines) {
            String trimmedLine = line.trim();

            // 跳过空行
            if (trimmedLine.isEmpty()) {
                continue;
            }

            // 跳过序号行
            if (INDEX_PATTERN.matcher(trimmedLine).matches()) {
                continue;
            }

            // 跳过时间戳行
            if (TIME_PATTERN.matcher(trimmedLine).matches()) {
                continue;
            }

            // 去除HTML标签
            String cleanLine = HTML_TAG_PATTERN.matcher(trimmedLine).replaceAll("");

            // 添加到内容中
            if (!cleanLine.isEmpty()) {
                if (content.length() > 0) {
                    content.append(" ");
                }
                content.append(cleanLine);
            }
        }

        return content.toString().trim();
    }

    /**
     * 截断文本到指定长度
     * 优先保留开头和结尾的内容，中间部分用省略号表示
     *
     * @param text 原始文本
     * @param maxLength 最大长度
     * @return 截断后的文本
     */
    public static String truncateText(String text, int maxLength) {
        if (text == null || text.length() <= maxLength) {
            return text;
        }

        // 如果文本超过最大长度，保留开头和结尾各40%，中间用省略号
        int keepLength = maxLength / 2 - 50; // 每边保留的长度
        String start = text.substring(0, keepLength);
        String end = text.substring(text.length() - keepLength);

        return start + "\n\n... [内容已省略，中间部分跳过] ...\n\n" + end;
    }

    /**
     * 截断文本到默认长度
     *
     * @param text 原始文本
     * @return 截断后的文本
     */
    public static String truncateText(String text) {
        return truncateText(text, DEFAULT_MAX_LENGTH);
    }

    /**
     * 清理并优化字幕文本
     * 去除多余空格、重复标点等
     *
     * @param text 原始文本
     * @return 清理后的文本
     */
    public static String cleanText(String text) {
        if (text == null || text.isEmpty()) {
            return "";
        }

        // 去除多余空格
        String cleaned = text.replaceAll("\\s+", " ");

        // 去除重复标点
        cleaned = cleaned.replaceAll("([。，！？；：])\\1+", "$1");

        // 去除空格+标点的组合
        cleaned = cleaned.replaceAll("\\s+([。，！？；：])", "$1");

        return cleaned.trim();
    }

    /**
     * 获取字幕文本的字符数统计
     *
     * @param text 字幕文本
     * @return 字符数
     */
    public static int getCharCount(String text) {
        return text == null ? 0 : text.length();
    }

    /**
     * 估算token数量（粗略估算：中文字符1个token，英文单词约0.75个token）
     *
     * @param text 文本
     * @return 估算的token数量
     */
    public static int estimateTokenCount(String text) {
        if (text == null || text.isEmpty()) {
            return 0;
        }

        int chineseCount = 0;
        int englishWordCount = 0;

        for (char c : text.toCharArray()) {
            if (Character.UnicodeBlock.of(c) == Character.UnicodeBlock.CJK_UNIFIED_IDEOGRAPHS
                || Character.UnicodeBlock.of(c) == Character.UnicodeBlock.CJK_COMPATIBILITY_IDEOGRAPHS
                || Character.UnicodeBlock.of(c) == Character.UnicodeBlock.CJK_UNIFIED_IDEOGRAPHS_EXTENSION_A) {
                chineseCount++;
            } else if (Character.isLetter(c)) {
                englishWordCount++;
            }
        }

        // 英文单词数 = 字母数 / 平均单词长度(约5)
        int englishWords = englishWordCount / 5;

        return chineseCount + englishWords;
    }
}
