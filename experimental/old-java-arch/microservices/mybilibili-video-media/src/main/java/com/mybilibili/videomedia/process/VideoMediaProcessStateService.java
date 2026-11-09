package com.mybilibili.videomedia.process;

import com.mybilibili.common.entity.Video;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.mq.VideoProcessAnalyticsEvent;
import com.mybilibili.mq.VideoProcessProgressEvent;
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
            publishProgress(previous, videoId, status, progress, stage, stageText(status), error, VideoProcessProgressEvent.EVENT_PROGRESS);
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
            publishProgress(previous, videoId, Video.PROCESS_STATUS_TRANSCODE_SUCCESS, 100, "TRANSCODE_SUCCESS", "转码完成", null, VideoProcessProgressEvent.EVENT_PROGRESS);
        }
    }

    public void markFailed(Integer videoId, Integer failedStatus, String stage, String error) {
        Video previous = videoMapper.selectById(videoId);
        int updated = videoMapper.updateProcessError(videoId, failedStatus, error);
        if (updated > 0) {
            publishAnalytics(previous, videoId, failedStatus, 0, stage, error);
            publishProgress(previous, videoId, failedStatus, 0, stage, error, error, VideoProcessProgressEvent.EVENT_ERROR);
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

    private void publishProgress(Video previous,
                                 Integer videoId,
                                 Integer status,
                                 Integer progress,
                                 String stage,
                                 String stageText,
                                 String error,
                                 String eventName) {
        Integer manuscriptId = previous != null ? previous.getManuscriptId() : null;
        String title = previous != null ? previous.getTitle() : null;
        VideoProcessProgressEvent event = VideoProcessProgressEvent.of(
                videoId,
                manuscriptId,
                title,
                stage,
                stageText,
                progress,
                status,
                error,
                eventName,
                "mybilibili-video-media"
        );
        videoMQProducer.sendVideoProcessProgressEvent(event);
    }

    private String stageText(Integer status) {
        if (status == null) {
            return "处理中";
        }
        switch (status) {
            case Video.PROCESS_STATUS_TRANSCODING: return "视频转码中";
            case Video.PROCESS_STATUS_TRANSCODE_SUCCESS: return "转码完成";
            case Video.PROCESS_STATUS_AUDIO_EXTRACTING: return "音频提取中";
            case Video.PROCESS_STATUS_AUDIO_SUCCESS: return "音频提取完成";
            default: return "处理中";
        }
    }
}
