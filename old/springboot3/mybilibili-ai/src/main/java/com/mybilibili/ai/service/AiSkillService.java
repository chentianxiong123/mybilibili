package com.mybilibili.ai.service;

import com.mybilibili.ai.entity.AiSkill;
import java.util.List;

public interface AiSkillService {
    List<AiSkill> listAll();
    List<AiSkill> listByType(String type);
    AiSkill getById(Long id);
    AiSkill getEnabledByType(String type);
    AiSkill create(AiSkill skill);
    AiSkill update(Long id, AiSkill skill);
    void delete(Long id);
    void toggleEnabled(Long id);
}