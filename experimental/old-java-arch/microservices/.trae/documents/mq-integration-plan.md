# 消息队列(MQ)中间件集成方案

## 一、现状分析

### 1.1 已有基础设施

| 模块 | 内容 | 状态 |
|------|------|------|
| `mybilibili-mq` | RocketMQ依赖、生产者、消息定义 | ✅ 已存在 |
| `VideoMQProducer` | 消息发送服务 | ✅ 已实现 |
| `VideoProcessMessage` | 消息体定义（TRANSCODE/EXTRACT_AUDIO/GENERATE_SUBTITLE/AI_SUMMARY/ALL） | ✅ 已定义 |
| `AiSubtitleServiceImpl` | 音频提取 + Whisper字幕生成 | ✅ 已实现 |
| `AiSummaryServiceImpl` | DeepSeek AI摘要生成 | ✅ 已实现 |
| `AiTaskService` | 任务调度服务 | ✅ 已实现 |

### 1.2 缺失部分

| 模块 | 说明 |
|------|------|
| 消费者监听器 | 没有实现 `@RocketMQMessageListener` |
| 视频转码服务 | 没有独立的转码处理逻辑 |
| AI服务MQ依赖 | `mybilibili-ai` 未引入 `mybilibili-mq` |
| 串行处理机制 | 需要限制并发数为1 |

### 1.3 硬件限制

- **CPU限制**：只支持一次处理一种任务
- **解决方案**：消费者并发数设为1，实现串行处理

---

## 二、架构设计

### 2.1 处理流程

```
视频上传 → 发送消息 → MQ队列 → 消费者(串行) → 处理任务 → 更新状态
                                    ↓
                    ┌───────────────┼───────────────┐
                    ↓               ↓               ↓
              转码处理        音频提取+字幕      AI摘要
              (FFmpeg)       (Whisper)        (DeepSeek)
```

### 2.2 任务类型与优先级

| 任务类型 | Tag | 优先级 | 说明 |
|---------|-----|-------|------|
| TRANSCODE | 转码 | 高 | 视频转码，生成多清晰度 |
| EXTRACT_AUDIO | 音频提取 | 高 | 从视频提取音频 |
| GENERATE_SUBTITLE | 字幕生成 | 中 | Whisper生成字幕 |
| AI_SUMMARY | AI摘要 | 低 | DeepSeek生成摘要 |
| ALL | 全流程 | 高 | 按顺序执行所有步骤 |

### 2.3 串行处理策略

由于CPU限制，采用以下策略：

1. **单消费者实例**：`consumeMode = ConsumeMode.CONCURRENTLY`, `consumeThreadNumber = 1`
2. **任务队列**：使用RocketMQ的FIFO特性保证顺序
3. **处理锁**：Redis分布式锁防止多实例并发

---

## 三、实施步骤

### 阶段一：依赖配置（优先级：高）

#### 3.1.1 为 mybilibili-ai 添加 MQ 依赖

修改 `mybilibili-ai/pom.xml`：
```xml
<dependency>
    <groupId>com.mybilibili</groupId>
    <artifactId>mybilibili-mq</artifactId>
</dependency>
```

#### 3.1.2 添加 RocketMQ 配置

在 `mybilibili-ai/src/main/resources/application.yml` 添加：
```yaml
rocketmq:
  name-server: 127.0.0.1:9876
  consumer:
    group: video-process-group
```

### 阶段二：创建消费者（优先级：高）

#### 3.2.1 创建视频处理消费者

文件：`mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessConsumer.java`

```java
@Component
@RocketMQMessageListener(
    topic = "video-process-topic",
    consumerGroup = "video-process-group",
    consumeMode = ConsumeMode.CONCURRENTLY,
    consumeThreadNumber = 1,  // 关键：串行处理
    maxReconsumeTimes = 3
)
public class VideoProcessConsumer implements RocketMQListener<VideoProcessMessage> {
    // 实现消息处理逻辑
}
```

#### 3.2.2 处理逻辑设计

```java
@Override
public void onMessage(VideoProcessMessage message) {
    String processType = message.getProcessType();
    
    // 获取分布式锁（防止多实例并发）
    String lockKey = "video:process:lock";
    Boolean locked = redisTemplate.opsForValue().setIfAbsent(lockKey, "1", 30, TimeUnit.MINUTES);
    
    if (!locked) {
        // 当前有任务在处理，重新入队延迟处理
        throw new RuntimeException("系统繁忙，任务将重试");
    }
    
    try {
        switch (processType) {
            case "TRANSCODE":
                processTranscode(message);
                break;
            case "EXTRACT_AUDIO":
                processExtractAudio(message);
                break;
            case "GENERATE_SUBTITLE":
                processGenerateSubtitle(message);
                break;
            case "AI_SUMMARY":
                processAiSummary(message);
                break;
            case "ALL":
                processAll(message);
                break;
        }
    } finally {
        redisTemplate.delete(lockKey);
    }
}
```

### 阶段三：实现处理服务（优先级：高）

#### 3.3.1 创建视频转码服务

文件：`mybilibili-ai/src/main/java/com/mybilibili/ai/service/VideoTranscodeService.java`

```java
public interface VideoTranscodeService {
    boolean transcode(Integer manuscriptId, Integer videoId);
}
```

实现逻辑：
- 使用FFmpeg转码生成多清晰度（HD/SD/LD）
- 更新视频状态
- 记录处理进度

#### 3.3.2 创建音频提取服务

