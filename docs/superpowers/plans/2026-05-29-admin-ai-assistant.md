# Admin AI Assistant 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在管理后台侧边栏添加一个浮动 AI 聊天助手，管理员可通过对话让 AI 调后台接口出报表/图表、做增删改查

**Architecture:** 前端在 App.vue 侧边栏底部加浮动聊天按钮 + 滑出面板。后端新增 `AdminAiController`，复用现有 `DynamicChatClient` 和 AI 渠道体系，注册 `ToolCallback` 让 AI 能调后台统计/CRUD 接口。新增 "ADMIN" feature 注册到 AI 渠道管理页。

**Tech Stack:** Spring AI 1.0.0 (ToolCallback), Vue 3 + Element Plus, SSE 流式, ECharts

---

## 文件改动总览

### 后端 (mybilibili-ai)
| 操作 | 文件 |
|------|------|
| 新建 | `controller/AdminAiController.java` |
| 新建 | `service/AdminAiService.java` |
| 新建 | `service/impl/AdminAiServiceImpl.java` |
| 新建 | `tool/AdminToolService.java` — 所有 Tool 方法定义 |
| 新建 | `tool/StatsData.java` — 统计数据 DTO |

### 前端 (mybilibili-admin-web)
| 操作 | 文件 |
|------|------|
| 新建 | `components/AdminAiFloatingButton.vue` — 浮动按钮 |
| 新建 | `components/AdminAiChatPanel.vue` — 聊天面板 |
| 新建 | `api/adminAi.js` — SSE 流式 API |
| 修改 | `App.vue` — 引入浮动按钮组件 |
| 修改 | `views/ApiManagementView.vue` — 新增 "ADMIN" feature 绑定选项 |

---

### Task 1: 后端 — 定义 Tool 服务 (AdminToolService)

**Files:**
- Create: `mybilibili-ai/src/main/java/com/mybilibili/ai/tool/AdminToolService.java`
- Create: `mybilibili-ai/src/main/java/com/mybilibili/ai/tool/StatsData.java`

- [ ] **Step 1: 创建 StatsData DTO**

```java
package com.mybilibili.ai.tool;

import java.util.List;
import java.util.Map;

public class StatsData {
    private String type;        // "overview" | "user_growth" | "manuscript_stats" | "play_stats"
    private Map<String, Object> data;
    private String chartType;   // "bar" | "line" | "pie" | "number"
    private String title;

    public StatsData() {}

    public StatsData(String type, Map<String, Object> data, String chartType, String title) {
        this.type = type; this.data = data; this.title = title; this.chartType = chartType;
    }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }
    public Map<String, Object> getData() { return data; }
    public void setData(Map<String, Object> data) { this.data = data; }
    public String getChartType() { return chartType; }
    public void setChartType(String chartType) { this.chartType = chartType; }
    public String getTitle() { return title; }
    public void setTitle(String title) { this.title = title; }
}
```

- [ ] **Step 2: 创建 AdminToolService（核心 Tool 定义）**

