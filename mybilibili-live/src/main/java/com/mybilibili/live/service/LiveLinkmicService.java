package com.mybilibili.live.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.live.entity.LiveLinkmic;
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

    private static final int MAX_LINKMICS = 3;

    public LiveLinkmic applyLinkmic(Long roomId, Long viewerId, String viewerName) {
        // 检查是否已经申请过
        LiveLinkmic existing = linkmicMapper.selectOne(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getViewerId, viewerId)
                .in(LiveLinkmic::getStatus, 0, 1));
        if (existing != null) {
            return existing;
        }

        // 检查当前连麦人数
        long currentCount = linkmicMapper.selectCount(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getStatus, 1));
        if (currentCount >= MAX_LINKMICS) {
            return null; // 人数已满
        }

        LiveLinkmic linkmic = new LiveLinkmic();
        linkmic.setRoomId(roomId);
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
                .eq(LiveLinkmic::getStatus, 0));
    }

    public LiveLinkmic getByViewerId(Long roomId, Long viewerId) {
        return linkmicMapper.selectOne(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getViewerId, viewerId)
                .eq(LiveLinkmic::getStatus, 1));
    }
}