package com.mybilibili.live.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.entity.LiveLinkmic;
import com.mybilibili.live.entity.LiveRoom;
import com.mybilibili.live.mapper.LiveLinkmicMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.time.LocalDateTime;
import java.util.List;
import java.util.function.Consumer;

@Slf4j
@Service
public class LiveLinkmicService {

    @Autowired
    private LiveLinkmicMapper linkmicMapper;

    @Autowired
    private LiveRoomService liveRoomService;

    private static final int MAX_LINKMICS = 3;

    @Transactional
    public LiveLinkmic applyLinkmic(Long roomId, Long viewerId, String viewerName) {
        LiveRoom liveRoom = requireLiveRoomAvailableForLinkmic(roomId);
        requireViewerIsNotStreamer(liveRoom, viewerId);

        LiveLinkmic existing = findPendingOrConnectedLinkmic(roomId, viewerId);
        if (existing != null) {
            return existing;
        }

        LiveLinkmic linkmic = new LiveLinkmic();
        linkmic.setRoomId(roomId);
        linkmic.setStreamerId(Long.valueOf(liveRoom.getUserId()));
        linkmic.setViewerId(viewerId);
        linkmic.setViewerName(StringUtils.hasText(viewerName) ? viewerName.trim() : "用户" + viewerId);
        linkmic.setStatus(LiveConstants.LinkmicStatus.PENDING);
        linkmic.setAudioEnabled(LiveConstants.SwitchState.OFF);
        linkmic.setVideoEnabled(LiveConstants.SwitchState.OFF);
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
                .eq(LiveLinkmic::getStatus, LiveConstants.LinkmicStatus.PENDING)
                .orderByAsc(LiveLinkmic::getApplyTime));
        for (int i = 0; i < pending.size(); i++) {
            if (pending.get(i).getViewerId().equals(viewerId)) return i + 1;
        }
        return 0;
    }

    @Transactional
    public void acceptLinkmic(Long linkmicId, Long operatorId) {
        LiveLinkmic linkmic = requireLinkmic(linkmicId);
        requireStreamer(linkmic, operatorId);
        requirePending(linkmic, "该连麦申请状态不可同意");
        ensureLinkmicCapacity(linkmic);
        markConnected(linkmic);
    }

    @Transactional
    public void rejectLinkmic(Long linkmicId, Long operatorId) {
        LiveLinkmic linkmic = requireLinkmic(linkmicId);
        requireStreamer(linkmic, operatorId);
        requirePending(linkmic, "该连麦申请状态不可拒绝");
        markFinished(linkmic, LiveConstants.LinkmicStatus.REJECTED);
    }

    @Transactional
    public void disconnectLinkmic(Long linkmicId, Long operatorId) {
        LiveLinkmic linkmic = requireLinkmic(linkmicId);
        requireStreamerOrViewer(linkmic, operatorId);
        if (LiveConstants.LinkmicStatus.isFinished(linkmic.getStatus())) {
            return;
        }
        markFinished(linkmic, LiveConstants.LinkmicStatus.DISCONNECTED);
    }

    public void toggleAudio(Long linkmicId, Long operatorId, boolean enabled) {
        updateMediaState(linkmicId, operatorId, linkmic ->
                linkmic.setAudioEnabled(toSwitchState(enabled)));
    }

    public void toggleVideo(Long linkmicId, Long operatorId, boolean enabled) {
        updateMediaState(linkmicId, operatorId, linkmic ->
                linkmic.setVideoEnabled(toSwitchState(enabled)));
    }

    private void updateMediaState(Long linkmicId, Long operatorId, Consumer<LiveLinkmic> updater) {
        LiveLinkmic linkmic = requireLinkmic(linkmicId);
        requireStreamerOrViewer(linkmic, operatorId);
        requireConnected(linkmic);
        updater.accept(linkmic);
        linkmicMapper.updateById(linkmic);
    }

    public List<LiveLinkmic> getActiveLinkmics(Long roomId) {
        return linkmicMapper.selectList(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getStatus, LiveConstants.LinkmicStatus.CONNECTED));
    }

    public List<LiveLinkmic> getPendingApplications(Long roomId, Long streamerId) {
        return linkmicMapper.selectList(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getStreamerId, streamerId)
                .eq(LiveLinkmic::getStatus, LiveConstants.LinkmicStatus.PENDING)
                .orderByAsc(LiveLinkmic::getApplyTime));
    }

    public LiveLinkmic getByViewerId(Long roomId, Long viewerId) {
        return linkmicMapper.selectOne(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getViewerId, viewerId)
                .eq(LiveLinkmic::getStatus, LiveConstants.LinkmicStatus.CONNECTED));
    }

    private LiveRoom requireLiveRoomAvailableForLinkmic(Long roomId) {
        LiveRoom liveRoom = liveRoomService.getById(roomId.intValue());
        if (liveRoom == null) {
            throw new BusinessException("直播间不存在");
        }
        if (!LiveConstants.RoomStatus.LIVE.equals(liveRoom.getStatus())) {
            throw new BusinessException("直播未开播，暂不能申请连麦");
        }
        return liveRoom;
    }

    private void requireViewerIsNotStreamer(LiveRoom liveRoom, Long viewerId) {
        if (Long.valueOf(liveRoom.getUserId()).equals(viewerId)) {
            throw new BusinessException("主播无需申请连麦");
        }
    }

    private LiveLinkmic findPendingOrConnectedLinkmic(Long roomId, Long viewerId) {
        return linkmicMapper.selectOne(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, roomId)
                .eq(LiveLinkmic::getViewerId, viewerId)
                .in(LiveLinkmic::getStatus,
                        LiveConstants.LinkmicStatus.PENDING,
                        LiveConstants.LinkmicStatus.CONNECTED));
    }

    private void requirePending(LiveLinkmic linkmic, String message) {
        if (!LiveConstants.LinkmicStatus.isPending(linkmic.getStatus())) {
            throw new BusinessException(message);
        }
    }

    private void requireConnected(LiveLinkmic linkmic) {
        if (!LiveConstants.LinkmicStatus.isConnected(linkmic.getStatus())) {
            throw new BusinessException("连麦未连接，不能切换媒体状态");
        }
    }

    private void ensureLinkmicCapacity(LiveLinkmic linkmic) {
        long currentCount = linkmicMapper.selectCount(new LambdaQueryWrapper<LiveLinkmic>()
                .eq(LiveLinkmic::getRoomId, linkmic.getRoomId())
                .eq(LiveLinkmic::getStatus, LiveConstants.LinkmicStatus.CONNECTED));
        int maxLinkmics = linkmic.getMaxLinkmics() == null ? MAX_LINKMICS : linkmic.getMaxLinkmics();
        if (currentCount >= maxLinkmics) {
            throw new BusinessException("当前连麦人数已满");
        }
    }

    private void markConnected(LiveLinkmic linkmic) {
        linkmic.setStatus(LiveConstants.LinkmicStatus.CONNECTED);
        linkmic.setConnectTime(LocalDateTime.now());
        linkmicMapper.updateById(linkmic);
    }

    private void markFinished(LiveLinkmic linkmic, int status) {
        linkmic.setStatus(status);
        linkmic.setEndTime(LocalDateTime.now());
        linkmicMapper.updateById(linkmic);
    }

    private int toSwitchState(boolean enabled) {
        return enabled ? LiveConstants.SwitchState.ON : LiveConstants.SwitchState.OFF;
    }

    private LiveLinkmic requireLinkmic(Long linkmicId) {
        LiveLinkmic linkmic = linkmicMapper.selectById(linkmicId);
        if (linkmic == null) {
            throw new BusinessException("连麦记录不存在");
        }
        return linkmic;
    }

    private void requireStreamer(LiveLinkmic linkmic, Long operatorId) {
        if (!linkmic.getStreamerId().equals(operatorId)) {
            throw new BusinessException(403, "无权操作该连麦");
        }
    }

    private void requireStreamerOrViewer(LiveLinkmic linkmic, Long operatorId) {
        if (!linkmic.getStreamerId().equals(operatorId) && !linkmic.getViewerId().equals(operatorId)) {
            throw new BusinessException(403, "无权操作该连麦");
        }
    }
}