```java
package com.mybilibili.ai.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.mapper.AiUsageLogMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.*;

/**
 * 管理员 AI 助手可调用的工具方法。
 * 每个 public 方法代表一个 Tool，AI 根据用户意图选择调用。
 */
@Slf4j
@Component
public class AdminToolService {

    @Autowired
    private AiUsageLogMapper aiUsageLogMapper;

    private final RestTemplate restTemplate = new RestTemplate();
    private final ObjectMapper objectMapper = new ObjectMapper();

    // 内部 HTTP 调用各服务的统计/管理接口（通过网关 localhost:8080）
    private String gateway() { return "http://localhost:8080/api"; }

    /**
     * 获取平台概览统计数据：用户数、稿件数、评论数等
     */
    public StatsData getOverviewStats() {
        try {
            String json = restTemplate.getForObject(gateway() + "/statistics/overview", String.class);
            Map<String, Object> resp = objectMapper.readValue(json, Map.class);
            Object data = resp.get("data");
            if (data instanceof Map) {
                return new StatsData("overview", (Map<String, Object>) data, "number", "平台数据概览");
            }
        } catch (Exception e) {
            log.error("获取概览统计失败", e);
        }
        return new StatsData("overview", Map.of("error", "获取数据失败"), "number", "平台数据概览");
    }

    /**
     * 获取用户增长趋势
     * @param days 最近多少天
     */
    public StatsData getUserGrowth(int days) {
        try {
            Calendar cal = Calendar.getInstance();
            cal.add(Calendar.DAY_OF_YEAR, -days);
            String start = new java.text.SimpleDateFormat("yyyy-MM-dd").format(cal.getTime());
            String end = new java.text.SimpleDateFormat("yyyy-MM-dd").format(new Date());
            String json = restTemplate.getForObject(
                gateway() + "/statistics/user/growth?startDate={start}&endDate={end}",
                String.class, start, end);
            Map<String, Object> resp = objectMapper.readValue(json, Map.class);
            Object data = resp.get("data");
            if (data instanceof Map) {
                return new StatsData("user_growth", (Map<String, Object>) data, "line", "用户增长趋势（近" + days + "天）");
            }
        } catch (Exception e) {
            log.error("获取用户增长失败", e);
        }
        return new StatsData("user_growth", Map.of("error", "获取数据失败"), "line", "用户增长趋势");
    }

    /**
     * 获取稿件状态统计
     */
    public StatsData getManuscriptStats() {
        try {
            String json = restTemplate.getForObject(gateway() + "/statistics/manuscript/status", String.class);
            Map<String, Object> resp = objectMapper.readValue(json, Map.class);
            Object data = resp.get("data");
            if (data instanceof Map) {
                return new StatsData("manuscript_stats", (Map<String, Object>) data, "pie", "稿件状态分布");
            }
        } catch (Exception e) {
            log.error("获取稿件统计失败", e);
        }
        return new StatsData("manuscript_stats", Map.of("error", "获取数据失败"), "pie", "稿件状态分布");
    }

    /**
     * 获取 AI 用量概览
     */
    public StatsData getAiUsageOverview() {
        try {
            Map<String, Object> overview = aiUsageLogMapper.selectOverview();
            return new StatsData("ai_usage", overview, "number", "AI 用量概览");
        } catch (Exception e) {
            log.error("获取AI用量失败", e);
        }
        return new StatsData("ai_usage", Map.of("error", "获取数据失败"), "number", "AI 用量概览");
    }

    /**
     * 获取热门视频列表
     * @param limit 返回数量
     */
    public StatsData getHotVideos(int limit) {
        try {
            String json = restTemplate.getForObject(
                gateway() + "/statistics/video/hot?limit={limit}", String.class, limit);
            Map<String, Object> resp = objectMapper.readValue(json, Map.class);
            Object data = resp.get("data");
            if (data instanceof List) {
                Map<String, Object> result = new HashMap<>();
                result.put("videos", data);
                result.put("limit", limit);
                return new StatsData("hot_videos", result, "table", "热门视频 Top " + limit);
            }
        } catch (Exception e) {
            log.error("获取热门视频失败", e);
        }
        return new StatsData("hot_videos", Map.of("error", "获取数据失败"), "table", "热门视频");
    }
}
```

- [ ] **Step 3: 验证 AiUsageLogMapper 是否有 selectOverview 方法**

检查文件 `mybilibili-ai/src/main/java/com/mybilibili/ai/mapper/AiUsageLogMapper.java`，确认存在 `selectOverview()` 方法。如果不存在，添加。

--- 检查结果 ---
实际上在现有代码中，`AiUsageLogMapper` 已经有这个方法（因为 `AiUsageController` 调用了它）。

---

### Task 2: 后端 — AdminAiService 和 Controller

