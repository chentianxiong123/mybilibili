package com.mybilibili.user.service;

import com.mybilibili.user.entity.OperationTask;

import java.util.List;

public interface OperationTaskService {
    List<OperationTask> getTasksByCondition(String taskType, String status, String targetKeyword,
                                            String keyword, String startTime, String endTime,
                                            Integer page, Integer size);

    Integer countByCondition(String taskType, String status, String targetKeyword,
                             String keyword, String startTime, String endTime);

    OperationTask getById(Long id);
}
