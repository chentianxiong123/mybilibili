# 任务队列与前端协调方案（WebSocket + MySQL）

## 技术方案

### 核心设计

1. **WebSocket 实时推送**：后端主动推送处理进度到前端
2. **MySQL 状态存储**：直接使用 video 表的 processStatus 字段，不需要 Redis
3. **前端实时展示**：WebSocket 连接，实时更新 UI

---

## 实施步骤

### 步骤 1：后端 WebSocket 配置

**新增文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\config\WebSocketConfig.java`

```java
package com.mybilibili.ai.config;

import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.socket.config.annotation.EnableWebSocket;
import org.springframework.web.socket.config.annotation.WebSocketConfigurer;
import org.springframework.web.socket.config.annotation.WebSocketHandlerRegistry;

@Configuration
@EnableWebSocket
public class WebSocketConfig implements WebSocketConfigurer {
    
    @Override
    public void registerWebSocketHandlers(WebSocketHandlerRegistry registry) {
        registry.addHandler(videoProcessWebSocketHandler(), "/ws/video-process")
                .setAllowedOrigins("*");
    }
    
    public VideoProcessWebSocketHandler videoProcessWebSocketHandler() {
        return new VideoProcessWebSocketHandler();
    }
}
```

### 步骤 2：WebSocket 处理器

**新增文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\websocket\VideoProcessWebSocketHandler.java`

```java
package com.mybilibili.ai.websocket;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.CloseStatus;
import org.springframework.web.socket.TextMessage;
import org.springframework.web.socket.WebSocketSession;
import org.springframework.web.socket.handler.TextWebSocketHandler;

import java.util.Map;
import java.util.concurrent.CopyOnWriteArraySet;

@Component
public class VideoProcessWebSocketHandler extends TextWebSocketHandler {
    
    private static final CopyOnWriteArraySet<WebSocketSession> sessions = new CopyOnWriteArraySet<>();
    private static final ObjectMapper objectMapper = new ObjectMapper();
    
    @Override
    public void afterConnectionEstablished(WebSocketSession session) {
        sessions.add(session);
        System.out.println("[WebSocket] 新连接，当前连接数: " + sessions.size());
    }
    
    @Override
    public void afterConnectionClosed(WebSocketSession session, CloseStatus status) {
        sessions.remove(session);
        System.out.println("[WebSocket] 连接关闭，当前连接数: " + sessions.size());
    }
    
    @Override
    protected void handleTextMessage(WebSocketSession session, TextMessage message) throws Exception {
        // 可以处理客户端发来的消息
        System.out.println("[WebSocket] 收到消息: " + message.getPayload());
    }
    
    public static void broadcast(Map<String, Object> data) {
        try {
            String json = objectMapper.writeValueAsString(data);
            TextMessage message = new TextMessage(json);
            for (WebSocketSession session : sessions) {
                if (session.isOpen()) {
                    session.sendMessage(message);
                }
            }
        } catch (Exception e) {
            System.err.println("[WebSocket] 广播失败: " + e.getMessage());
        }
    }
    
    public static void broadcastProgress(Integer videoId, Integer manuscriptId, String videoTitle,
                                          String stage, String stageText, int progress, Integer status) {
        Map<String, Object> data = Map.of(
            "type", "progress",
            "videoId", videoId,
            "manuscriptId", manuscriptId,
            "videoTitle", videoTitle,
            "stage", stage,
            "stageText", stageText,
            "progress", progress,
            "status", status
        );
        broadcast(data);
    }
    
    public static void broadcastComplete(Integer videoId, Integer manuscriptId, String videoTitle) {
        Map<String, Object> data = Map.of(
            "type", "complete",
            "videoId", videoId,
            "manuscriptId", manuscriptId,
            "videoTitle", videoTitle
        );
        broadcast(data);
    }
    
    public static void broadcastError(Integer videoId, Integer manuscriptId, String videoTitle,
                                       String stage, String error) {
        Map<String, Object> data = Map.of(
            "type", "error",
            "videoId", videoId,
            "manuscriptId", manuscriptId,
            "videoTitle", videoTitle,
            "stage", stage,
            "error", error
        );
        broadcast(data);
    }
}
```

### 步骤 3：修改 VideoProcessConsumer

**修改文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\consumer\VideoProcessConsumer.java`

移除 Redis 相关代码，改用 WebSocket 推送：

```java
// 移除
// private static final String REDIS_KEY_CURRENT_TASK = "video:process:current";
// @Autowired private StringRedisTemplate redisTemplate;

// 导入 WebSocket
import com.mybilibili.ai.websocket.VideoProcessWebSocketHandler;

// 修改 setTaskStatus 方法
private void setTaskStatus(Integer videoId, Integer manuscriptId, String videoTitle, 
                           String processType, String stage, int progress, Integer status) {
    String stageText = getStageText(stage);
    
    // WebSocket 推送进度
    VideoProcessWebSocketHandler.broadcastProgress(
        videoId, manuscriptId, videoTitle, stage, stageText, progress, status
    );
    
    System.out.println("[进度] " + videoTitle + " - " + stageText + " (" + progress + "%)");
}

// 移除 clearTaskStatus 方法，不再需要
```

### 步骤 4：修改 VideoProcessAdminController

**修改文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\controller\VideoProcessAdminController.java`

移除 Redis 依赖，直接查询数据库：