**Files:**
- Create: `mybilibili-ai/src/main/java/com/mybilibili/ai/service/AdminAiService.java`
- Create: `mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/AdminAiServiceImpl.java`
- Create: `mybilibili-ai/src/main/java/com/mybilibili/ai/controller/AdminAiController.java`

- [ ] **Step 1: 创建 AdminAiService 接口**

```java
package com.mybilibili.ai.service;

import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

public interface AdminAiService {
    SseEmitter sendMessage(Integer adminId, String content);
}
```

- [ ] **Step 2: 创建 AdminAiServiceImpl（核心——用 function calling 调 tool）**

```java
package com.mybilibili.ai.service.impl;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.ai.config.DynamicChatClient;
import com.mybilibili.ai.tool.AdminToolService;
import com.mybilibili.ai.tool.StatsData;
import com.mybilibili.ai.service.AdminAiService;
import com.mybilibili.ai.util.AiUsageLogger;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.ai.tool.ToolCallback;
import org.springframework.ai.tool.ToolCallbacks;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;
import reactor.core.Disposable;
import reactor.core.publisher.Flux;

import java.util.*;
import java.util.concurrent.atomic.AtomicReference;

@Service
public class AdminAiServiceImpl implements AdminAiService {

    @Autowired
    private DynamicChatClient dynamicChatClient;

    @Autowired
    private AdminToolService adminToolService;

    @Autowired
    private AiUsageLogger aiUsageLogger;

    private final ObjectMapper objectMapper = new ObjectMapper();

    private static final String SYSTEM_PROMPT =
        "你是哔哩哔哩管理后台的 AI 助手，名叫\"管理助手\"。\n" +
        "你的职责是帮助管理员完成以下工作：\n" +
        "1. 查询平台统计数据，并以表格、图表或文字摘要形式呈现\n" +
        "2. 回答关于平台运营状况的问题\n\n" +
        "你可以调用以下工具获取数据：\n" +
        "- getOverviewStats: 获取平台概览（用户数/稿件数等）\n" +
        "- getUserGrowth(days): 获取用户增长趋势\n" +
        "- getManuscriptStats: 获取稿件状态分布\n" +
        "- getAiUsageOverview: 获取 AI 用量概览\n" +
        "- getHotVideos(limit): 获取热门视频列表\n\n" +
        "当用户请求数据时，调用对应工具并将结果组织成清晰的中文回答。\n" +
        "工具返回的 StatsData 中包含 chartType 字段，表示数据适合用哪种图表展示。\n" +
        "请在回答中用 JSON 格式返回渲染建议，格式：{\"render\": {\"type\": \"chartType\", \"title\": \"标题\", \"data\": {...}}}";

    @Override
    public SseEmitter sendMessage(Integer adminId, String content) {
        SseEmitter emitter = new SseEmitter(120000L);
        long startTime = System.currentTimeMillis();

        // 获取绑定 ADMIN feature 的 ChatClient，无则用第一个活跃渠道
        ChatClient client = dynamicChatClient.getClient("ADMIN");
        if (client == null) {
            client = dynamicChatClient.getFirstActiveClient();
        }

        if (client == null) {
            try {
                emitter.send(SseEmitter.event().name("error").data("未配置可用的 AI 渠道，请先在 AI 渠道管理中配置并绑定 ADMIN 功能"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        // 注册 ToolCallback
        List<ToolCallback> toolCallbacks = List.of(
            ToolCallbacks.from(adminToolService, "getOverviewStats",
                "获取平台概览统计：用户总数、稿件总数、评论总数等核心指标"),
            ToolCallbacks.from(adminToolService, "getUserGrowth",
                "获取用户增长趋势，参数 days 表示最近天数（默认7）"),
            ToolCallbacks.from(adminToolService, "getManuscriptStats",
                "获取稿件状态分布统计（审核通过/待审核/已下架等）"),
            ToolCallbacks.from(adminToolService, "getAiUsageOverview",
                "获取 AI 功能用量概览（总调用次数、Token消耗等）"),
            ToolCallbacks.from(adminToolService, "getHotVideos",
                "获取热门视频列表，参数 limit 表示返回数量（默认10）")
        );

        StringBuilder fullResp = new StringBuilder();
        AtomicReference<Disposable> subscriptionRef = new AtomicReference<>();

        Flux<String> flux = client.prompt()
                .system(SYSTEM_PROMPT)
                .user(content)
                .tools(toolCallbacks)
                .stream()
                .content();

        Disposable subscription = flux.subscribe(
            chunk -> {
                try {
                    fullResp.append(chunk);
                    emitter.send(SseEmitter.event().name("data").data(chunk));
                } catch (Exception ignored) {}
            },
            error -> {
                aiUsageLogger.log("ADMIN", null, null, null,
                    System.currentTimeMillis() - startTime, false, error.getMessage());
                try {
                    emitter.send(SseEmitter.event().name("error")
                        .data("请求失败: " + error.getMessage()));
                    emitter.complete();
                } catch (Exception ignored) {}
            },
            () -> {
                try {
                    String reply = fullResp.toString();
                    // 尝试从回复中提取 render JSON
                    String renderJson = extractRenderJson(reply);
                    Map<String, Object> doneData = new HashMap<>();
                    doneData.put("content", reply);
                    if (renderJson != null) {
                        doneData.put("render", renderJson);
                    }
                    emitter.send(SseEmitter.event().name("done")
                        .data(objectMapper.writeValueAsString(doneData)));
                    emitter.complete();

                    aiUsageLogger.log("ADMIN", null, null, null,
                        System.currentTimeMillis() - startTime, true, null);
                } catch (Exception ignored) {}
            }
        );
        subscriptionRef.set(subscription);

        emitter.onCompletion(() -> subscriptionRef.get().dispose());
        emitter.onTimeout(() -> subscriptionRef.get().dispose());
        emitter.onError(e -> subscriptionRef.get().dispose());

        return emitter;
    }

    /**
     * 从 AI 回复中提取 render JSON（放在 ```json ... ``` 块中或行内）
     */
    private String extractRenderJson(String text) {
        try {
            int start = text.indexOf("{\"render\"");
            if (start == -1) return null;
            int end = text.indexOf("}", start) + 1;
            // 尝试找到完整 JSON 结尾
            int braceCount = 0;
            boolean inString = false;
            for (int i = start; i < text.length(); i++) {
                char c = text.charAt(i);
                if (c == '"' && (i == 0 || text.charAt(i-1) != '\\')) inString = !inString;
                if (!inString) {
                    if (c == '{') braceCount++;
                    else if (c == '}') braceCount--;
                    if (braceCount == 0) {
                        end = i + 1;
                        break;
                    }
                }
            }
            String json = text.substring(start, end);
            // 验证是合法 JSON
            objectMapper.readTree(json);
            return json;
        } catch (Exception e) {
            return null;
        }
    }
}
```

