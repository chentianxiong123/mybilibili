package com.mybilibili.comment.service;

import com.mybilibili.common.vo.Result;

import java.util.Map;

public interface ReportService {

    Result<?> submitReport(Integer reporterId, String targetType, Integer targetId,
                           Integer manuscriptId, String reason, String description);

    Result<?> getReportList(String status, String targetType, Integer page, Integer size);

    Result<?> processReport(Integer reportId, String action, String adminRemark, Integer adminId);

    Result<?> updateAiReviewResult(Map<String, Object> result);
}
