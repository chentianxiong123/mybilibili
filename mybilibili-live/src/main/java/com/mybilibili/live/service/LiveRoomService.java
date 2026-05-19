package com.mybilibili.live.service;

import com.mybilibili.live.entity.LiveRoom;

import java.util.List;

public interface LiveRoomService {
    LiveRoom createRoom(Integer userId, String roomName);
    LiveRoom getByUserId(Integer userId);
    LiveRoom getById(Integer id);
    LiveRoom getByStreamKey(String streamKey);
    List<LiveRoom> getLiveList();
    void updateStatus(Integer id, String status);
    void updateViewerCount(Integer id, int delta);
}