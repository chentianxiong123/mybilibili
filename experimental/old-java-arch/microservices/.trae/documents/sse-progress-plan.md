# SSE实时推送视频处理进度方案

## 目标
将视频处理进度从轮询改为服务器推送，实时更新前端进度条。

## 实现步骤

### 1. 后端 - mybilibili-ai 模块

#### 1.1 创建进度推送服务
- 新建 `VideoProgressSseService.java`
- 使用 `Map<videoId, SseEmitter>` 存储连接
- 提供 `pushProgress(videoId, progress, status)` 方法
- 连接超时设置为 30 分钟（处理时间可能较长）

#### 1.2 修改 VideoProcessConsumer
- 注入 `VideoProgressSseService`
- 转码/字幕生成过程中调用 `pushProgress` 推送实时进度
- 完成后发送最终状态并关闭连接

#### 1.3 创建 SSE 端点
- 新建 `VideoProcessSseController.java`
- `GET /api/video/process/sse/{videoId}` - 获取实时进度流

### 2. 前端 - mybilibili-admin-web 模块

#### 2.1 修改 VideoProcessView.vue
- 使用 `EventSource` 连接 SSE
- 监听 `message` 事件接收进度更新
- 更新 `currentTask` 的 `progress` 和 `status`
- 组件卸载时关闭连接

### 3. 进度数据格式
```json
{
  "videoId": 27,
  "title": "视频标题",
  "progress": 50,
  "status": "字幕生成中",
  "stage": "subtitle"
}
```

## 文件清单

### 新建文件
- `mybilibili-ai/src/main/java/com/mybilibili/ai/service/VideoProgressSseService.java`
- `mybilibili-ai/src/main/java/com/mybilibili/ai/controller/VideoProcessSseController.java`

### 修改文件
- `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/VideoTranscodeServiceImpl.java` - 注入推送服务
- `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/AiSubtitleServiceImpl.java` - 注入推送服务
- `mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessConsumer.java` - 注入推送服务
- `mybilibili-admin-web/src/views/VideoProcessView.vue` - SSE连接和进度更新
