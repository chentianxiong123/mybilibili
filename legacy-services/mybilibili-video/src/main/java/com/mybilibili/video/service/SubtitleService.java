package com.mybilibili.video.service;

import com.mybilibili.common.entity.Subtitle;

import java.util.List;

public interface SubtitleService {

    List<Subtitle> getSubtitlesByVideoId(Integer videoId);

    Subtitle getSubtitleByVideoIdAndLanguage(Integer videoId, String language);

    Subtitle uploadSubtitle(Subtitle subtitle);

    Subtitle parseAndSaveSrt(Integer videoId, String srtContent, String language, String languageName, Integer uploadedBy);

    void deleteSubtitle(String subtitleId);

    void setDefaultSubtitle(Integer videoId, String language);
}