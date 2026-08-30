package com.mybilibili.admin.service;

import java.util.Map;
import java.util.List;

public interface StatisticsService {
    Map<String, Object> getVideoPlayStatistics(String startDate, String endDate);
    
    Map<String, Object> getUserGrowthStatistics(String startDate, String endDate);
    
    Map<String, Object> getCommentStatistics(String startDate, String endDate);
    
    List<Map<String, Object>> getHotVideos(int limit);
    
    Map<String, Object> getOverviewStatistics();
    
    List<Map<String, Object>> getManuscriptStatusStatistics();
    
    List<Map<String, Object>> getRecentManuscripts(int limit);
}