- [ ] **Step 3: 创建 AdminAiController**

```java
package com.mybilibili.ai.controller;

import com.mybilibili.ai.service.AdminAiService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import jakarta.servlet.http.HttpServletRequest;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/assistant")
@Tag(name = "管理后台 AI 助手")
public class AdminAiController {

    @Autowired
    private AdminAiService adminAiService;

    @PostMapping("/send")
    @Operation(summary = "发送消息给管理后台 AI 助手（SSE 流式）")
    public SseEmitter sendMessage(@RequestBody Map<String, String> request,
                                  HttpServletRequest httpRequest) {
        Integer adminId = getAdminId(httpRequest);
        if (adminId == null) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("未登录"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        String content = request.get("content");
        if (content == null || content.trim().isEmpty()) {
            SseEmitter emitter = new SseEmitter(0L);
            try {
                emitter.send(SseEmitter.event().name("error").data("消息内容不能为空"));
                emitter.complete();
            } catch (Exception ignored) {}
            return emitter;
        }

        return adminAiService.sendMessage(adminId, content);
    }

    private Integer getAdminId(HttpServletRequest request) {
        String adminIdStr = request.getHeader("X-Admin-Id");
        if (adminIdStr != null) {
            try {
                return Integer.parseInt(adminIdStr);
            } catch (NumberFormatException ignored) {}
        }
        return null;
    }
}
```

