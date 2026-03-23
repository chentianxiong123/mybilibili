package com.mybilibili.common.utils;

/**
 * 时长工具类
 */
public class DurationUtils {

    /**
     * 将秒数格式化为时长字符串 (如 384 -> "06:24")
     * @param seconds 秒数
     * @return 格式化后的时长字符串
     */
    public static String formatDuration(Integer seconds) {
        if (seconds == null || seconds <= 0) {
            return "00:00";
        }
        int minutes = seconds / 60;
        int secs = seconds % 60;
        return String.format("%02d:%02d", minutes, secs);
    }

    /**
     * 将秒数格式化为时长字符串 (带小时，如 7384 -> "02:03:04")
     * @param seconds 秒数
     * @return 格式化后的时长字符串
     */
    public static String formatDurationWithHour(Integer seconds) {
        if (seconds == null || seconds <= 0) {
            return "00:00";
        }
        int hours = seconds / 3600;
        int minutes = (seconds % 3600) / 60;
        int secs = seconds % 60;
        if (hours > 0) {
            return String.format("%02d:%02d:%02d", hours, minutes, secs);
        }
        return String.format("%02d:%02d", minutes, secs);
    }
}
