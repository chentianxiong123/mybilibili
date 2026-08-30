package com.mybilibili.ai.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.ai.entity.AiFeedback;
import com.mybilibili.ai.mapper.AiFeedbackMapper;
import com.mybilibili.ai.service.AiFeedbackService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.Date;
import java.util.List;

@Service
public class AiFeedbackServiceImpl implements AiFeedbackService {

    @Autowired
    private AiFeedbackMapper aiFeedbackMapper;

    @Override
    public List<AiFeedback> listAll() {
        return aiFeedbackMapper.selectList(new LambdaQueryWrapper<AiFeedback>()
                .orderByDesc(AiFeedback::getCreatedAt));
    }

    @Override
    public List<AiFeedback> listByStatus(String status) {
        return aiFeedbackMapper.selectList(new LambdaQueryWrapper<AiFeedback>()
                .eq(AiFeedback::getStatus, status)
                .orderByDesc(AiFeedback::getCreatedAt));
    }

    @Override
    public AiFeedback getById(Long id) {
        return aiFeedbackMapper.selectById(id);
    }

    @Override
    public AiFeedback create(AiFeedback feedback) {
        feedback.setCreatedAt(new Date());
        feedback.setUpdatedAt(new Date());
        if (feedback.getStatus() == null) {
            feedback.setStatus("PENDING");
        }
        aiFeedbackMapper.insert(feedback);
        return feedback;
    }

    @Override
    public AiFeedback update(Long id, AiFeedback feedback) {
        feedback.setId(id);
        feedback.setUpdatedAt(new Date());
        aiFeedbackMapper.updateById(feedback);
        return feedback;
    }

    @Override
    public void delete(Long id) {
        aiFeedbackMapper.deleteById(id);
    }

    @Override
    public void process(Long id, String adminReply) {
        AiFeedback feedback = aiFeedbackMapper.selectById(id);
        if (feedback != null) {
            feedback.setAdminReply(adminReply);
            feedback.setStatus("PROCESSED");
            feedback.setUpdatedAt(new Date());
            aiFeedbackMapper.updateById(feedback);
        }
    }
}