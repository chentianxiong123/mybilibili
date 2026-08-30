package com.mybilibili.live.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.time.LocalDateTime;

@Data
@TableName("meeting_participant")
public class MeetingParticipant {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long roomId;
    private Long userId;
    private String userName;
    private String userAvatar;
    private Integer role;  // 0=参与者 1=主持人
    private Integer audioEnabled;  // 0=关闭 1=开启
    private Integer videoEnabled;  // 0=关闭 1=开启
    private Integer screenShareEnabled;  // 0=关闭 1=开启
    private LocalDateTime joinTime;
    private LocalDateTime leaveTime;
}