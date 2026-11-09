package com.mybilibili.ai.process;

import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessMessage;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class VideoProcessDispatcherImpl implements VideoProcessDispatcher {

    @Autowired
    private VideoMQProducer videoMQProducer;

    @Override
    public void dispatchNext(VideoProcessContext context, VideoProcessStepType nextStep) {
        VideoProcessMessage message = VideoProcessMessage.of(
                context.getManuscriptId(),
                context.getVideoId(),
                context.getUserId(),
                nextStep.toMqProcessType()
        );
        videoMQProducer.sendVideoProcessMessage(message);
    }
}
