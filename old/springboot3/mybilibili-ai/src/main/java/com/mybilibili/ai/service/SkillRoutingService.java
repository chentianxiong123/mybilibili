package com.mybilibili.ai.service;

import com.mybilibili.ai.entity.AiSkill;
import java.util.List;

public interface SkillRoutingService {
    /**
     * Given a user question, return the best matching skill (or null if none match)
     */
    AiSkill routeSkill(String userQuestion);

    /**
     * Given a user question, return ALL matching skills
     */
    List<AiSkill> routeSkills(String userQuestion);
}