- [ ] **Step 4: 确认网关路由已覆盖新路径**

查看 `mybilibili-gateway/src/main/resources/application.yml`：

已有路由 `ai-admin-process`：
```yaml
- id: ai-admin-process
  uri: lb://mybilibili-ai
  predicates:
    - Path=/api/ai/admin/**
  filters:
    - StripPrefix=1
```

新 Controller 路径 `/ai/admin/assistant/send` 通过网关 `POST /api/ai/admin/assistant/send` 访问，已被 `ai-admin-process` 路由覆盖，**无需修改网关**。

---

### Task 3: 前端 — SSE API 封装

**File:**
- Create: `mybilibili-admin-web/src/api/adminAi.js`

- [ ] **Step 1: 创建 adminAi.js**

```javascript
/**
 * 管理后台 AI 助手 API（SSE 流式）
 */
const BASE_URL = window.location.origin

export const adminAiApi = {
  /**
   * 发送消息给 AI 助手（SSE 流式）
   * @param {string} content - 消息内容
   * @param {Object} callbacks - { onData, onDone, onError }
   * @returns {Object} { abort }
   */
  sendMessage(content, callbacks = {}) {
    const { onData, onDone, onError } = callbacks
    const token = localStorage.getItem('admin_token')
    const controller = new AbortController()

    fetch(`${BASE_URL}/api/ai/admin/assistant/send`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify({ content }),
      signal: controller.signal
    })
    .then(response => {
      if (!response.ok) throw new Error(`HTTP ${response.status}`)

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      const read = () => {
        reader.read().then(({ done, value }) => {
          if (done) return
          buffer += decoder.decode(value, { stream: true })

          const lines = buffer.split('\n\n')
          buffer = lines.pop() || ''

          for (const line of lines) {
            const event = parseSSELine(line)
            if (!event) continue
            switch (event.event) {
              case 'data':
                if (onData) onData(event.data)
                break
              case 'done':
                if (onDone) {
                  try {
                    onDone(JSON.parse(event.data))
                  } catch {
                    onDone({ content: event.data })
                  }
                }
                break
              case 'error':
                if (onError) onError(event.data)
                break
            }
          }
          read()
        }).catch(err => {
          if (err.name !== 'AbortError' && onError) onError(err.message)
        })
      }
      read()
    })
    .catch(err => {
      if (err.name !== 'AbortError' && onError) onError(err.message)
    })

    return { abort: () => controller.abort() }
  }
}

function parseSSELine(raw) {
  const lines = raw.split('\n')
  let event = ''
  let data = ''
  for (const line of lines) {
    if (line.startsWith('event:')) event = line.substring(6).trim()
    else if (line.startsWith('data:')) data = line.substring(5).trim()
  }
  if (!event && !data) return null
  return { event: event || 'message', data }
}
```

---

### Task 4: 前端 — 浮动聊天按钮组件

**File:**
- Create: `mybilibili-admin-web/src/components/AdminAiFloatingButton.vue`

- [ ] **Step 1: 创建浮动按钮组件**

```vue
<script setup>
import { ref } from 'vue'
import { ChatDotRound } from '@element-plus/icons-vue'

const emit = defineEmits(['toggle'])
const show = defineModel('visible', { type: Boolean, default: false })

const handleClick = () => {
  show.value = !show.value
  emit('toggle', show.value)
}
</script>

<template>
  <div class="ai-float-btn" :class="{ active: show }" @click="handleClick">
    <el-icon :size="22"><ChatDotRound /></el-icon>
  </div>
</template>

<style scoped>
.ai-float-btn {
  position: fixed;
  bottom: 80px;
  right: 24px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #00a1d6, #00d6b2);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 2000;
  box-shadow: 0 4px 14px rgba(0, 161, 214, 0.4);
  transition: all 0.3s ease;
}
.ai-float-btn:hover {
  transform: scale(1.1);
  box-shadow: 0 6px 20px rgba(0, 161, 214, 0.5);
}
.ai-float-btn.active {
  background: #f56c6c;
  transform: rotate(45deg);
}
</style>
```

