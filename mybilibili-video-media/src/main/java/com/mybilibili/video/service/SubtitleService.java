package com.mybilibili.video.service;

import com.mybilibili.common.entity.Subtitle;
import org.springframework.web.multipart.MultipartFile;

import java.util.List;
import java.util.Map;

public interface SubtitleService {

    List<Subtitle> getSubtitlesByVideoId(Integer videoId);

    List<Map<String, Object>> getVideosWithSubtitleInfo();

    Subtitle getSubtitleByVideoIdAndLanguage(Integer videoId, String language);

    Subtitle uploadSubtitle(Subtitle subtitle);

    Subtitle uploadSrtFile(Integer videoId, MultipartFile file, String language, Boolean isDefault, Integer uploadedBy);

    Subtitle parseAndSaveSrt(Integer videoId, String srtContent, String language, String languageName, Integer uploadedBy);

    Subtitle importSrtFilePath(Integer videoId, String srtFilePath, String language, Boolean isDefault, Integer uploadedBy);

    List<Subtitle> getPendingSubtitles();

    Subtitle approveSubtitle(String subtitleId, Integer reviewerId);

    Subtitle rejectSubtitle(String subtitleId, String reason, Integer reviewerId);

    Map<String, Object> previewSubtitle(String subtitleId);

    List<Map<String, Object>> scanSystemSubtitles(Integer videoId);

    Subtitle importSystemSubtitle(Integer videoId, String language, Boolean isDefault, Integer uploadedBy);

    void deleteSubtitle(String subtitleId);

    void setDefaultSubtitle(Integer videoId, String language);

    void setDefaultSubtitleById(String subtitleId);
}