文件：`mybilibili-ai/src/main/java/com/mybilibili/ai/service/AudioExtractService.java`

```java
public interface AudioExtractService {
    boolean extractAudio(Integer manuscriptId, Integer videoId);
}
```

实现逻辑：
- 从视频提取16kHz单声道PCM音频
- 保存到指定目录
- 更新视频状态

### 阶段四：全流程处理（优先级：中）

#### 3.4.1 实现串行全流程

```java
private void processAll(VideoProcessMessage message) {
    Integer manuscriptId = message.getManuscriptId();
    Integer videoId = message.getVideoId();
    
    // 步骤1：转码
    updateStatus(videoId, Video.PROCESS_STATUS_TRANSCODING);
    if (!videoTranscodeService.transcode(manuscriptId, videoId)) {
        updateStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_FAILED);
        return;
    }
    updateStatus(videoId, Video.PROCESS_STATUS_TRANSCODE_SUCCESS);
    
    // 步骤2：音频提取
    updateStatus(videoId, Video.PROCESS_STATUS_AUDIO_EXTRACTING);
    if (!audioExtractService.extractAudio(manuscriptId, videoId)) {
        updateStatus(videoId, Video.PROCESS_STATUS_AUDIO_FAILED);
        return;
    }
    updateStatus(videoId, Video.PROCESS_STATUS_AUDIO_SUCCESS);
    
    // 步骤3：字幕生成
    updateStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_GENERATING);
    if (!aiSubtitleService.generateSubtitle(manuscriptId, videoId)) {
        updateStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_FAILED);
        return;
    }
    updateStatus(videoId, Video.PROCESS_STATUS_SUBTITLE_SUCCESS);
    
    // 步骤4：AI摘要
    updateStatus(videoId, Video.PROCESS_STATUS_AI_SUMMARIZING);
    if (!aiSummaryService.generateSummary(...)) {
        updateStatus(videoId, Video.PROCESS_STATUS_AI_FAILED);
        return;
    }
    
    // 完成
    updateStatus(videoId, Video.PROCESS_STATUS_COMPLETED);
}
```

### 阶段五：集成到上传流程（优先级：高）

#### 3.5.1 修改稿件上传服务

在 `ManuscriptServiceImpl.uploadManuscript()` 中添加：

```java
// 上传完成后，发送处理消息
VideoProcessMessage message = VideoProcessMessage.of(
    manuscript.getId(),
    video.getId(),
    userId,
    VideoProcessMessage.PROCESS_TYPE_ALL
);
videoMQProducer.sendVideoProcessMessage(message);
```

### 阶段六：错误处理与重试（优先级：中）

#### 3.6.1 消息重试机制

- 利用RocketMQ的重试特性：`maxReconsumeTimes = 3`
- 失败后更新视频状态为对应的失败状态
- 记录错误日志到数据库

#### 3.6.2 死信队列处理

```java
@Component
@RocketMQMessageListener(
    topic = "%DLQ%video-process-group",
    consumerGroup = "video-process-dlq-group"
)
public class VideoProcessDLQConsumer implements RocketMQListener<VideoProcessMessage> {
    // 处理最终失败的消息
}
```

---

## 四、文件清单

### 4.1 新建文件

| 文件路径 | 说明 |
|---------|------|
| `mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessConsumer.java` | 视频处理消费者 |
| `mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessDLQConsumer.java` | 死信队列消费者 |
| `mybilibili-ai/src/main/java/com/mybilibili/ai/service/VideoTranscodeService.java` | 转码服务接口 |
| `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/VideoTranscodeServiceImpl.java` | 转码服务实现 |
| `mybilibili-ai/src/main/java/com/mybilibili/ai/service/AudioExtractService.java` | 音频提取接口 |
| `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/AudioExtractServiceImpl.java` | 音频提取实现 |

### 4.2 修改文件

| 文件路径 | 修改内容 |
|---------|---------|
| `mybilibili-ai/pom.xml` | 添加 mybilibili-mq 依赖 |
| `mybilibili-ai/src/main/resources/application.yml` | 添加 RocketMQ 配置 |
| `mybilibili-video/src/main/java/.../ManuscriptServiceImpl.java` | 上传后发送MQ消息 |
| `mybilibili-video/pom.xml` | 添加 mybilibili-mq 依赖 |

---

## 五、配置说明

### 5.1 RocketMQ 服务端配置

需要本地安装 RocketMQ：
- NameServer: 127.0.0.1:9876
- Broker: 默认配置

### 5.2 消费者配置

```yaml
rocketmq:
  name-server: 127.0.0.1:9876
  consumer:
    group: video-process-group
    consume-thread-number: 1  # 串行处理
    max-reconsume-times: 3    # 最大重试次数
```

### 5.3 Redis锁配置

```yaml
spring:
  redis:
    host: 127.0.0.1
    port: 6379
```

---

## 六、测试计划

1. **单元测试**：各处理服务的独立测试
2. **集成测试**：消息发送→消费→处理的完整流程
3. **压力测试**：多个视频同时上传的队列处理
4. **异常测试**：处理失败的重试机制

---

## 七、注意事项

1. **串行保证**：`consumeThreadNumber = 1` 确保单线程消费
2. **分布式锁**：多实例部署时需要Redis锁
3. **资源释放**：处理完成后及时释放锁
4. **状态同步**：实时更新视频处理状态到数据库
5. **日志记录**：详细记录处理过程便于排查问题
