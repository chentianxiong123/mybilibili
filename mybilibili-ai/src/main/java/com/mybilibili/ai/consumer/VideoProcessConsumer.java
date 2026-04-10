package com.mybilibili.ai.consumer;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.service.AiSubtitleService;
import com.mybilibili.ai.service.AiSummaryService;
import com.mybilibili.ai.service.AudioExtractService;
import com.mybilibili.ai.service.VideoTranscodeService;
import com.mybilibili.ai.utils.SubtitleTextUtils;
import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;
import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.MQConstants;
import com.mybilibili.mq.VideoProcessMessage;
import com.mybilibili.ai.mapper.VideoMapper;
import org.apache.rocketmq.client.producer.DefaultMQProducer;
import org.apache.rocketmq.common.message.Message;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.apache.rocketmq.spring.annotation.ConsumeMode;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.data.redis.core.StringRedisTemplate;

@Component
@RocketMQMessageListener(
    topic = MQConstants.TOPIC_VIDEO_PROCESS,
    consumerGroup = MQConstants.GROUP_VIDEO_PROCESS,
    consumeMode = ConsumeMode.ORDERLY
)
public class VideoProcessConsumer implements RocketMQListener<VideoProcessMessage> {

    @Autowired
    private VideoTranscodeService videoTranscodeService;

    @Autowired
    private AudioExtractService audioExtractService;

    @Autowired
    private AiSubtitleService aiSubtitleService;

    @Autowired
    private AiSummaryService aiSummaryService;

    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private ObjectMapper objectMapper;

    @Value("${rocketmq.name-server:127.0.0.1:9876}")
    private String rocketmqNameserver;

    private DefaultMQProducer producer = new DefaultMQProducer("video-process-producer-group");

    public void init() {
        producer.setNamesrvAddr(rocketmqNameserver);
        try {
            producer.start();
            System.out.println("RocketMQ Producer启动成功");
        } catch (Exception e) {
            System.err.println("RocketMQ Producer启动失败: " + e.getMessage());
        }
    }

    public void shutdown() {
        if (producer != null) {
            producer.shutdown();
        }
    }

    private void sendMessage(VideoProcessMessage message) {
        try {
            Message msg = new Message(
                MQConstants.TOPIC_VIDEO_PROCESS,
                message.getProcessType(),
                objectMapper.writeValueAsBytes(message)
            );
            producer.send(msg);
            System.out.println("[MQ] 发送消息成功: " + message.getProcessType() + ", videoId: " + message.getVideoId());
        } catch (Exception e) {
            System.err.println("[MQ] 发送消息失败: " + e.getMessage());
        }
    }

    @Override
    public void onMessage(VideoProcessMessage message) {
        System.out.println("========================================");
        System.out.println("[MQ消费者] 收到消息: " + message.getProcessType());
        System.out.println("[MQ消费者] 稿件ID: " + message.getManuscriptId());
        System.out.println("[MQ消费者] 视频ID: " + message.getVideoId());
        System.out.println("========================================");

        Integer videoId = message.getVideoId();
        Integer manuscriptId = message.getManuscriptId();
        String processType = message.getProcessType();

        Video video = videoMapper.selectById(videoId);
        String videoTitle = video != null ? video.getTitle() : "未知视频";

        try {
            broadcastProgress(videoId, manuscriptId, videoTitle, "STARTING", 0, null);

            switch (processType) {
                case VideoProcessMessage.PROCESS_TYPE_TRANSCODE:
                    processTranscode(manuscriptId, videoId, videoTitle);
                    break;
                case VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO:
                    processExtractAudio(manuscriptId, videoId, videoTitle);
                    break;
                case VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE:
                    processGenerateSubtitle(manuscriptId, videoId, videoTitle);
                    break;
                case VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY:
                    processAiSummary(manuscriptId, videoId, videoTitle);
                    break;
                default:
                    System.err.println("[MQ消费者] 未知的处理类型: " + processType);
            }

            System.out.println("[MQ消费者] 处理完成");

        } catch (Exception e) {
            System.err.println("[MQ消费者] 处理异常: " + e.getMessage());
            e.printStackTrace();
            throw e;
        }
    }