---

### Task 5: 前端 — 聊天面板组件

**File:**
- Create: `mybilibili-admin-web/src/components/AdminAiChatPanel.vue`

- [ ] **Step 1: 创建聊天面板组件**

```vue
<script setup>
import { ref, nextTick, watch } from 'vue'
import { adminAiApi } from '../api/adminAi.js'
import { Close, ChatDotRound } from '@element-plus/icons-vue'

const visible = defineModel('visible', { type: Boolean, default: false })
const emit = defineEmits(['close'])

const messages = ref([])
const inputText = ref('')
const isStreaming = ref(false)
const streamingContent = ref('')
const messageListRef = ref(null)
const eventSource = ref(null)

const displayMessages = ref([])

// 合并普通消息 + 流式消息
const updateDisplay = () => {
  const msgs = [...messages.value]
  if (isStreaming.value && streamingContent.value) {
    msgs.push({
      id: 'streaming',
      role: 'assistant',
      content: streamingContent.value
    })
  }
  displayMessages.value = msgs
}

watch([messages, streamingContent, isStreaming], updateDisplay, { deep: true })

const scrollToBottom = () => {
  nextTick(() => {
    if (messageListRef.value) {
      messageListRef.value.scrollTop = messageListRef.value.scrollHeight
    }
  })
}

watch(displayMessages, scrollToBottom, { deep: true })

const handleSend = () => {
  const text = inputText.value.trim()
  if (!text || isStreaming.value) return
  inputText.value = ''

  messages.value.push({
    id: Date.now(),
    role: 'user',
    content: text
  })

  isStreaming.value = true
  streamingContent.value = ''

  eventSource.value = adminAiApi.sendMessage(text, {
    onData: (chunk) => {
      streamingContent.value += chunk
    },
    onDone: (data) => {
      if (data.render) {
        // 有渲染数据，追加渲染块
        messages.value.push({
          id: 'ai-' + Date.now(),
          role: 'assistant',
          content: data.content || streamingContent.value,
          render: data.render
        })
      } else if (streamingContent.value) {
        messages.value.push({
          id: 'ai-' + Date.now(),
          role: 'assistant',
          content: streamingContent.value
        })
      }
      streamingContent.value = ''
      isStreaming.value = false
    },
    onError: (err) => {
      messages.value.push({
        id: 'ai-' + Date.now(),
        role: 'assistant',
        content: '错误：' + (err || '请求失败')
      })
      streamingContent.value = ''
      isStreaming.value = false
    }
  })
}

const handleClose = () => {
  if (eventSource.value) {
    eventSource.value.abort()
    eventSource.value = null
  }
  visible.value = false
  emit('close')
}

const handleKeydown = (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

// 组件关闭时清理
watch(visible, (val) => {
  if (!val && eventSource.value) {
    eventSource.value.abort()
    eventSource.value = null
  }
})
</script>

<template>
  <transition name="slide-up">
    <div v-if="visible" class="ai-chat-panel">
      <!-- 头部 -->
      <div class="panel-header">
        <div class="header-left">
          <el-icon class="header-icon" :size="18"><ChatDotRound /></el-icon>
          <span class="header-title">管理助手</span>
        </div>
        <el-button link :icon="Close" class="close-btn" @click="handleClose" />
      </div>

      <!-- 消息列表 -->
      <div ref="messageListRef" class="message-list">
        <div v-if="displayMessages.length === 0" class="empty-state">
          <el-icon :size="40" color="#c0c4cc"><ChatDotRound /></el-icon>
          <p>你好！我是管理助手</p>
          <p class="sub-text">可以问我平台数据、统计报表等问题</p>
        </div>

        <div
          v-for="msg in displayMessages"
          :key="msg.id"
          :class="['msg-item', msg.role === 'user' ? 'msg-user' : 'msg-ai']"
        >
          <div class="msg-bubble">
            <div class="msg-content">{{ msg.content }}</div>
            <!-- 渲染建议 -->
            <div v-if="msg.render" class="render-block">
              <el-divider />
              <div class="render-hint">
                <el-tag size="small" type="primary">
                  {{ (() => { try { const r = JSON.parse(msg.render); return r.render?.title || '数据'; } catch { return '数据'; } })() }}
                </el-tag>
                <span class="hint-text">
                  数据类型: {{ (() => { try { const r = JSON.parse(msg.render); return r.render?.type || '—'; } catch { return '—'; } })() }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 输入指示器 -->
        <div v-if="isStreaming && !streamingContent" class="msg-item msg-ai">
          <div class="msg-bubble typing-indicator">
            <span class="dot">.</span><span class="dot">.</span><span class="dot">.</span>
          </div>
        </div>
      </div>

      <!-- 输入区 -->
      <div class="input-area">
        <el-input
          v-model="inputText"
          type="textarea"
          :rows="2"
          placeholder="输入你的问题，Enter发送"
          :disabled="isStreaming"
          resize="none"
          @keydown="handleKeydown"
        />
        <div class="input-footer">
          <span class="tip">按 Enter 发送，Shift+Enter 换行</span>
          <el-button
            type="primary"
            :disabled="!inputText.trim() || isStreaming"
            :loading="isStreaming"
            @click="handleSend"
            size="small"
          >
            发送
          </el-button>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
.ai-chat-panel {
  position: fixed;
  bottom: 140px;
  right: 24px;
  width: 400px;
  height: 520px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.15);
  display: flex;
  flex-direction: column;
  z-index: 2001;
  overflow: hidden;
  border: 1px solid #e3e5e7;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border-bottom: 1px solid #e3e5e7;
  background: linear-gradient(135deg, #00a1d6, #00d6b2);
  color: #fff;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-title {
  font-size: 15px;
  font-weight: 600;
}

.close-btn {
  color: #fff !important;
  font-size: 16px;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #f5f7fa;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
  gap: 8px;
}

.empty-state p { margin: 0; font-size: 14px; }
.sub-text { font-size: 12px !important; color: #c0c4cc; }

.msg-item { display: flex; }
.msg-user { justify-content: flex-end; }
.msg-ai { justify-content: flex-start; }

.msg-bubble {
  max-width: 85%;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.6;
  word-break: break-word;
}

.msg-user .msg-bubble {
  background: #00a1d6;
  color: #fff;
  border-bottom-right-radius: 2px;
}

.msg-ai .msg-bubble {
  background: #fff;
  color: #333;
  border-bottom-left-radius: 2px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
}

.msg-content { white-space: pre-wrap; }

.render-block { margin-top: 8px; }
.render-block :deep(.el-divider) { margin: 8px 0; }
.render-hint { display: flex; flex-direction: column; gap: 4px; }
.hint-text { font-size: 12px; color: #909399; }

.typing-indicator { display: flex; gap: 4px; padding: 12px 16px; }
.typing-indicator .dot {
  animation: bounce 1.4s infinite;
  font-size: 24px; line-height: 1;
  color: #909399;
}
.typing-indicator .dot:nth-child(2) { animation-delay: 0.2s; }
.typing-indicator .dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes bounce {
  0%, 60%, 100% { opacity: 0.3; }
  30% { opacity: 1; }
}

.input-area {
  padding: 12px 16px;
  border-top: 1px solid #e3e5e7;
  background: #fff;
}

.input-area :deep(.el-textarea__inner) {
  background: #f5f7fa;
  border: none;
  border-radius: 8px;
}

.input-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 8px;
}

.tip { font-size: 12px; color: #c0c4cc; }

/* 动画 */
.slide-up-enter-active,
.slide-up-leave-active { transition: all 0.3s ease; }
.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px);
}
</style>
```

