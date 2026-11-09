package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.entity.AiSkill;
import com.mybilibili.ai.mapper.AiSkillMapper;
import com.mybilibili.ai.service.AiSkillService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.Date;
import java.util.List;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;

@Service
public class AiSkillServiceImpl implements AiSkillService {

    @Autowired
    private AiSkillMapper aiSkillMapper;

    @Override
    public List<AiSkill> listAll() {
        return aiSkillMapper.selectList(null);
    }

    @Override
    public List<AiSkill> listByType(String type) {
        return aiSkillMapper.selectList(
                new LambdaQueryWrapper<AiSkill>()
                        .eq(AiSkill::getType, type)
        );
    }

    @Override
    public AiSkill getById(Long id) {
        return aiSkillMapper.selectById(id);
    }

    @Override
    public AiSkill getEnabledByType(String type) {
        return aiSkillMapper.selectOne(
                new LambdaQueryWrapper<AiSkill>()
                        .eq(AiSkill::getType, type)
                        .eq(AiSkill::getEnabled, true)
        );
    }

    @Override
    public AiSkill create(AiSkill skill) {
        skill.setEnabled(true);
        skill.setCreatedAt(new Date());
        skill.setUpdatedAt(new Date());
        aiSkillMapper.insert(skill);
        return skill;
    }

    @Override
    public AiSkill update(Long id, AiSkill skill) {
        skill.setId(id);
        skill.setUpdatedAt(new Date());
        aiSkillMapper.updateById(skill);
        return skill;
    }

    @Override
    public void delete(Long id) {
        aiSkillMapper.deleteById(id);
    }

    @Override
    public void toggleEnabled(Long id) {
        AiSkill skill = aiSkillMapper.selectById(id);
        if (skill != null) {
            skill.setEnabled(!skill.getEnabled());
            skill.setUpdatedAt(new Date());
            aiSkillMapper.updateById(skill);
        }
    }
}