package com.mybilibili.interaction.service;

import com.mybilibili.common.vo.WatchHistoryVO;

import java.util.List;

public interface WatchHistoryService {
    List<WatchHistoryVO> getWatchHistoryList(Integer userId, Integer page, Integer size);
    void recordWatchHistory(Integer userId, Integer videoId, Integer progressSeconds, Integer videoDuration);
    void clearWatchHistory(Integer userId);
    void deleteWatchHistory(Integer id, Integer userId);
}
