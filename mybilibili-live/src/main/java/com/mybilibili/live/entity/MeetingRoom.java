package com.mybilibili.live.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.time.LocalDateTime;

@Data
@TableName("meeting_room")
public class MeetingRoom {
    @TableId(type = IdType.AUTO)
    private Long id;
    private String roomName;
    private String roomCode;  // 6位邀请码
    private Long creatorId;
    private String creatorName;
    private Integer maxParticipants;  // 最大参与人数，默认5
    private Integer status;  // 0=未开始 1=进行中 2=已结束
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private LocalDateTime createTime;
    private LocalDateTime updateTime;
}