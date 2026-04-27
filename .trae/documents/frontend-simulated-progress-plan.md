# 前端模拟进度条方案

## 目标
- 前端使用模拟数据展示进度条，更稳定流畅
- 后端只在任务完成时发送SSE消息
- 不同阶段有不同的进度增长速度

## 进度增长规则

### 转码 (transcode)
- 进度范围: 0% → 100%
- 增长速度: 每秒 +1%
- 阶段划分:
  - 0-30%: 1080p转码中
  - 30-60%: 720p转码中
  - 60-90%: 480p转码中
  - 90-100%: 完成中（变慢）

### 音频提取 (audio)
- 进度范围: 0% → 100%
- 增长速度: 每秒 +2%

### 字幕生成 (subtitle)
- 进度范围: 0% → 100%
- 增长速度: 每秒 +1%
- 90%后变慢

### AI摘要 (ai)
- 进度范围: 0% → 100%
- 增长速度: 每秒 +2%
- 90%后变慢（每2秒+1%）

## 实现步骤

### 1. 前端修改 (VideoProcessView.vue)

#### 1.1 添加模拟进度定时器
```javascript
const progressTimer = ref(null)
const simulatedProgress = ref(0)
```

#### 1.2 创建模拟进度函数
```javascript
const startSimulatedProgress = (stage, videoId) => {
  stopSimulatedProgress()
  simulatedProgress.value = 0
  
  const speeds = {
    transcode: 1,
    audio: 2,
    subtitle: 1,
    ai: 2
  }
  
  const interval = 1000 // 1秒
  let speed = speeds[stage] || 1
  
  progressTimer.value = setInterval(() => {
    if (simulatedProgress.value >= 90) {
      // 90%后变慢
      speed = 0.5
    }
    
    simulatedProgress.value = Math.min(99, simulatedProgress.value + speed)
    
    // 更新currentTask
    currentTask.value.progress = Math.floor(simulatedProgress.value)
  }, interval)
}

const stopSimulatedProgress = () => {
  if (progressTimer.value) {
    clearInterval(progressTimer.value)
    progressTimer.value = null
  }
}
```

#### 1.3 修改handleSseMessage
- 收到完成消息(progress=100)时，停止模拟进度，直接设为100%
- 更新表格状态为完成（变绿）

#### 1.4 修改按钮点击处理
- 点击按钮时立即开始模拟进度
- 设置初始状态

### 2. 后端修改

#### 2.1 简化SSE推送
- 只在任务完成时发送 progress=100
- 移除中间进度推送

#### 2.2 修改文件
- `VideoTranscodeServiceImpl.java`: 只在完成时推送100%
- `VideoProcessConsumer.java`: 音频、字幕、AI摘要只在完成时推送100%
- `AiSubtitleServiceImpl.java`: 移除Whisper进度推送

### 3. 状态更新逻辑

#### 前端状态转换
```
点击按钮 → 立即变蓝 → 开始模拟进度 → 收到完成SSE → 变绿
```

#### SSE消息格式
```json
{
  "videoId": 28,
  "progress": 100,
  "status": "转码完成",
  "stage": "transcode"
}
```

## 文件修改清单

### 前端
- `mybilibili-admin-web/src/views/VideoProcessView.vue`

### 后端
- `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/VideoTranscodeServiceImpl.java`
- `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/AiSubtitleServiceImpl.java`
- `mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessConsumer.java`

## 优点
1. 前端进度更稳定，不受网络波动影响
2. 后端压力减小，不需要频繁推送
3. 用户体验更好，进度条流畅
4. 90%后变慢，给用户"即将完成"的心理预期
