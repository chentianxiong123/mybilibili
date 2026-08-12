# Live 服务设计：语聊房取代会议

## 1. 背景：从会议到语聊房

原有代码中有一个"会议"模块（meeting_room + meeting_participant），定位是类腾讯会议的独立功能。经过分析，该模块对直播站没有产品场景，但它的数据模型（房间 + 参与者 + 角色 + 音视频开关）**恰好是语聊房的雏形**。

**决策：会议 → 语聊房**，live 服务核心为语聊房（多人语音房），1:1 连麦（linkmic）作为附属功能保留。

| 维度 | 会议（旧） | 语聊房（新） |
|------|-----------|-------------|
| 产品定位 | 办公工具 | 直播社交玩法 |
| 核心机制 | 预约/审核/主持人 | **麦位状态机**（上麦/禁麦/锁麦） |
| 用户场景 | 视频会议 | 房主+8嘉宾+海量听众语音互动 |
| 送礼链路 | 无 | 打通（给麦上用户送礼） |
| 行业验证 | 不适合直播站 | B站战略专项、YY/TT语音/Clubhouse |

---

## 2. 语聊房核心模型

### 2.1 角色定义

| 角色 | 权限 | 说明 |
|------|------|------|
| 房主（host） | 全部 | 创建房间、管理麦位、踢人、禁麦、锁麦 |
| 管理员（admin） | 部分 | 禁麦、踢人，由房主指定 |
| 麦上嘉宾（broadcaster） | 推流+拉流 | 在麦位上，可发言 |
| 听众（audience） | 只拉流 | 不在麦上，只收听 |

### 2.2 麦位状态机

```
听众 ──申请上麦──▶ 排队中 ──房主同意──▶ 麦上 ──主动下麦──▶ 听众
                        └─被拒绝──▶ 听众
麦上 ──被禁麦──▶ 麦上(静音) ──解禁──▶ 麦上
麦上 ──被踢────▶ 房间外
麦上 ──断线────▶ 房间外（超时自动释放麦位）
```

**麦位状态定义**：

| 状态 | 说明 | 触发条件 |
|------|------|---------|
| EMPTY | 空位 | 初始/下麦/踢人 |
| PENDING | 申请中 | 听众申请上麦 |
| OCCUPIED | 已占用 | 房主同意/房主直接安排 |
| MUTED | 被禁麦 | 房主禁言麦位（嘉宾仍在麦上但静音） |
| LOCKED | 锁定 | 房主锁定麦位，禁止申请 |

### 2.3 服务端权威

麦位状态**必须由服务端维护和变更**，客户端只读。所有麦位操作（申请/同意/拒绝/禁麦/锁麦/踢人）走 RPC 到服务端，服务端变更后广播给房间内所有人。

> 参考 B站经验：早期麦位控制放在客户端，多人语聊上线后状态不一致问题频发，最终迁移到服务端。详见《B站语聊房架构演进实践》。

---

## 3. 音频策略

| 角色 | 采集 | 推流 | 拉流 | 码率 |
|------|------|------|------|------|
| 麦上嘉宾 | 开 | 开（推 RTC 流） | 拉所有麦上流 | 48-64kbps |
| 听众 | 关 | 关 | 拉**服务端混音流** | 32-48kbps |

**关键：听众只拉服务端混音后的 1 路流**，而不是拉每个麦位各自的流。混音由 SRS（SFU 模式）或 RTC 服务器完成，业务服务端下发混音流地址。

**与 SRS 的配合**：

```
OBS(主播) ──RTMP──▶ SRS ──HLS/FLV──▶ 观众(CDN播放)
                         │
                    SFU 转发(WebRTC)
                         │
                    ┌────┴────┐
                    麦上嘉宾    麦上嘉宾
                    (WebRTC推流)  (WebRTC推流)
```

- 主播开播（OBS→RTMP→SRS→HLS/FLV）：传统直播链路
- 语聊房上麦者（WebRTC→SRS SFU→混音→HLS）：混入主播流，听众一路播放
- 房主/管理员通过 WS 信令控制麦位

---

## 4. 信令设计

### 4.1 信令协议（WS）

