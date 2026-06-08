package com.mybilibili.ai.consumer;

import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.ai.process.ProcessMode;
import com.mybilibili.ai.process.VideoProcessContext;
import com.mybilibili.ai.process.VideoProcessOrchestrator;
import com.mybilibili.ai.process.VideoProcessStepType;
import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoProcessMessage;
import org.apache.rocketmq.spring.annotation.ConsumeMode;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
@RocketMQMessageListener(
    topic = MQConstants.TOPIC_VIDEO_PROCESS,
    consumerGroup = MQConstants.GROUP_VIDEO_AI_PROCESS,
    selectorExpression = VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE + " || " + VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY,
    consumeMode = ConsumeMode.ORDERLY
)
public class VideoAiProcessConsumer implements RocketMQListener<VideoProcessMessage> {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private VideoProcessOrchestrator orchestrator;

    @Override
    public void onMessage(VideoProcessMessage message) {
        Integer videoId = message.getVideoId();
        Video video = videoMapper.selectById(videoId);
        String videoTitle = video != null ? video.getTitle() : "未知视频";

        VideoProcessContext context = new VideoProcessContext();
        context.setVideoId(videoId);
        context.setManuscriptId(message.getManuscriptId());
        context.setUserId(message.getUploaderId());
        context.setVideoTitle(videoTitle);
        context.setVideoDescription(video != null ? video.getDescription() : null);
        context.setCurrentStep(VideoProcessStepType.fromMqProcessType(message.getProcessType()));
        context.setProcessMode(resolveProcessMode(message.getProcessMode()));
        context.setTriggerSource("MQ");

        orchestrator.executeStep(context);
    }

    private ProcessMode resolveProcessMode(String processMode) {
        if (VideoProcessMessage.PROCESS_MODE_MANUAL_SINGLE.equals(processMode)) {
            return ProcessMode.MANUAL_SINGLE;
        }
        return ProcessMode.AUTO_CHAIN;
    }
}
