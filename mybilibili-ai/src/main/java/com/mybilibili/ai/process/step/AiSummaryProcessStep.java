package com.mybilibili.ai.process.step;

import com.mybilibili.ai.config.UploadFilePathConfig;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.StepExecutionResult;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessStep;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.ai.repository.SubtitleRepository;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.common.entity.Subtitle;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class AiSummaryProcessStep implements VideoProcessStep {

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private SubtitleRepository subtitleRepository;

    @Autowired
    private UploadFilePathConfig uploadFilePathConfig;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoMapper videoMapper;

    @Override
    public VideoProcessStepType getStepType() {
        return VideoProcessStepType.AI_SUMMARY;
    }

    @Override
    public StepExecutionResult execute(VideoProcessContext context) {
        List<Subtitle> subtitles = subtitleRepository.findByVideoId(context.getVideoId());
        if (subtitles == null || subtitles.isEmpty()) {
            return StepExecutionResult.fail("请先生成字幕");
        }

        StringBuilder subtitleText = new StringBuilder();
        for (Subtitle subtitle : subtitles) {
            if (subtitle.getContent() != null) {
                for (Subtitle.SubtitleItem item : subtitle.getContent()) {
                    if (item.getText() != null) {
                        subtitleText.append(item.getText()).append("\n");
                    }
                }
            }
        }

        String title = context.getVideoTitle() != null ? context.getVideoTitle() : "视频";
        String description = context.getVideoDescription() != null ? context.getVideoDescription() : "";
        String summary = aiSummaryService.generateSummary(subtitleText.toString(), title, description);
        if (summary == null || summary.isEmpty()) {
            return StepExecutionResult.fail("AI总结失败");
        }

        String summaryPath = uploadFilePathConfig.getAiSummaryPath(context.getManuscriptId(), context.getVideoId());
        boolean saved = aiSummaryService.saveSummaryToFile(summary, summaryPath, title);
        if (!saved) {
            return StepExecutionResult.fail("摘要保存失败");
        }

        redisTemplate.opsForValue().set("summary:" + context.getVideoId(), summary);
        videoMapper.updateHasSummary(context.getVideoId(), 1);
        return StepExecutionResult.success(summary);
    }
}
