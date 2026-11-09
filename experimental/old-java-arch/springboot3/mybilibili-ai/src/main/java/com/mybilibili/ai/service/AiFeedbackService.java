package com.mybilibili.ai.service;

import com.mybilibili.ai.entity.AiFeedback;
import java.util.List;

public interface AiFeedbackService {
    List<AiFeedback> listAll();
    List<AiFeedback> listByStatus(String status);
    AiFeedback getById(Long id);
    AiFeedback create(AiFeedback feedback);
    AiFeedback update(Long id, AiFeedback feedback);
    void delete(Long id);
    void process(Long id, String adminReply);
}