# 视频转码模块完善计划

## 一、现状分析

### 1.1 后端现状
- **消息队列**：使用RocketMQ，Topic为`video-process-topic`
- **并发控制**：使用Redis分布式锁实现并发数为1（`video:process:lock`）
- **消费者配置**：`ConsumeMode.CONCURRENTLY`（并发消费）
- **状态码系统**：已定义完整（0-待处理，1-转码中，10-失败，11-成功等）
- **处理流程**：转码 → 音频提取 → 字幕生成 → AI摘要

### 1.2 前端现状
- **页面**：`VideoProcessView.vue` 已实现基本管理界面
- **功能**：统计卡片、筛选、流水线进度展示、手动触发操作
- **缺失**：无实时状态更新、无当前任务展示、无队列信息

### 1.3 存在的问题
1. **Redis锁超时风险**：锁设置30分钟超时，长任务可能并发
2. **无队列状态查询**：前端无法知道当前是否有任务在执行
3. **无实时进度反馈**：用户不知道任务执行到哪个阶段
4. **消费者模式不当**：应使用顺序消费而非并发消费+锁

---

## 二、设计方案

### 2.1 并发数为1的正确实现方式

**方案A：RocketMQ顺序消费（推荐）**
```java
@RocketMQMessageListener(
    topic = MQConstants.TOPIC_VIDEO_PROCESS,
    consumerGroup = MQConstants.GROUP_VIDEO_PROCESS,
    consumeMode = ConsumeMode.ORDERLY  // 顺序消费，单线程处理
)
```
- 优点：RocketMQ原生支持，无需额外锁机制
- 缺点：需要确保消息发送时使用相同的MessageQueueSelector

**方案B：Redis分布式锁（当前方案优化）**
- 增加锁续期机制（看门狗）
- 记录当前处理任务信息到Redis

**最终选择**：方案A + 方案B结合
- 使用顺序消费确保消息按顺序处理
- 使用Redis记录当前任务状态供前端查询

### 2.2 任务状态存储设计

在Redis中存储当前处理任务信息：
```
Key: video:process:current
Value: {
  "videoId": 123,
  "manuscriptId": 45,
  "processType": "ALL",
  "stage": "TRANSCODING",
  "progress": 30,
  "startTime": "2026-04-03T10:00:00",
  "videoTitle": "测试视频标题"
}
```

### 2.3 前端状态展示设计

