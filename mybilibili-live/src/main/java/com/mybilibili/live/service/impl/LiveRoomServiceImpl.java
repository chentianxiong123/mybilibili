package com.mybilibili.live.service.impl;

import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveRoomMapper;
import com.mybilibili.live.service.LiveRoomService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;

import java.util.Date;
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
        room.setRoomName(normalizeRoomName(roomName));
        room.setStreamKey(UUID.randomUUID().toString().replace("-", ""));
        room.setStatus(LiveConstants.RoomStatus.OFFLINE);
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
        if (!LiveConstants.RoomStatus.isValid(status)) {
            throw new BusinessException("直播间状态无效");
        }
        LiveRoom current = liveRoomMapper.selectById(id);
        if (current == null) {
            throw new BusinessException("直播间不存在");
        }
        LiveRoom room = new LiveRoom();
        room.setId(id);
        room.setStatus(status);
        // 开播或下播时把观众数清零（开播是新一场，下播也避免下次显示残留）
        room.setViewerCount(0);
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
        if (room == null || room.getId() == null) {
            throw new BusinessException("直播间不存在");
        }
        if (room.getRoomName() != null) {
            if (!StringUtils.hasText(room.getRoomName())) {
                throw new BusinessException("直播间名称不能为空");
            }
            room.setRoomName(room.getRoomName().trim());
        }
        if (room.getCoverUrl() != null) {
            room.setCoverUrl(blankToNull(room.getCoverUrl()));
        }
        if (room.getCategory() != null) {
            room.setCategory(blankToNull(room.getCategory()));
        }
        liveRoomMapper.updateById(room);
    }

    @Override
    public void scheduleRoom(Integer id, Date scheduledAt) {
        LiveRoom room = liveRoomMapper.selectById(id);
        if (room == null) {
            throw new BusinessException("直播间不存在");
        }
        if (scheduledAt != null && scheduledAt.before(new Date())) {
            throw new BusinessException("定时开播时间必须晚于当前时间");
        }
        room.setScheduledAt(scheduledAt);
        liveRoomMapper.updateById(room);
    }

    @Override
    public IPage<LiveRoom> listAll(int page, int size, String status) {
        if (StringUtils.hasText(status) && !LiveConstants.RoomStatus.isValid(status)) {
            throw new BusinessException("直播间状态无效");
        }
        Page<LiveRoom> pageObj = new Page<>(Math.max(page, 1), Math.min(Math.max(size, 1), 100));
        LambdaQueryWrapper<LiveRoom> wrapper = new LambdaQueryWrapper<>();
        if (StringUtils.hasText(status)) {
            wrapper.eq(LiveRoom::getStatus, status);
        }
        wrapper.orderByDesc(LiveRoom::getViewerCount);
        return liveRoomMapper.selectPage(pageObj, wrapper);
    }

    @Override
    public Map<String, Object> getAdminStats() {
        long totalRooms = liveRoomMapper.selectCount(null);
        LambdaQueryWrapper<LiveRoom> liveWrapper = new LambdaQueryWrapper<>();
        liveWrapper.eq(LiveRoom::getStatus, LiveConstants.RoomStatus.LIVE);
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

    private String normalizeRoomName(String roomName) {
        return StringUtils.hasText(roomName) ? roomName.trim() : "我的直播间";
    }

    private String blankToNull(String value) {
        return StringUtils.hasText(value) ? value.trim() : null;
    }
}
