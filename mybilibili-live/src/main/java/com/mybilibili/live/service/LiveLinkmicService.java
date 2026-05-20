package com.mybilibili.live.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.live.entity.LiveLinkmic;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveLinkmicMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@Service
public class LiveLinkmicService {

    @Autowired
    private LiveLinkmicMapper linkmicMapper;

    @Autowired
    private LiveRoomService liveRoomService;

    private static final int MAX_LINKMICS = 3;

    public LiveLinkmic applyLinkmic(Long roomId, Long viewerId, String viewerName) {
        // 获取直播间信息
        LiveRoom liveRoom = liveRoomService.getById(roomId.intValue());
        if (liveRoom == null) {
            return null;
        }

        // 检查是否已经申请过（申请中或已连接的都视为已申请）
        LiveLinkmic existing = linkmicMapper.selectOne(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getViewerId, viewerId)
                .in(LiveLinkmic::getStatus, 0, 1));
        if (existing != null) {
            return existing;
        }

        // 满员也允许进队列；主播看到排队后可逐个同意
        LiveLinkmic linkmic = new LiveLinkmic();
        linkmic.setRoomId(roomId);
        linkmic.setStreamerId(Long.valueOf(liveRoom.getUserId()));
        linkmic.setViewerId(viewerId);
        linkmic.setViewerName(viewerName);
        linkmic.setStatus(0);
        linkmic.setAudioEnabled(0);
        linkmic.setVideoEnabled(0);
        linkmic.setMaxLinkmics(MAX_LINKMICS);
        linkmic.setApplyTime(LocalDateTime.now());
        linkmicMapper.insert(linkmic);
        return linkmic;
    }

    /**
     * 查询某观众在排队中的位置（按申请时间排序，1 起算；返回 0 表示不在队列）
     */
    public int getQueuePosition(Long roomId, Long viewerId) {
        List<LiveLinkmic> pending = linkmicMapper.selectList(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getStatus, 0)
                .orderByAsc(LiveLinkmic::getApplyTime));
        for (int i = 0; i < pending.size(); i++) {
            if (pending.get(i).getViewerId().equals(viewerId)) return i + 1;
        }
        return 0;
    }

    public void acceptLinkmic(Long linkmicId) {
        LiveLinkmic linkmic = linkmicMapper.selectById(linkmicId);
        if (linkmic != null) {
            // 检查连麦人数
            long currentCount = linkmicMapper.selectCount(new LambdaQueryWrapper<LiveLinkmic>()
                    .eq(LiveLinkmic::getRoomId, linkmic.getRoomId())
                    .eq(LiveLinkmic::getStatus, 1));
            if (currentCount >= linkmic.getMaxLinkmics()) {
                return;
            }
            linkmic.setStatus(1);
            linkmic.setConnectTime(LocalDateTime.now());
            linkmicMapper.updateById(linkmic);
        }
    }

    public void rejectLinkmic(Long linkmicId) {
        LiveLinkmic linkmic = linkmicMapper.selectById(linkmicId);
        if (linkmic != null) {
            linkmic.setStatus(3);
            linkmic.setEndTime(LocalDateTime.now());
            linkmicMapper.updateById(linkmic);
        }
    }

    public void disconnectLinkmic(Long linkmicId) {
        LiveLinkmic linkmic = linkmicMapper.selectById(linkmicId);
        if (linkmic != null) {
            linkmic.setStatus(2);
            linkmic.setEndTime(LocalDateTime.now());
            linkmicMapper.updateById(linkmic);
        }
    }

    public void toggleAudio(Long linkmicId, boolean enabled) {
        LiveLinkmic linkmic = linkmicMapper.selectById(linkmicId);
        if (linkmic != null) {
            linkmic.setAudioEnabled(enabled ? 1 : 0);
            linkmicMapper.updateById(linkmic);
        }
    }

    public void toggleVideo(Long linkmicId, boolean enabled) {
        LiveLinkmic linkmic = linkmicMapper.selectById(linkmicId);
        if (linkmic != null) {
            linkmic.setVideoEnabled(enabled ? 1 : 0);
            linkmicMapper.updateById(linkmic);
        }
    }

    public List<LiveLinkmic> getActiveLinkmics(Long roomId) {
        return linkmicMapper.selectList(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getStatus, 1));
    }

    public List<LiveLinkmic> getPendingApplications(Long roomId, Long streamerId) {
        return linkmicMapper.selectList(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getStreamerId, streamerId)
                .eq(LiveLinkmic::getStatus, 0)
                .orderByAsc(LiveLinkmic::getApplyTime));
    }

    public LiveLinkmic getByViewerId(Long roomId, Long viewerId) {
        return linkmicMapper.selectOne(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getViewerId, viewerId)
                .eq(LiveLinkmic::getStatus, 1));
    }
}