```java
// 移除
// private static final String REDIS_KEY_CURRENT_TASK = "video:process:current";
// @Autowired private StringRedisTemplate redisTemplate;
// @Autowired private ObjectMapper objectMapper;

@GetMapping("/current")
public Result<Map<String, Object>> getCurrentTask() {
    // 查询正在处理中的视频（状态为 1, 2, 3, 4）
    Video currentVideo = videoMapper.selectProcessing();
    
    if (currentVideo == null) {
        Map<String, Object> emptyResult = new HashMap<>();
        emptyResult.put("processing", false);
        return Result.success("当前无处理任务", emptyResult);
    }
    
    Map<String, Object> taskInfo = new HashMap<>();
    taskInfo.put("processing", true);
    taskInfo.put("videoId", currentVideo.getId());
    taskInfo.put("manuscriptId", currentVideo.getManuscriptId());
    taskInfo.put("videoTitle", currentVideo.getTitle());
    taskInfo.put("status", currentVideo.getProcessStatus());
    taskInfo.put("statusText", getStatusText(currentVideo.getProcessStatus()));
    
    return Result.success("获取成功", taskInfo);
}

private String getStatusText(Integer status) {
    if (status == null) return "未知";
    switch (status) {
        case 1: return "视频转码中";
        case 2: return "音频提取中";
        case 3: return "字幕生成中";
        case 4: return "AI总结中";
        default: return "状态: " + status;
    }
}
```

### 步骤 5：VideoMapper 新增查询方法

**修改文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\mapper\VideoMapper.java`

```java
Video selectProcessing();

List<Video> selectByProcessStatus(@Param("processStatus") Integer processStatus);
```

**修改文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\resources\mapper\VideoMapper.xml`

```xml
<select id="selectProcessing" resultType="com.mybilibili.common.entity.Video">
    SELECT * FROM video 
    WHERE process_status IN (1, 2, 3, 4)
    LIMIT 1
</select>

<select id="selectByProcessStatus" resultType="com.mybilibili.common.entity.Video">
    SELECT * FROM video WHERE process_status = #{processStatus}
</select>
```

### 步骤 6：前端 WebSocket 连接

**修改文件**：`d:\files\mybilibili-next\mybilibili-admin-web\src\views\VideoProcessView.vue`

```javascript
import { ref, onMounted, onUnmounted } from 'vue'

const ws = ref(null)
const wsConnected = ref(false)

const connectWebSocket = () => {
  const wsUrl = 'ws://localhost:8088/ws/video-process'
  ws.value = new WebSocket(wsUrl)
  
  ws.value.onopen = () => {
    wsConnected.value = true
    console.log('WebSocket 已连接')
  }
  
  ws.value.onmessage = (event) => {
    const data = JSON.parse(event.data)
    handleWsMessage(data)
  }
  
  ws.value.onclose = () => {
    wsConnected.value = false
    console.log('WebSocket 已断开，3秒后重连')
    setTimeout(connectWebSocket, 3000)
  }
  
  ws.value.onerror = (error) => {
    console.error('WebSocket 错误:', error)
  }
}

const handleWsMessage = (data) => {
  switch (data.type) {
    case 'progress':
      // 更新当前任务显示
      currentTask.value = {
        processing: true,
        videoId: data.videoId,
        manuscriptId: data.manuscriptId,
        videoTitle: data.videoTitle,
        stage: data.stage,
        stageText: data.stageText,
        progress: data.progress,
        status: data.status
      }
      // 更新表格中对应视频的状态
      updateVideoStatus(data.videoId, data.status)
      break
      
    case 'complete':
      // 处理完成
      currentTask.value = { processing: false }
      ElMessage.success(`视频 "${data.videoTitle}" 处理完成`)
      loadVideos()
      loadStatistics()
      break
      
    case 'error':
      // 处理错误
      currentTask.value = { processing: false }
      ElMessage.error(`视频 "${data.videoTitle}" 处理失败: ${data.error}`)
      loadVideos()
      loadStatistics()
      break
  }
}

onMounted(() => {
  connectWebSocket()
  loadAllData()
})

onUnmounted(() => {
  if (ws.value) {
    ws.value.close()
  }
})
```

### 步骤 7：前端显示 WebSocket 连接状态

在页面头部添加连接状态指示器：

```vue
<div class="header-actions">
  <el-tag :type="wsConnected ? 'success' : 'danger'" size="small">
    {{ wsConnected ? '实时连接' : '已断开' }}
  </el-tag>
  <el-switch v-model="autoRefresh" active-text="自动刷新" @change="toggleAutoRefresh" />
  <el-button type="primary" :icon="Refresh" @click="loadAllData" :loading="loading">刷新</el-button>
</div>
```

---

## 数据流程

```
MQ消息到达 → VideoProcessConsumer 处理
    ↓
更新 MySQL video.processStatus
    ↓
WebSocket 广播进度
    ↓
前端实时更新 UI
```

---

## 文件清单

### 新增文件
1. `WebSocketConfig.java` - WebSocket 配置
2. `VideoProcessWebSocketHandler.java` - WebSocket 处理器

### 修改文件
1. `VideoProcessConsumer.java` - 移除 Redis，添加 WebSocket 推送
2. `VideoProcessAdminController.java` - 移除 Redis，改用数据库查询
3. `VideoMapper.java` - 新增 selectProcessing 方法
4. `VideoMapper.xml` - 新增 SQL
5. `VideoProcessView.vue` - WebSocket 连接和消息处理