    private void sendNextMessage(Integer manuscriptId, Integer videoId, String currentProcessType) {
        String nextProcessType = null;

        switch (currentProcessType) {
            case VideoProcessMessage.PROCESS_TYPE_TRANSCODE:
                nextProcessType = VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO;
                break;
            case VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO:
                nextProcessType = VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE;
                break;
            case VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE:
                nextProcessType = VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY;
                break;
        }

        if (nextProcessType != null) {
            VideoProcessMessage nextMessage = VideoProcessMessage.of(manuscriptId, videoId, null, nextProcessType);
            sendMessage(nextMessage);
            System.out.println("[MQ] 自动发送下一阶段消息: " + nextProcessType + ", videoId: " + videoId);
        }
    }

    private void processTranscode(Integer manuscriptId, Integer videoId, String videoTitle) {
        System.out.println("[转码] 开始转码...");
        broadcastProgress(videoId, manuscriptId, videoTitle, "TRANSCODING", 10, Video.PROCESS_STATUS_TRANSCODING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);

        boolean success = videoTranscodeService.transcode(manuscriptId, videoId);

        if (success) {
            broadcastProgress(videoId, manuscriptId, videoTitle, "TRANSCODE_SUCCESS", 30, Video.PROCESS_STATUS_TRANSCODE_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_SUCCESS);
            System.out.println("[转码] 转码成功");

            sendNextMessage(manuscriptId, videoId, VideoProcessMessage.PROCESS_TYPE_TRANSCODE);
        } else {
            broadcastError(videoId, manuscriptId, videoTitle, "TRANSCODE", "转码失败");
            updateVideoStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_FAILED);
            System.err.println("[转码] 转码失败");
        }
    }

    private void processExtractAudio(Integer manuscriptId, Integer videoId, String videoTitle) {
        System.out.println("[音频提取] 开始提取音频...");
        broadcastProgress(videoId, manuscriptId, videoTitle, "AUDIO_EXTRACTING", 35, Video.PROCESS_STATUS_AUDIO_EXTRACTING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_AUDIO_EXTRACTING);

        boolean success = audioExtractService.extractAudio(manuscriptId, videoId);

        if (success) {
            broadcastProgress(videoId, manuscriptId, videoTitle, "AUDIO_SUCCESS", 50, Video.PROCESS_STATUS_AUDIO_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AUDIO_SUCCESS);
            System.out.println("[音频提取] 提取成功");

            sendNextMessage(manuscriptId, videoId, VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO);
        } else {
            broadcastError(videoId, manuscriptId, videoTitle, "EXTRACT_AUDIO", "音频提取失败");
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AUDIO_FAILED);
            System.err.println("[音频提取] 提取失败");
        }
    }

    private void processGenerateSubtitle(Integer manuscriptId, Integer videoId, String videoTitle) {
        System.out.println("[字幕生成] 开始生成字幕...");
        broadcastProgress(videoId, manuscriptId, videoTitle, "SUBTITLE_GENERATING", 55, Video.PROCESS_STATUS_SUBTITLE_GENERATING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_GENERATING);

        boolean success = aiSubtitleService.generateSubtitle(manuscriptId, videoId);

        if (success) {
            String subtitlePath = aiSubtitleService.getSubtitlePath(manuscriptId, videoId);
            try {
                String subtitleContent = SubtitleTextUtils.extractPlainText(subtitlePath);
                redisTemplate.opsForValue().set("subtitle:" + videoId, subtitleContent);
            } catch (Exception e) {
                System.err.println("[字幕生成] 读取字幕文件失败: " + e.getMessage());
            }

            broadcastProgress(videoId, manuscriptId, videoTitle, "SUBTITLE_SUCCESS", 80, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
            updateVideoHasSubtitle(videoId, 1);
            System.out.println("[字幕生成] 生成成功");

            sendNextMessage(manuscriptId, videoId, VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE);
        } else {
            broadcastError(videoId, manuscriptId, videoTitle, "GENERATE_SUBTITLE", "字幕生成失败");
            updateVideoStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
            System.err.println("[字幕生成] 生成失败");
        }
    }

    private void processAiSummary(Integer manuscriptId, Integer videoId, String videoTitle) {
        System.out.println("[AI摘要] 开始生成摘要...");
        broadcastProgress(videoId, manuscriptId, videoTitle, "AI_SUMMARIZING", 85, Video.PROCESS_STATUS_AI_SUMMARIZING);
        updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_SUMMARIZING);

        try {
            Video video = videoMapper.selectById(videoId);
            String subtitlePath = getSubtitlePath(manuscriptId, videoId);
            String subtitlePlainText = "";

            java.io.File subtitleFile = new java.io.File(subtitlePath);
            if (subtitleFile.exists()) {
                subtitlePlainText = SubtitleTextUtils.extractPlainText(subtitlePath);
            }

            String title = video != null ? video.getTitle() : "未知视频";
            String summary = aiSummaryService.generateSummary(subtitlePlainText, title, "");

            redisTemplate.opsForValue().set("summary:" + videoId, summary);

            broadcastProgress(videoId, manuscriptId, videoTitle, "AI_SUCCESS", 100, Video.PROCESS_STATUS_COMPLETED);
            updateVideoStatus(videoId, Video.PROCESS_STATUS_COMPLETED);
            updateVideoHasSummary(videoId, 1);
            
            VideoProcessWebSocketHandler.broadcastComplete(videoId, manuscriptId, videoTitle);
            System.out.println("[AI摘要] 生成成功");

        } catch (Exception e) {
            broadcastError(videoId, manuscriptId, videoTitle, "AI_SUMMARY", e.getMessage());
            updateVideoStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
            System.err.println("[AI摘要] 生成失败: " + e.getMessage());
        }
    }

    private void broadcastProgress(Integer videoId, Integer manuscriptId, String videoTitle, 
                                    String stage, int progress, Integer status) {
        String stageText = getStageText(stage);
        VideoProcessWebSocketHandler.broadcastProgress(videoId, manuscriptId, videoTitle, stage, stageText, progress, status);
        System.out.println("[进度] " + videoTitle + " - " + stageText + " (" + progress + "%)");
    }

    private void broadcastError(Integer videoId, Integer manuscriptId, String videoTitle, 
                                 String stage, String error) {
        VideoProcessWebSocketHandler.broadcastError(videoId, manuscriptId, videoTitle, stage, error);
    }

    private String getStageText(String stage) {
        switch (stage) {
            case "STARTING": return "启动中";
            case "TRANSCODING": return "视频转码中";
            case "TRANSCODE_SUCCESS": return "转码完成";
            case "AUDIO_EXTRACTING": return "音频提取中";
            case "AUDIO_SUCCESS": return "音频提取完成";
            case "SUBTITLE_GENERATING": return "字幕生成中";
            case "SUBTITLE_SUCCESS": return "字幕生成完成";
            case "AI_SUMMARIZING": return "AI摘要生成中";
            case "AI_SUCCESS": return "AI摘要完成";
            default: return stage;
        }
    }

    private void updateVideoStatus(Integer videoId, Integer status) {
        try {
            videoMapper.updateProcessStatus(videoId, status);
            System.out.println("[状态更新] 视频 " + videoId + " 状态更新为: " + status);
        } catch (Exception e) {
            System.err.println("[状态更新] 更新状态失败: " + e.getMessage());
        }
    }

    private void updateVideoHasSubtitle(Integer videoId, Integer hasSubtitle) {
        try {
            videoMapper.updateHasSubtitle(videoId, hasSubtitle);
        } catch (Exception e) {
            System.err.println("[状态更新] 更新字幕标记失败: " + e.getMessage());
        }
    }

    private void updateVideoHasSummary(Integer videoId, Integer hasSummary) {
        try {
            videoMapper.updateHasSummary(videoId, hasSummary);
        } catch (Exception e) {
            System.err.println("[状态更新] 更新摘要标记失败: " + e.getMessage());
        }
    }

    private String getSubtitlePath(Integer manuscriptId, Integer videoId) {
        return "d:/files/mybilibili/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/subtitles/zh-CN.srt";
    }
}
