package com.mybilibili.ai.service;

public interface AiSummaryService {

    String generateSummary(String subtitleText, String videoTitle, String videoDescription);

    String generateSummary(String subtitleText, String videoTitle);

    boolean saveSummaryToFile(String summary, String filePath, String videoTitle);
}
