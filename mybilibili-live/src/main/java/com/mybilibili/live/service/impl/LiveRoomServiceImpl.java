package com.mybilibili.live.service.impl;

import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveRoomMapper;
import com.mybilibili.live.service.LiveRoomService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;

@Service
public class LiveRoomServiceImpl implements LiveRoomService {

    @Autowired
    private LiveRoomMapper liveRoomMapper;

    @Override
    public LiveRoom createRoom(Integer userId, String roomName) {
        LiveRoom room = new LiveRoom();
        room.setUserId(userId);
        room.setRoomName(roomName != null ? roomName : "我的直播间");
        room.setStreamKey(UUID.randomUUID().toString().replace("-", ""));
        room.setStatus("offline");
        room.setViewerCount(0);
        liveRoomMapper.insert(room);
        return room;
    }

    @Override
    public LiveRoom getByUserId(Integer userId) {
        return liveRoomMapper.selectByUserId(userId);
    }

    @Override
    public LiveRoom getById(Integer id) {
        return liveRoomMapper.selectById(id);
    }

    @Override
    public LiveRoom getByStreamKey(String streamKey) {
        return liveRoomMapper.selectByStreamKey(streamKey);
    }

    @Override
    public List<LiveRoom> getLiveList() {
        return liveRoomMapper.selectLiveList();
    }

    @Override
    public void updateStatus(Integer id, String status) {
        LiveRoom room = new LiveRoom();
        room.setId(id);
        room.setStatus(status);
        liveRoomMapper.updateById(room);
    }

    @Override
    public void updateViewerCount(Integer id, int delta) {
        LiveRoom room = liveRoomMapper.selectById(id);
        if (room != null) {
            int count = Math.max(0, (room.getViewerCount() != null ? room.getViewerCount() : 0) + delta);
            room.setViewerCount(count);
            liveRoomMapper.updateById(room);
        }
    }
}