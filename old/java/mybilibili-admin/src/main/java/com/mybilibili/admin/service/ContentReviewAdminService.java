package com.mybilibili.admin.service;

import com.mybilibili.common.vo.Result;

import java.util.List;
import java.util.Map;

public interface ContentReviewAdminService {
    Result<?> getPendingList(String contentType, Integer page, Integer size);
    Result<?> getAllContent(String contentType, String status, Integer page, Integer size);
    Result<?> restoreContent(String type, Integer id);
    Result<?> deleteContent(String type, Integer id);
    Result<?> batchProcess(String action, List<Map<String, Object>> items);
}
