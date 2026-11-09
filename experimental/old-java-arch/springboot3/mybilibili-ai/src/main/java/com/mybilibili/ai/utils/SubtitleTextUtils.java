package com.mybilibili.ai.utils;

import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.IOException;
import java.util.regex.Pattern;

@Component
public class SubtitleTextUtils {

    private static final Pattern TIME_PATTERN = Pattern.compile(
        "\\d{2}:\\d{2}:\\d{2},\\d{3}\\s*-->\\s*\\d{2}:\\d{2}:\\d{2},\\d{3}"
    );

    private static final Pattern INDEX_PATTERN = Pattern.compile("^\\d+$");

    private static final Pattern HTML_TAG_PATTERN = Pattern.compile("<[^>]+>");

    private static final int DEFAULT_MAX_LENGTH = 12000;

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

                if (trimmedLine.isEmpty()) {
                    continue;
                }

                if (INDEX_PATTERN.matcher(trimmedLine).matches()) {
                    continue;
                }

                if (TIME_PATTERN.matcher(trimmedLine).matches()) {
                    continue;
                }

                String cleanLine = HTML_TAG_PATTERN.matcher(trimmedLine).replaceAll("");

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

    public static String extractPlainTextFromContent(String srtContent) {
        if (srtContent == null || srtContent.isEmpty()) {
            return "";
        }

        StringBuilder content = new StringBuilder();
        String[] lines = srtContent.split("\\r?\\n");

        for (String line : lines) {
            String trimmedLine = line.trim();

            if (trimmedLine.isEmpty()) {
                continue;
            }

            if (INDEX_PATTERN.matcher(trimmedLine).matches()) {
                continue;
            }

            if (TIME_PATTERN.matcher(trimmedLine).matches()) {
                continue;
            }

            String cleanLine = HTML_TAG_PATTERN.matcher(trimmedLine).replaceAll("");

            if (!cleanLine.isEmpty()) {
                if (content.length() > 0) {
                    content.append(" ");
                }
                content.append(cleanLine);
            }
        }

        return content.toString().trim();
    }

    public static String truncateText(String text, int maxLength) {
        if (text == null || text.length() <= maxLength) {
            return text;
        }

        int keepLength = maxLength / 2 - 50;
        String start = text.substring(0, keepLength);
        String end = text.substring(text.length() - keepLength);

        return start + "\n\n... [内容已省略，中间部分跳过] ...\n\n" + end;
    }

    public static String truncateText(String text) {
        return truncateText(text, DEFAULT_MAX_LENGTH);
    }

    public static String cleanText(String text) {
        if (text == null || text.isEmpty()) {
            return "";
        }

        String cleaned = text.replaceAll("\\s+", " ");
        cleaned = cleaned.replaceAll("([。，！？；：])\\1+", "$1");
        cleaned = cleaned.replaceAll("\\s+([。，！？；：])", "$1");

        return cleaned.trim();
    }

    public static int getCharCount(String text) {
        return text == null ? 0 : text.length();
    }

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

        int englishWords = englishWordCount / 5;

        return chineseCount + englishWords;
    }
}