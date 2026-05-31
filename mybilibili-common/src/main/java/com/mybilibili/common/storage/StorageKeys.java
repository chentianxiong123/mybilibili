package com.mybilibili.common.storage;

public final class StorageKeys {

    private StorageKeys() {}

    public static String avatar(Integer userId) {
        return "avatars/%d/avatar.jpg".formatted(userId);
    }

    public static String manuscriptCover(Integer manuscriptId) {
        return "manuscripts/%d/cover.jpg".formatted(manuscriptId);
    }

    public static String videoSource(Integer manuscriptId, Integer videoId, String ext) {
        return "manuscripts/%d/videos/%d/source/video%s".formatted(manuscriptId, videoId, ext);
    }

    public static String videoHlsObject(Integer manuscriptId, Integer videoId, String quality, String fileName) {
        return "manuscripts/%d/videos/%d/transcoded/%s/%s".formatted(manuscriptId, videoId, quality, fileName);
    }

    public static String videoHlsPlaylist(Integer manuscriptId, Integer videoId, String quality) {
        return videoHlsObject(manuscriptId, videoId, quality, "playlist.m3u8");
    }

    public static String videoTranscoded(Integer manuscriptId, Integer videoId, String resolution) {
        return "manuscripts/%d/videos/%d/transcoded/%s.mp4".formatted(manuscriptId, videoId, resolution);
    }

    public static String videoHd(Integer manuscriptId, Integer videoId) {
        return videoTranscoded(manuscriptId, videoId, "1080p");
    }

    public static String videoSd(Integer manuscriptId, Integer videoId) {
        return videoTranscoded(manuscriptId, videoId, "720p");
    }

    public static String videoLd(Integer manuscriptId, Integer videoId) {
        return videoTranscoded(manuscriptId, videoId, "480p");
    }

    public static String videoAudio(Integer manuscriptId, Integer videoId) {
        return "manuscripts/%d/videos/%d/audio/audio.wav".formatted(manuscriptId, videoId);
    }

    public static String videoSubtitle(Integer manuscriptId, Integer videoId, String lang) {
        return "manuscripts/%d/videos/%d/subtitles/%s.srt".formatted(manuscriptId, videoId, lang);
    }

    public static String videoSummary(Integer manuscriptId, Integer videoId) {
        return "manuscripts/%d/videos/%d/summary/summary.txt".formatted(manuscriptId, videoId);
    }

    public static String bannerImage(String fileName) {
        return "images/banners/" + fileName;
    }

    public static String dynamicImage(Integer userId, String fileName) {
        return "images/dynamics/%d/%s".formatted(userId, fileName);
    }

    public static String tempUpload(String uploadId, String part, String chunk) {
        return "temp-uploads/%s/%s/%s".formatted(uploadId, part, chunk);
    }
}
