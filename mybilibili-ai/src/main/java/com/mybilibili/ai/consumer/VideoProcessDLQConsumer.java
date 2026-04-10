package com.mybilibili.ai.consumer;

import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.VideoProcessMessage;
import com.mybilibili.ai.mapper.VideoMapper;
import org.apache.rocketmq.spring.annotation.RocketMQMessageListener;
import org.apache.rocketmq.spring.core.RocketMQListener;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Component
@RocketMQMessageListener(
    topic = "%DLQ%video-process-group",
    consumerGroup = "video-process-dlq-group"
)
public class VideoProcessDLQConsumer implements RocketMQListener<VideoProcessMessage> {

    @Autowired
    private VideoMapper videoMapper;

    @Override
    public void onMessage(VideoProcessMessage message) {
        System.err.println("[死信队列] 收到最终失败的消息: " + message);
        
        Integer videoId = message.getVideoId();
        String processType = message.getProcessType();
        
        try {
            Integer processStatus;
            switch (processType) {
                case VideoProcessMessage.PROCESS_TYPE_TRANSCODE:
                    processStatus = Video.PROCESS_STATUS_TRANSCODE_FAILED;
                    break;
                case VideoProcessMessage.PROCESS_TYPE_EXTRACT_AUDIO:
                    processStatus = Video.PROCESS_STATUS_AUDIO_FAILED;
                    break;
                case VideoProcessMessage.PROCESS_TYPE_GENERATE_SUBTITLE:
                    processStatus = Video.PROCESS_STATUS_SUBTITLE_FAILED;
                    break;
                case VideoProcessMessage.PROCESS_TYPE_AI_SUMMARY:
                    processStatus = Video.PROCESS_STATUS_AI_FAILED;
                    break;
                case VideoProcessMessage.PROCESS_TYPE_ALL:
                    processStatus = Video.PROCESS_STATUS_TRANSCODE_FAILED;
                    break;
                default:
                    processStatus = Video.PROCESS_STATUS_TRANSCODE_FAILED;
            }
            
            videoMapper.updateProcessError(videoId, processStatus, "处理失败，已达到最大重试次数");
            System.err.println("[死信队列] 已更新视频 " + videoId + " 状态为失败");
            
        } catch (Exception e) {
            System.err.println("[死信队列] 处理死信消息异常: " + e.getMessage());
        }
    }
}
