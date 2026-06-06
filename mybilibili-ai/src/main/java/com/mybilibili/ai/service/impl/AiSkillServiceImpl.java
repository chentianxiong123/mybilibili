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

    private static final String TYPE_CUSTOMER_SERVICE = "CUSTOMER_SERVICE";

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
    public List<AiSkill> createMissingCustomerServiceDefaults() {
        List<AiSkill> created = new java.util.ArrayList<>();
        for (AiSkill skill : buildCustomerServiceDefaults()) {
            Long existingCount = aiSkillMapper.selectCount(
                    new LambdaQueryWrapper<AiSkill>()
                            .eq(AiSkill::getType, skill.getType())
                            .eq(AiSkill::getName, skill.getName())
            );
            if (existingCount != null && existingCount > 0) {
                continue;
            }
            Date now = new Date();
            skill.setEnabled(true);
            skill.setCreatedAt(now);
            skill.setUpdatedAt(now);
            aiSkillMapper.insert(skill);
            created.add(skill);
        }
        return created;
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

    private List<AiSkill> buildCustomerServiceDefaults() {
        return List.of(
                customerSkill(
                        "账号与登录",
                        "账号、登录、注册、密码、绑定手机、资料修改、账号异常、账号安全",
                        "你负责处理账号与登录相关咨询。优先给出可自助完成的步骤。涉及账号归属、封禁申诉、隐私验证、人工授权时，建议转人工客服。",
                        "用户：我忘记密码了怎么办？\n客服：请先在登录页选择忘记密码，通过绑定手机号或邮箱重置。若无法验证身份，可以转人工处理。\n\n用户：我的账号好像被盗了\n客服：请立即修改密码并检查绑定信息。如果无法登录或存在资产风险，我会建议转人工客服。"
                ),
                customerSkill(
                        "稿件审核",
                        "稿件、投稿、审核、驳回、下架、申诉、创作中心、审核状态",
                        "你负责解释稿件审核流程、常见驳回原因和申诉路径。不要编造具体审核结论；如果用户要求查看具体稿件状态或申诉处理，说明需要人工或后台数据确认。",
                        "用户：我的视频为什么审核不通过？\n客服：常见原因包括内容违规、版权风险、封面标题不规范等。请先查看创作中心的驳回原因；如果你认为误判，可以提交申诉。"
                ),
                customerSkill(
                        "视频播放",
                        "视频播放、播放失败、播放地址缺失、清晰度、字幕、摘要、弹幕、播放器",
                        "你负责处理视频播放体验问题。引导用户刷新、检查网络、切换清晰度、重新打开页面。涉及播放地址缺失、转码未完成、字幕摘要缺失时，说明可能是媒体处理任务未完成或异常。",
                        "用户：视频一会能看一会不能看\n客服：可能是播放地址生成或转码状态尚未稳定。请刷新页面后重试；如果持续出现播放地址缺失，建议联系人工排查该视频的媒体处理状态。"
                ),
                customerSkill(
                        "上传与处理",
                        "上传、分片上传、上传进度、转码、视频处理、音频提取、处理失败、重试",
                        "你负责解释上传和媒体处理链路。强调视频文件在对象存储中，后端任务通常只传递对象 key 和处理信号。遇到转码失败、进度卡住、重复处理时建议转人工或查看任务看板。",
                        "用户：上传进度条卡住了\n客服：请先确认网络稳定并不要重复关闭页面。如果文件已经上传到对象存储但处理进度卡住，可能是转码或任务队列异常，需要后台任务看板排查。"
                ),
                customerSkill(
                        "AI功能",
                        "AI字幕、AI摘要、AI审核、语音转文字、Whisper、客服机器人、模型能力",
                        "你负责解释平台 AI 功能边界。说明 AI 可能受模型配置、任务队列、文件状态影响。不要承诺百分百准确；涉及具体任务失败要建议查看任务状态或转人工。",
                        "用户：为什么没有生成摘要？\n客服：摘要依赖字幕或音频处理结果。如果转写、字幕或摘要任务尚未完成，页面可能暂时看不到摘要。可以稍后刷新，或转人工排查处理任务。"
                ),
                customerSkill(
                        "投诉与反馈",
                        "投诉、举报、建议、反馈、情绪激动、人工、客服、申诉、误判",
                        "你负责接收投诉和反馈。先安抚并收集必要信息；遇到用户明确要求人工、投诉升级、账号/审核争议、情绪激动时，应调用转人工工具。",
                        "用户：我要投诉，你们审核太离谱了\n客服：我理解你的不满。为了让人工客服继续核查，请提供稿件标题或相关页面信息，我会帮你转人工处理。"
                )
        );
    }

    private AiSkill customerSkill(String name, String description, String systemPrompt, String fewShotExamples) {
        AiSkill skill = new AiSkill();
        skill.setName(name);
        skill.setType(TYPE_CUSTOMER_SERVICE);
        skill.setDescription(description);
        skill.setSystemPrompt(systemPrompt);
        skill.setFewShotExamples(fewShotExamples);
        return skill;
    }
}