| 方向 | 消息类型 | 说明 |
|------|---------|------|
| 客户端→服务端 | `apply_seat` | 申请上麦（指定麦位序号） |
| 客户端→服务端 | `leave_seat` | 主动下麦 |
| 客户端→服务端 | `invite_accept` | 接受房主邀请上麦 |
| 服务端→客户端 | `seat_updated` | 麦位状态变更广播（全量/增量） |
| 客户端→服务端 | `mute_self` | 自己静音/取消静音 |
| 服务端→客户端 | `seat_muted` | 房主/管理员禁麦某麦位 |
| 服务端→客户端 | `seat_locked` | 房主锁定/解锁某麦位 |
| 服务端→客户端 | `kicked` | 被踢出房间 |

**麦位变更广播格式**：

```json
{
  "type": "seat_updated",
  "room_id": 1001,
  "seats": [
    { "index": 0, "status": "occupied", "user_id": 101, "user_name": "admin", "muted": false },
    { "index": 1, "status": "occupied", "user_id": 102, "user_name": "guest1", "muted": true },
    { "index": 2, "status": "pending", "user_id": 103, "user_name": "guest2" },
    { "index": 3, "status": "empty" },
    { "index": 4, "status": "locked" }
  ]
}
```

### 4.2 信令通道

语聊房信令走 **WebSocket**（真双向，保留在 live 服务的 ws-hub 模块）。不拆 SSE 理由：
- 上麦申请/审批/禁麦/踢人都是双向交互
- 麦位广播是高频小消息，WS 全双工无握手开销
- live 独立服务，不拖累 core 胖单体

---

## 5. 与现有代码映射

| 现有表/模块 | 去向 | 改造点 |
|------------|------|--------|
| `meeting_room` | 保留→语聊房房间 | 去掉预约/审核字段，加麦位数/房间模式 |
| `meeting_participant` | 保留→麦位状态 | 改为 `seat` 表，加状态机（seat_index/status/muted） |
| `live_linkmic` | 保留 | 1:1 连麦作为附属功能，不改 |
| MeetingWebSocketHandler | 改造→ws-hub 信令处理器 | 消息类型从会议切语聊房 |
| MeetingAdminController | **删除** | 预约/审核/后台管理，语聊房不需要 |

---

## 6. 数据模型（PG）

```sql
-- 语聊房房间表（从 meeting_room 改造）
CREATE TABLE live_rooms (
    id BIGSERIAL PRIMARY KEY,
    room_name VARCHAR(100) NOT NULL,
    room_code VARCHAR(10) UNIQUE NOT NULL,      -- 邀请码
    host_id BIGINT NOT NULL,                     -- 房主
    max_seats INT DEFAULT 8,                     -- 麦位数（默认 8 麦）
    status INT DEFAULT 0,                        -- 0=未开始 1=进行中 2=已结束
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 麦位表（从 meeting_participant 改造）
CREATE TABLE live_seats (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES live_rooms(id),
    seat_index INT NOT NULL,                     -- 麦位序号 (0~max_seats-1)
    user_id BIGINT,                              -- 占用用户（空麦位时为 NULL）
    status INT DEFAULT 0,                        -- 0=EMPTY 1=PENDING 2=OCCUPIED 3=MUTED 4=LOCKED
    muted BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMPTZ,
    UNIQUE(room_id, seat_index)
);

-- 1:1 连麦表（保留，linkmic）
CREATE TABLE live_linkmic (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL,
    streamer_id BIGINT NOT NULL,
    viewer_id BIGINT NOT NULL,
    status INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## 7. 服务模块划分

```
live/ 服务
├── room/      房间 CRUD + 邀请码 + SRS 回调（已有，改造）
├── seat/      麦位状态机 ★核心★（新增）
│   ├── 麦位分配（原子操作，防并发）
│   ├── 状态变更广播（WS 全量/增量推送）
│   └── 断线自动释放（超时检测）
├── linkmic/   1:1 连麦（保留，不改）
├── ws-hub/    信令处理器（从 meeting 改造）
│   ├── apply_seat / leave_seat / mute_self
│   ├── seat_muted / seat_locked / kicked
│   └── 房间内广播
└── chat/      文字聊天 + 礼物（新增，二期）
```

---

## 8. 总结

1. **会议 → 语聊房**：模型几乎一样，改造代价低，产品价值高
2. **麦位状态机是灵魂**：服务端权威，原子操作防并发，广播保一致性
3. **音频分级**：麦上推流，听众拉混音，SRS 做 SFU 转发
4. **信令走 WS**：保留在 live 独立服务，不拖累 core
5. **1:1 连麦保留**：语聊房和连麦是两种玩法，并行