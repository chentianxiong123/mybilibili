package com.mybilibili.live.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import java.time.LocalDateTime;

@Data
@TableName("live_linkmic")
public class LiveLinkmic {
    @TableId(type = IdType.AUTO)
    private Long id;
    private Long roomId;           // 直播间ID
    private Long streamerId;       // 主播ID
    private Long viewerId;         // 观众ID（连麦者）
    private String viewerName;
    private Integer status;         // 0=申请中 1=已连接 2=已断开 3=被拒绝
    private Integer audioEnabled;  // 0=关闭 1=开启
    private Integer videoEnabled;  // 0=关闭 1=开启
    private Integer maxLinkmics;   // 最大连麦人数，默认3
    private LocalDateTime applyTime;
    private LocalDateTime connectTime;
    private LocalDateTime endTime;
}