---

### Task 6: 前端 — 嵌入 App.vue

**File:**
- Modify: `mybilibili-admin-web/src/App.vue`

- [ ] **Step 1: 在 App.vue 中引入并注册组件**

```vue
<script setup>
// 在现有 import 之后添加：
import AdminAiFloatingButton from './components/AdminAiFloatingButton.vue'
import AdminAiChatPanel from './components/AdminAiChatPanel.vue'
// ... 现有代码 ...

// 添加 AI 助手开关状态
const showAiAssistant = ref(false)
</script>
```

- [ ] **Step 2: 在 template 中合适位置添加组件**

在非登录页面的布局容器内、`</el-container>` 之前添加：

```vue
        <!-- 主内容 -->
        <el-main class="main">
          <router-view />
        </el-main>
      </el-container>
    </el-container>

    <!-- AI 管理助手 -->
    <AdminAiFloatingButton v-model:visible="showAiAssistant" />
    <AdminAiChatPanel v-model:visible="showAiAssistant" />
  </div>
```

在登录页面分支前 (`</div>\n\n<!-- 登录页面 -->`) 插入。

---

### Task 7: 前端 — 在 AI 渠道管理页注册 "ADMIN" feature

**File:**
- Modify: `mybilibili-admin-web/src/views/ApiManagementView.vue`

