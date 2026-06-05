package com.mybilibili.live.service;

import com.baomidou.mybatisplus.core.metadata.IPage;
import com.mybilibili.live.entity.LiveRoom;

import java.util.List;
import java.util.Map;

public interface LiveRoomService {
    LiveRoom createRoom(Integer userId, String roomName);
    LiveRoom getByUserId(Integer userId);
    LiveRoom getById(Integer id);
    LiveRoom getByStreamKey(String streamKey);
    List<LiveRoom> getLiveList();
    void updateStatus(Integer id, String status);
    void updateViewerCount(Integer id, int delta);
    void updateRoom(LiveRoom room);
    void scheduleRoom(Integer id, java.util.Date scheduledAt);
    IPage<LiveRoom> listAll(int page, int size, String status);
    Map<String, Object> getAdminStats();
}