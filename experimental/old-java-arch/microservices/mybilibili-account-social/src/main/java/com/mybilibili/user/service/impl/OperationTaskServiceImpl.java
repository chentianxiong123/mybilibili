package com.mybilibili.user.service.impl;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.user.entity.OperationTask;
import com.mybilibili.user.mapper.OperationTaskMapper;
import com.mybilibili.user.service.OperationTaskService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class OperationTaskServiceImpl implements OperationTaskService {

    @Autowired
    private OperationTaskMapper operationTaskMapper;

    @Override
    public List<OperationTask> getTasksByCondition(String taskType, String status, String targetKeyword,
                                                   String keyword, String startTime, String endTime,
                                                   Integer page, Integer size) {
        int normalizedPage = normalizePage(page);
        int normalizedSize = normalizeSize(size);
        int offset = (normalizedPage - 1) * normalizedSize;
        return operationTaskMapper.selectByCondition(taskType, status, targetKeyword, keyword,
                startTime, endTime, offset, normalizedSize);
    }

    @Override
    public Integer countByCondition(String taskType, String status, String targetKeyword,
                                    String keyword, String startTime, String endTime) {
        return operationTaskMapper.countByCondition(taskType, status, targetKeyword, keyword, startTime, endTime);
    }

    @Override
    public OperationTask getById(Long id) {
        if (id == null) {
            throw new BusinessException("任务ID不能为空");
        }
        return operationTaskMapper.selectById(id);
    }

    private int normalizePage(Integer page) {
        return page == null || page < 1 ? 1 : page;
    }

    private int normalizeSize(Integer size) {
        if (size == null || size < 1) {
            return 10;
        }
        return Math.min(size, 100);
    }
}
