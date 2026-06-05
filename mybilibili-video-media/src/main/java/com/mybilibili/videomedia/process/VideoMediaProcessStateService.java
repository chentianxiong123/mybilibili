package com.mybilibili.videomedia.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessAnalyticsEvent;
import com.mybilibili.video.mapper.VideoMapper;
import org.springframework.stereotype.Service;

@Service
public class VideoMediaProcessStateService {

    private final VideoMapper videoMapper;
    private final VideoMQProducer videoMQProducer;

    public VideoMediaProcessStateService(VideoMapper videoMapper, VideoMQProducer videoMQProducer) {
        this.videoMapper = videoMapper;
        this.videoMQProducer = videoMQProducer;
    }

    public void markProgress(Integer videoId, Integer status, Integer progress, String stage, String error) {
        Video previous = videoMapper.selectById(videoId);
        int updated = videoMapper.updateProcessState(videoId, status, progress, stage, error);
        if (updated > 0) {
            publishAnalytics(previous, videoId, status, progress, stage, error);
        }
    }

    public void markTranscodeSuccess(Integer videoId, String hdUrl, String sdUrl, String ldUrl) {
        Video previous = videoMapper.selectById(videoId);
        int updated = videoMapper.updateTranscodeResult(
                videoId,
                Video.PROCESS_STATUS_TRANSCODE_SUCCESS,
                100,
                "TRANSCODE_SUCCESS",
                hdUrl,
                sdUrl,
                ldUrl
        );
        if (updated > 0) {
            publishAnalytics(previous, videoId, Video.PROCESS_STATUS_TRANSCODE_SUCCESS, 100, "TRANSCODE_SUCCESS", null);
        }
    }

    public void markFailed(Integer videoId, Integer failedStatus, String stage, String error) {
        Video previous = videoMapper.selectById(videoId);
        int updated = videoMapper.updateProcessError(videoId, failedStatus, error);
        if (updated > 0) {
            publishAnalytics(previous, videoId, failedStatus, 0, stage, error);
        }
    }

    private void publishAnalytics(Video previous, Integer videoId, Integer toStatus, Integer progress, String stage, String error) {
        Integer manuscriptId = previous != null ? previous.getManuscriptId() : null;
        Integer fromStatus = previous != null ? previous.getProcessStatus() : null;
        VideoProcessAnalyticsEvent event = VideoProcessAnalyticsEvent.of(
                videoId,
                manuscriptId,
                fromStatus,
                toStatus,
                stage,
                progress,
                error,
                "SYSTEM",
                null
        );
        videoMQProducer.sendVideoProcessAnalyticsEvent(event);
    }
}
