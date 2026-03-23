package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.service.StatisticsService;
import com.mybilibili.admin.mapper.VideoMapper;
import com.mybilibili.admin.mapper.UserMapper;
import com.mybilibili.admin.mapper.CommentMapper;
import com.mybilibili.admin.mapper.ManuscriptMapper;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class StatisticsServiceImpl implements StatisticsService {

    @Autowired
    private VideoMapper videoMapper;

    @Autowired
    private UserMapper userMapper;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ManuscriptMapper manuscriptMapper;

    @Override
    public Map<String, Object> getVideoPlayStatistics(String startDate, String endDate) {
        Map<String, Object> result = new HashMap<>();
        List<Map<String, Object>> dailyData = videoMapper.getDailyPlayCount(startDate, endDate);
        result.put("dailyData", dailyData);
        Integer totalPlays = videoMapper.getTotalPlayCount(startDate, endDate);
        result.put("totalPlays", totalPlays);
        return result;
    }

    @Override
    public Map<String, Object> getUserGrowthStatistics(String startDate, String endDate) {
        Map<String, Object> result = new HashMap<>();
        List<Map<String, Object>> dailyData = userMapper.getDailyUserCount(startDate, endDate);
        result.put("dailyData", dailyData);
        Integer totalUsers = userMapper.getTotalUserCount();
        result.put("totalUsers", totalUsers);
        return result;
    }

    @Override
    public Map<String, Object> getCommentStatistics(String startDate, String endDate) {
        Map<String, Object> result = new HashMap<>();
        List<Map<String, Object>> dailyData = commentMapper.getDailyCommentCount(startDate, endDate);
        result.put("dailyData", dailyData);
        Integer totalComments = commentMapper.getTotalCommentCount(startDate, endDate);
        result.put("totalComments", totalComments);
        return result;
    }

    @Override
    public List<Map<String, Object>> getHotVideos(int limit) {
        return videoMapper.getHotVideos(limit);
    }

    @Override
    public Map<String, Object> getOverviewStatistics() {
        Map<String, Object> result = new HashMap<>();
        
        int userCount = userMapper.getTotalUserCount();
        result.put("userCount", userCount);
        
        int manuscriptCount = manuscriptMapper.countAll();
        result.put("manuscriptCount", manuscriptCount);
        
        int videoCount = videoMapper.countVideosByKeyword(null, null);
        result.put("videoCount", videoCount);
        
        Long totalViewCount = manuscriptMapper.sumTotalViewCount();
        result.put("viewCount", totalViewCount != null ? totalViewCount : 0L);
        
        int pendingCount = manuscriptMapper.countByStatus(0);
        result.put("pendingManuscriptCount", pendingCount);
        
        return result;
    }

    @Override
    public List<Map<String, Object>> getManuscriptStatusStatistics() {
        return manuscriptMapper.countGroupByStatus();
    }

    @Override
    public List<Map<String, Object>> getRecentManuscripts(int limit) {
        List<Map<String, Object>> result = new java.util.ArrayList<>();
        List<com.mybilibili.common.entity.Manuscript> manuscripts = manuscriptMapper.selectRecent(limit);
        for (com.mybilibili.common.entity.Manuscript m : manuscripts) {
            Map<String, Object> item = new HashMap<>();
            item.put("id", m.getId());
            item.put("title", m.getTitle());
            item.put("status", m.getStatus());
            item.put("uploadTime", m.getUploadTime());
            item.put("viewCount", m.getViewCount());
            result.add(item);
        }
        return result;
    }
}