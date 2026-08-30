package com.mybilibili.live.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.UpdateWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.live.common.LiveConstants;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.entity.MeetingParticipant;
import com.mybilibili.live.mapper.MeetingRoomMapper;
import com.mybilibili.live.mapper.MeetingParticipantMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.StringUtils;

import java.security.SecureRandom;
import java.time.Duration;
import java.time.LocalDateTime;
import java.util.List;
import java.util.function.Consumer;

@Slf4j
@Service
public class MeetingService {

    @Autowired
    private MeetingRoomMapper roomMapper;

    @Autowired
    private MeetingParticipantMapper participantMapper;

    private final SecureRandom random = new SecureRandom();

    @Transactional
    public MeetingRoom reserve(String roomName, Long creatorId, String creatorName,
                                LocalDateTime scheduledStart, LocalDateTime scheduledEnd, String reason) {
        validateReservationWindow(scheduledStart, scheduledEnd);
        MeetingRoom room = createRoom(roomName, creatorId, creatorName);
        room.setScheduledStart(scheduledStart);
        room.setScheduledEnd(scheduledEnd);
        room.setScheduledReason(StringUtils.hasText(reason) ? reason.trim() : null);
        room.setStatus(LiveConstants.MeetingStatus.PENDING_APPROVAL);
        room.setUpdateTime(LocalDateTime.now());
        roomMapper.updateById(room);
        return room;
    }

    public IPage<MeetingRoom> getAllRooms(int page, int size, Integer status) {
        if (status != null && !LiveConstants.MeetingStatus.isValid(status)) {
            throw new BusinessException("会议状态无效");
        }
        Page<MeetingRoom> pageObj = new Page<>(Math.max(page, 1), Math.min(Math.max(size, 1), 100));
        LambdaQueryWrapper<MeetingRoom> wrapper = new LambdaQueryWrapper<>();
        if (status != null) {
            wrapper.eq(MeetingRoom::getStatus, status);
        }
        wrapper.orderByDesc(MeetingRoom::getCreateTime);
        return roomMapper.selectPage(pageObj, wrapper);
    }

    public List<MeetingRoom> getPendingReservations() {
        return roomMapper.selectList(new LambdaQueryWrapper<MeetingRoom>()
                .isNotNull(MeetingRoom::getScheduledStart)
                .eq(MeetingRoom::getStatus, LiveConstants.MeetingStatus.PENDING_APPROVAL)
                .orderByAsc(MeetingRoom::getScheduledStart));
    }

    @Transactional
    public boolean approveReservation(Long roomId) {
        return updatePendingReservation(roomId, true, room ->
                room.setStatus(LiveConstants.MeetingStatus.NOT_STARTED));
    }

    @Transactional
    public boolean rejectReservation(Long roomId) {
        return updatePendingReservation(roomId, false, room -> {
            room.setStatus(LiveConstants.MeetingStatus.REJECTED);
            room.setEndTime(LocalDateTime.now());
        });
    }

    @Transactional
    public boolean leaveAll(Long roomId) {
        MeetingRoom room = roomMapper.selectById(roomId);
        if (room == null) {
            return false;
        }
        markParticipantsLeft(roomId);
        finishRoom(room);
        return true;
    }

    @Transactional
    public MeetingRoom createRoom(String roomName, Long creatorId, String creatorName) {
        MeetingRoom room = new MeetingRoom();
        room.setRoomName(normalizeRoomName(roomName));
        room.setRoomCode(generateRoomCode());
        room.setCreatorId(creatorId);
        room.setCreatorName(StringUtils.hasText(creatorName) ? creatorName.trim() : "用户" + creatorId);
        room.setMaxParticipants(5);
        room.setStatus(LiveConstants.MeetingStatus.NOT_STARTED);
        room.setCreateTime(LocalDateTime.now());
        room.setUpdateTime(LocalDateTime.now());
        roomMapper.insert(room);
        return room;
    }

    public MeetingRoom getRoomByCode(String roomCode) {
        if (!StringUtils.hasText(roomCode)) {
            return null;
        }
        return roomMapper.selectOne(new LambdaQueryWrapper<MeetingRoom>()
                .eq(MeetingRoom::getRoomCode, roomCode.trim()));
    }

    public List<MeetingRoom> getMyRooms(Long userId) {
        return roomMapper.selectList(new LambdaQueryWrapper<MeetingRoom>()
                .eq(MeetingRoom::getCreatorId, userId)
                .orderByDesc(MeetingRoom::getCreateTime));
    }

    @Transactional
    public MeetingParticipant joinRoom(Long roomId, Long userId, String userName) {
        MeetingRoom room = roomMapper.selectById(roomId);
        if (room == null) {
            throw new BusinessException("会议室不存在");
        }
        ensureJoinable(room);

        MeetingParticipant existing = getActiveParticipant(roomId, userId);
        if (existing != null) {
            return existing;
        }

        long activeCount = countActiveParticipants(roomId);
        int maxParticipants = room.getMaxParticipants() == null ? 5 : room.getMaxParticipants();
        if (activeCount >= maxParticipants) {
            throw new BusinessException("会议人数已满");
        }

        MeetingParticipant participant = new MeetingParticipant();
        participant.setRoomId(roomId);
        participant.setUserId(userId);
        participant.setUserName(StringUtils.hasText(userName) ? userName.trim() : "用户" + userId);
        participant.setRole(room.getCreatorId().equals(userId)
                ? LiveConstants.ParticipantRole.HOST
                : LiveConstants.ParticipantRole.PARTICIPANT);
        participant.setAudioEnabled(LiveConstants.SwitchState.OFF);
        participant.setVideoEnabled(LiveConstants.SwitchState.OFF);
        participant.setScreenShareEnabled(LiveConstants.SwitchState.OFF);
        participant.setJoinTime(LocalDateTime.now());
        participantMapper.insert(participant);

        if (LiveConstants.MeetingStatus.isNotStarted(room.getStatus())) {
            room.setStatus(LiveConstants.MeetingStatus.IN_PROGRESS);
            if (room.getStartTime() == null) {
                room.setStartTime(LocalDateTime.now());
            }
            room.setUpdateTime(LocalDateTime.now());
            roomMapper.updateById(room);
        }

        return participant;
    }

