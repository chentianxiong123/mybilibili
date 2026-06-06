package com.mybilibili.live.dto;

import com.mybilibili.common.exception.BusinessException;
import lombok.Data;

import java.time.LocalDateTime;
import java.time.format.DateTimeParseException;

@Data
public class ReserveMeetingRequest {
    private String roomName;
    private String scheduledStart;
    private String scheduledEnd;
    private String reason;

    public LocalDateTime requireScheduledStart() {
        return parseDateTime(scheduledStart);
    }

    public LocalDateTime requireScheduledEnd() {
        return parseDateTime(scheduledEnd);
    }

    private LocalDateTime parseDateTime(String value) {
        if (value == null || value.isBlank()) {
            throw new BusinessException("请选择预约时间段");
        }
        try {
            return LocalDateTime.parse(value);
        } catch (DateTimeParseException e) {
            throw new BusinessException("预约时间格式无效");
        }
    }
}
