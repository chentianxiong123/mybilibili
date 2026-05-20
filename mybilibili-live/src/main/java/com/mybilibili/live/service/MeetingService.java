package com.mybilibili.live.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.live.entity.MeetingRoom;
import com.mybilibili.live.entity.MeetingParticipant;
import com.mybilibili.live.mapper.MeetingRoomMapper;
import com.mybilibili.live.mapper.MeetingParticipantMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Random;

@Slf4j
@Service
public class MeetingService {

    @Autowired
    private MeetingRoomMapper roomMapper;

    @Autowired
    private MeetingParticipantMapper participantMapper;

    private final Random random = new Random();

    public MeetingRoom createRoom(String roomName, Long creatorId, String creatorName) {
        MeetingRoom room = new MeetingRoom();
        room.setRoomName(roomName);
        room.setRoomCode(generateRoomCode());
        room.setCreatorId(creatorId);
        room.setCreatorName(creatorName);
        room.setMaxParticipants(5);
        room.setStatus(0);
        room.setCreateTime(LocalDateTime.now());
        room.setUpdateTime(LocalDateTime.now());
        roomMapper.insert(room);
        return room;
    }

    public MeetingRoom getRoomByCode(String roomCode) {
        return roomMapper.selectOne(new LambdaQueryWrapper<MeetingRoom>()
                .eq(MeetingRoom::getRoomCode, roomCode));
    }

    public List<MeetingRoom> getMyRooms(Long userId) {
        return roomMapper.selectList(new LambdaQueryWrapper<MeetingRoom>()
                .eq(MeetingRoom::getCreatorId, userId)
                .orderByDesc(MeetingRoom::getCreateTime));
    }

    @Transactional
    public MeetingParticipant joinRoom(Long roomId, Long userId, String userName) {
        MeetingParticipant participant = new MeetingParticipant();
        participant.setRoomId(roomId);
        participant.setUserId(userId);
        participant.setUserName(userName);
        participant.setRole(0);
        participant.setAudioEnabled(0);
        participant.setVideoEnabled(0);
        participant.setScreenShareEnabled(0);
        participant.setJoinTime(LocalDateTime.now());
        participantMapper.insert(participant);

        // 更新房间状态为进行中
        MeetingRoom room = roomMapper.selectById(roomId);
        if (room != null && room.getStatus() == 0) {
            room.setStatus(1);
            room.setStartTime(LocalDateTime.now());
            room.setUpdateTime(LocalDateTime.now());
            roomMapper.updateById(room);
        }

        return participant;
    }

    public void leaveRoom(Long roomId, Long userId) {
        participantMapper.delete(new LambdaQueryWrapper<MeetingParticipant>()
                .eq(MeetingParticipant::getRoomId, roomId)
                .eq(MeetingParticipant::getUserId, userId));

        // 检查房间是否还有人
        Long count = participantMapper.selectCount(new LambdaQueryWrapper<MeetingParticipant>()
                .eq(MeetingParticipant::getRoomId, roomId));
        if (count == 0) {
            MeetingRoom room = roomMapper.selectById(roomId);
            if (room != null) {
                room.setStatus(2);
                room.setEndTime(LocalDateTime.now());
                room.setUpdateTime(LocalDateTime.now());
                roomMapper.updateById(room);
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
        room.setStatus(2);
        room.setEndTime(LocalDateTime.now());
        room.setUpdateTime(LocalDateTime.now());
        roomMapper.updateById(room);
        participantMapper.delete(new LambdaQueryWrapper<MeetingParticipant>()
                .eq(MeetingParticipant::getRoomId, roomId));
        return true;
    }

    private String generateRoomCode() {
        String code;
        do {
            code = String.format("%06d", random.nextInt(1000000));
        } while (roomMapper.selectOne(new LambdaQueryWrapper<MeetingRoom>()
                .eq(MeetingRoom::getRoomCode, code)) != null);
        return code;
    }
}