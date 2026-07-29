package com.mybilibili.videomedia.process;

import org.springframework.stereotype.Service;

import java.nio.file.Files;
import java.nio.file.Path;

@Service
public class VideoMediaAudioService {

    private final VideoMediaProcessingStorageService storageService;
    private final FfmpegMediaTool ffmpeg;

    public VideoMediaAudioService(VideoMediaProcessingStorageService storageService, FfmpegMediaTool ffmpeg) {
        this.storageService = storageService;
        this.ffmpeg = ffmpeg;
    }

    public boolean extractAudio(Integer manuscriptId, Integer videoId) {
        try {
            Path sourceVideo = storageService.materializeSourceVideo(manuscriptId, videoId);
            Path audioPath = storageService.getAudioPath(manuscriptId, videoId);
            Files.createDirectories(audioPath.getParent());
            boolean success = ffmpeg.extractAudio(sourceVideo.toString(), audioPath.toString());
            if (!success) {
                return false;
            }
            storageService.uploadAudio(manuscriptId, videoId, audioPath);
            return true;
        } catch (Exception e) {
            return false;
        }
    }
}