```
┌─────────────────────────────────────────────────────────────┐
│  视频处理管理                                    [自动刷新: 开] │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 🔄 当前任务                                          │    │
│  │ 视频ID: 123 | 标题: 测试视频标题                      │    │
│  │ 阶段: 转码中 ████████░░░░░░░░ 45%                    │    │
│  │ 开始时间: 10:00:00 | 已用时: 5分钟                    │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ 📋 等待队列: 5 个任务                                 │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  [统计卡片区域]                                              │
│  ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐          │
│  │待处理│ │转码中│ │音频中│ │字幕中│ │AI中 │ │已完成│          │
│  │ 10  │ │ 1  │ │ 0  │ │ 0  │ │ 0  │ │ 20  │          │
│  └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘          │
│                                                             │
│  [视频列表表格...]                                           │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、实现步骤

### 3.1 后端改造

#### 步骤1：修改消费者为顺序消费
**文件**：`mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessConsumer.java`

修改内容：
1. 将 `ConsumeMode.CONCURRENTLY` 改为 `ConsumeMode.ORDERLY`
2. 移除Redis分布式锁逻辑（顺序消费已保证单线程）
3. 添加当前任务状态写入Redis的逻辑

#### 步骤2：创建任务状态API
**新建文件**：`mybilibili-ai/src/main/java/com/mybilibili/ai/controller/VideoProcessAdminController.java`

提供接口：
- `GET /ai/admin/process/current` - 获取当前处理任务
- `GET /ai/admin/process/queue` - 获取等待队列长度
- `GET /ai/admin/process/statistics` - 获取各状态统计

#### 步骤3：添加网关路由
**文件**：`mybilibili-gateway/src/main/resources/application.yml`

添加AI服务的admin路由：
```yaml
- id: ai-admin
  uri: lb://mybilibili-ai
  predicates:
    - Path=/api/ai/admin/**
  filters:
    - StripPrefix=1
```

#### 步骤4：优化任务进度更新
在各个处理阶段更新Redis中的进度信息：
- 转码开始：progress = 0-30%
- 音频提取：progress = 30-50%
- 字幕生成：progress = 50-80%
- AI摘要：progress = 80-100%

### 3.2 前端改造

#### 步骤1：添加API接口
**文件**：`mybilibili-admin-web/src/api/videoProcess.js`（新建）

```javascript
// 获取当前处理任务
export const getCurrentTask = () => request.get('/ai/admin/process/current')

// 获取队列信息
export const getQueueInfo = () => request.get('/ai/admin/process/queue')

// 获取统计数据
export const getStatistics = () => request.get('/ai/admin/process/statistics')
```

#### 步骤2：改造VideoProcessView.vue
**文件**：`mybilibili-admin-web/src/views/VideoProcessView.vue`

修改内容：
1. 添加"当前任务"卡片组件
2. 添加"等待队列"信息展示
3. 添加自动刷新功能（处理中时每5秒刷新）
4. 优化统计卡片数据来源（从后端获取）
5. 添加进度条动画效果

#### 步骤3：添加实时状态更新
使用定时器实现自动刷新：
```javascript
const autoRefresh = ref(true)
const refreshInterval = ref(null)

const startAutoRefresh = () => {
  if (refreshInterval.value) return
  refreshInterval.value = setInterval(() => {
    loadCurrentTask()
    loadStatistics()
  }, 5000)
}

const stopAutoRefresh = () => {
  if (refreshInterval.value) {
    clearInterval(refreshInterval.value)
    refreshInterval.value = null
  }
}
```

---

## 四、详细任务清单

### 后端任务

| 序号 | 任务 | 文件 | 优先级 |
|------|------|------|--------|
| 1 | 修改消费者为顺序消费模式 | VideoProcessConsumer.java | 高 |
| 2 | 添加当前任务状态写入Redis | VideoProcessConsumer.java | 高 |
| 3 | 创建VideoProcessAdminController | 新建 | 高 |
| 4 | 实现获取当前任务接口 | VideoProcessAdminController.java | 高 |
| 5 | 实现获取队列长度接口 | VideoProcessAdminController.java | 中 |
| 6 | 实现获取统计数据接口 | VideoProcessAdminController.java | 中 |
| 7 | 添加网关路由配置 | application.yml | 高 |
| 8 | 各阶段添加进度更新逻辑 | VideoProcessConsumer.java | 中 |

### 前端任务

| 序号 | 任务 | 文件 | 优先级 |
|------|------|------|--------|
| 1 | 创建videoProcess.js API文件 | 新建 | 高 |
| 2 | 添加当前任务卡片组件 | VideoProcessView.vue | 高 |
| 3 | 添加等待队列展示 | VideoProcessView.vue | 中 |
| 4 | 实现自动刷新功能 | VideoProcessView.vue | 高 |
| 5 | 优化统计卡片数据来源 | VideoProcessView.vue | 中 |
| 6 | 添加进度条动画 | VideoProcessView.vue | 低 |
| 7 | 添加任务完成通知 | VideoProcessView.vue | 低 |

---

## 五、接口设计

### 5.1 获取当前处理任务
```
GET /api/ai/admin/process/current

Response:
{
  "code": 200,
  "data": {
    "processing": true,
    "videoId": 123,
    "manuscriptId": 45,
    "videoTitle": "测试视频",
    "processType": "ALL",
    "stage": "TRANSCODING",
    "stageText": "视频转码中",
    "progress": 30,
    "startTime": "2026-04-03T10:00:00",
    "elapsedSeconds": 300
  }
}
```

### 5.2 获取队列信息
```
GET /api/ai/admin/process/queue

Response:
{
  "code": 200,
  "data": {
    "queueSize": 5,
    "messages": [
      { "videoId": 124, "manuscriptId": 46, "processType": "TRANSCODE" },
      { "videoId": 125, "manuscriptId": 47, "processType": "ALL" }
    ]
  }
}
```

### 5.3 获取统计数据
```
GET /api/ai/admin/process/statistics

Response:
{
  "code": 200,
  "data": {
    "pending": 10,
    "transcoding": 1,
    "audioExtracting": 0,
    "subtitleGenerating": 0,
    "aiSummarizing": 0,
    "completed": 20,
    "failed": 2
  }
}
```

---

## 六、Redis Key设计

| Key | 类型 | 说明 | TTL |
|-----|------|------|-----|
| `video:process:current` | String(JSON) | 当前处理任务信息 | 1小时 |
| `video:process:lock` | String | 处理锁（备用） | 30分钟 |
| `video:process:subtitle:{videoId}` | String | 视频字幕内容 | 永久 |
| `video:process:summary:{videoId}` | String | 视频AI摘要 | 永久 |

---

## 七、注意事项

1. **消息顺序性**：发送消息时需使用相同的MessageQueueSelector确保同一视频的消息按顺序处理
2. **错误处理**：处理失败时需正确更新状态，避免任务卡住
3. **前端轮询**：自动刷新间隔建议5秒，避免过于频繁请求
4. **状态一致性**：Redis状态与数据库状态需保持同步
