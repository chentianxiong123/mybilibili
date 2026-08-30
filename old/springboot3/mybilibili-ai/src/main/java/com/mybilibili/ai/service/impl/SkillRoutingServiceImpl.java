package com.mybilibili.ai.service.impl;

import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.entity.AiSkill;
import com.mybilibili.ai.mapper.AiSkillMapper;
import com.mybilibili.ai.service.SkillRoutingService;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

/**
 * Skill routing service that uses LLM to determine which skill(s) to use for a user question.
 */
@Service
public class SkillRoutingServiceImpl implements SkillRoutingService {

    @Autowired
    private AiSkillMapper aiSkillMapper;

    @Autowired
    private DynamicChatClient dynamicChatClient;

    private static final String META_SYSTEM_PROMPT = "你是技能路由专家。根据用户问题，从以下技能列表中选择最适合的技能。\n" +
            "技能列表：\n" +
            "{SKILLS_LIST}\n" +
            "\n你只需要输出技能名称，多个技能用逗号分隔。\n" +
            "如果没有技能适合回答用户问题，请输出：NONE";

    @Override
    public AiSkill routeSkill(String userQuestion) {
        List<AiSkill> skills = routeSkills(userQuestion);
        return skills.isEmpty() ? null : skills.get(0);
    }

    @Override
    public List<AiSkill> routeSkills(String userQuestion) {
        // Load all enabled CUSTOMER_SERVICE skills
        List<AiSkill> allSkills = aiSkillMapper.selectList(
                new LambdaQueryWrapper<AiSkill>()
                        .eq(AiSkill::getType, "CUSTOMER_SERVICE")
                        .eq(AiSkill::getEnabled, true)
        );

        if (allSkills.isEmpty()) {
            return new ArrayList<>();
        }

        // Get ChatClient for CHAT feature
        ChatClient chatClient = dynamicChatClient.getClient("CHAT");
        if (chatClient == null) {
            return new ArrayList<>();
        }

        // Build skills description for routing prompt
        StringBuilder skillsListBuilder = new StringBuilder();
        for (AiSkill skill : allSkills) {
            skillsListBuilder.append("- 名称：").append(skill.getName());
            if (skill.getDescription() != null && !skill.getDescription().isBlank()) {
                skillsListBuilder.append("，描述：").append(skill.getDescription());
            }
            skillsListBuilder.append("\n");
        }
        String skillsList = skillsListBuilder.toString();

        // Build routing prompt
        String routingPrompt = META_SYSTEM_PROMPT.replace("{SKILLS_LIST}", skillsList) +
                "\n\n用户问题：【" + userQuestion + "】";

        // Call LLM for skill routing
        String response;
        try {
            response = chatClient.prompt()
                    .system(routingPrompt)
                    .user("请判断")
                    .call()
                    .content();
        } catch (Exception e) {
            return new ArrayList<>();
        }

        if (response == null || response.trim().isEmpty()) {
            return new ArrayList<>();
        }

        response = response.trim();

        // Check if no skill matches
        if (response.equalsIgnoreCase("NONE")) {
            return new ArrayList<>();
        }

        // Parse skill names from response
        List<AiSkill> matchedSkills = new ArrayList<>();
        String[] skillNames = response.split("[,，]");

        for (String skillName : skillNames) {
            String trimmedName = skillName.trim();
            if (trimmedName.isEmpty()) continue;

            // Case-insensitive match by skill name
            for (AiSkill skill : allSkills) {
                if (skill.getName().equalsIgnoreCase(trimmedName)) {
                    matchedSkills.add(skill);
                    break;
                }
            }
        }

        return matchedSkills;
    }
}