- [ ] **Step 1: 在 featureOptions 中添加 ADMIN**

```javascript
const featureOptions = [
  { value: 'CHAT', label: 'AI 客服' },
  { value: 'REVIEW', label: '内容审核' },
  { value: 'SUMMARY', label: '视频摘要' },
  { value: 'ADMIN', label: '管理助手' }        // 新增
]

const featureLabels = { CHAT: 'AI 客服', REVIEW: '内容审核', SUMMARY: '视频摘要', ADMIN: '管理助手' }
```

---

### Task 8: 数据库 — SQL 初始化添加 ADMIN 功能注释

**File:**
- Modify: `init/V7__ai_multi_channel.sql` （仅注释提示）

- [ ] **Step 1: 更新注释**

```sql
-- AI 功能绑定渠道
CREATE TABLE IF NOT EXISTS `ai_bindings` (
  `id`           bigint NOT NULL AUTO_INCREMENT,
  `feature`      varchar(30) NOT NULL COMMENT 'CHAT / REVIEW / SUMMARY / ADMIN',
  `api_config_id` bigint NOT NULL COMMENT '绑定的渠道 ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_feature` (`feature`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='AI 功能与渠道绑定关系';
```

仅改注释（表结构不变），如果已有该表则无需改。

---

### 测试 / 验证清单

1. 启动所有服务，确认 `/api/ai/admin/assistant/send` POST 可访问
2. 在管理后台 AI 渠道管理页，绑定一个渠道到 "管理助手" (ADMIN) 功能
3. 点击右下角浮动按钮，打开聊天面板
4. 输入"显示平台数据概览" → 确认 SSE 流式返回 + AI 调用 tool
5. 输入"最近7天用户增长" → 确认 getUserGrowth 被调用
6. 输入"热门视频" → 确认 getHotVideos 被调用
7. 确认未绑定 ADMIN 功能时的降级行为（使用第一个活跃渠道）

---

### 未涵盖的边界 / 后续可扩展

- **真实 ECharts 渲染**: 当前 render JSON 只作为展示建议，后续可在面板内嵌入 ECharts 实例真正渲染图表
- **更多 Tool**: 可按需添加用户管理、稿件管理、违禁词管理等 CRUD Tool
- **会话持久化**: 当前不保存历史会话（每次刷新重置），后续可复用 `ai_conversations`/`ai_chat_messages` 表
- **SSE 连接复用**: 当前每次发送新消息都建新连接