    @Transactional
    public void leaveRoom(Long roomId, Long userId) {
        MeetingParticipant participant = getActiveParticipant(roomId, userId);
        if (participant == null) {
            return;
        }
        participant.setLeaveTime(LocalDateTime.now());
        participantMapper.updateById(participant);

        Long count = countActiveParticipants(roomId);
        if (count == 0) {
            MeetingRoom room = roomMapper.selectById(roomId);
            if (room != null) {
                finishRoom(room);
            }
        }
    }

    public List<MeetingParticipant> getParticipants(Long roomId) {
        return participantMapper.selectList(new LambdaQueryWrapper<MeetingParticipant>()
                .eq(MeetingParticipant::getRoomId, roomId)
                .isNull(MeetingParticipant::getLeaveTime));
    }

    /**
     * 主持人结束会议：把状态置 2，清空所有参与者。
     * 仅创建者可调用。
     */
    @Transactional
    public boolean endRoom(Long roomId, Long operatorId) {
        MeetingRoom room = roomMapper.selectById(roomId);
        if (room == null) return false;
        if (!room.getCreatorId().equals(operatorId)) return false;
        finishRoom(room);
        markParticipantsLeft(roomId);
        return true;
    }

    private String generateRoomCode() {
        for (int i = 0; i < 20; i++) {
            String code = String.format("%06d", random.nextInt(1000000));
            if (roomMapper.selectOne(new LambdaQueryWrapper<MeetingRoom>()
                    .eq(MeetingRoom::getRoomCode, code)) == null) {
                return code;
            }
        }
        throw new BusinessException("会议邀请码生成失败，请重试");
    }

    private String normalizeRoomName(String roomName) {
        if (!StringUtils.hasText(roomName)) {
            throw new BusinessException("会议室名称不能为空");
        }
        return roomName.trim();
    }

    private void ensureJoinable(MeetingRoom room) {
        Integer status = room.getStatus();
        if (LiveConstants.MeetingStatus.isPendingApproval(status)) {
            throw new BusinessException("会议预约待审批");
        }
        if (LiveConstants.MeetingStatus.isRejected(status)) {
            throw new BusinessException("会议预约已被拒绝");
        }
        if (LiveConstants.MeetingStatus.isEnded(status)) {
            throw new BusinessException("会议已结束");
        }
        if (!LiveConstants.MeetingStatus.isJoinable(status)) {
            throw new BusinessException("会议状态无效");
        }
    }

    private MeetingParticipant getActiveParticipant(Long roomId, Long userId) {
        return participantMapper.selectOne(new LambdaQueryWrapper<MeetingParticipant>()
                .eq(MeetingParticipant::getRoomId, roomId)
                .eq(MeetingParticipant::getUserId, userId)
                .isNull(MeetingParticipant::getLeaveTime)
                .last("LIMIT 1"));
    }

    private Long countActiveParticipants(Long roomId) {
        return participantMapper.selectCount(new LambdaQueryWrapper<MeetingParticipant>()
                .eq(MeetingParticipant::getRoomId, roomId)
                .isNull(MeetingParticipant::getLeaveTime));
    }

    private void finishRoom(MeetingRoom room) {
        room.setStatus(LiveConstants.MeetingStatus.ENDED);
        if (room.getEndTime() == null) {
            room.setEndTime(LocalDateTime.now());
        }
        room.setUpdateTime(LocalDateTime.now());
        roomMapper.updateById(room);
    }

    private void markParticipantsLeft(Long roomId) {
        participantMapper.update(null, new UpdateWrapper<MeetingParticipant>()
                .eq("room_id", roomId)
                .isNull("leave_time")
                .set("leave_time", LocalDateTime.now()));
    }

    private void validateReservationWindow(LocalDateTime start, LocalDateTime end) {
        if (start == null || end == null) {
            throw new BusinessException("请选择预约时间段");
        }
        if (!end.isAfter(start)) {
            throw new BusinessException("结束时间必须晚于开始时间");
        }
        if (start.isBefore(LocalDateTime.now())) {
            throw new BusinessException("预约开始时间不能早于现在");
        }
        long minutes = Duration.between(start, end).toMinutes();
        if (minutes > 8 * 60) {
            throw new BusinessException("单次预约最长8小时");
        }
        if (minutes < 60) {
            throw new BusinessException("单次预约最少1小时");
        }
    }

    private boolean updatePendingReservation(Long roomId, boolean requireScheduledStart, Consumer<MeetingRoom> updater) {
        MeetingRoom room = roomMapper.selectById(roomId);
        if (room == null || !LiveConstants.MeetingStatus.isPendingApproval(room.getStatus())) {
            return false;
        }
        if (requireScheduledStart && room.getScheduledStart() == null) {
            return false;
        }
        updater.accept(room);
        room.setUpdateTime(LocalDateTime.now());
        roomMapper.updateById(room);
        return true;
    }
}
