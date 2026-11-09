package com.mybilibili.live.service.impl;

import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveRoomMapper;
import com.mybilibili.live.service.LiveRoomService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;

import java.util.List;
import java.util.Map;
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
        // 开播或下播时把观众数清零（开播是新一场，下播也避免下次显示残留）
        if ("live".equals(status) || "offline".equals(status)) {
            room.setViewerCount(0);
        }
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

    @Override
    public void updateRoom(LiveRoom room) {
        liveRoomMapper.updateById(room);
    }

    @Override
    public void scheduleRoom(Integer id, java.util.Date scheduledAt) {
        LiveRoom room = liveRoomMapper.selectById(id);
        if (room != null) {
            room.setScheduledAt(scheduledAt);
            liveRoomMapper.updateById(room);
        }
    }

    @Override
    public IPage<LiveRoom> listAll(int page, int size, String status) {
        Page<LiveRoom> pageObj = new Page<>(page, size);
        LambdaQueryWrapper<LiveRoom> wrapper = new LambdaQueryWrapper<>();
        if (status != null && !status.isEmpty()) {
            wrapper.eq(LiveRoom::getStatus, status);
        }
        wrapper.orderByDesc(LiveRoom::getViewerCount);
        return liveRoomMapper.selectPage(pageObj, wrapper);
    }

    @Override
    public Map<String, Object> getAdminStats() {
        long totalRooms = liveRoomMapper.selectCount(null);
        LambdaQueryWrapper<LiveRoom> liveWrapper = new LambdaQueryWrapper<>();
        liveWrapper.eq(LiveRoom::getStatus, "live");
        long onlineRooms = liveRoomMapper.selectCount(liveWrapper);
        // 统计所有在线直播间的总观众数
        List<LiveRoom> liveRooms = liveRoomMapper.selectList(liveWrapper);
        int totalViewers = liveRooms.stream()
                .mapToInt(r -> r.getViewerCount() != null ? r.getViewerCount() : 0)
                .sum();
        return Map.of(
                "totalRooms", totalRooms,
                "onlineRooms", onlineRooms,
                "totalViewers", totalViewers
        );
    }
}