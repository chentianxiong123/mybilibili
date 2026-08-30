# 视频处理状态显示优化计划

## 一、现状分析

### 1.1 数据库表结构
**videos表**：
- `process_status` - 视频处理状态
- `process_progress` - 处理进度百分比
- `process_stage` - 处理阶段描述
- `has_subtitle` - 是否有字幕
- `has_summary` - 是否有AI摘要
- `process_error` - 处理错误信息

**manuscripts表**：
- `status` - 稿件状态
- `review_status` - 审核状态

### 1.2 状态码定义（Video.java）
```java
public static final int PROCESS_STATUS_PENDING = 0;        // 待处理
public static final int PROCESS_STATUS_TRANSCODING = 1;        // 转码中
public static final int PROCESS_STATUS_TRANSCODE_FAILED = 10;   // 转码失败
public static final int PROCESS_STATUS_TRANSCODE_SUCCESS = 11;  // 转码成功
public static final int PROCESS_STATUS_AUDIO_EXTRACTING = 2;    // 音频提取中
public static final int PROCESS_STATUS_AUDIO_FAILED = 20;   // 音频提取失败
public static final int PROCESS_STATUS_AUDIO_SUCCESS = 21;   // 音频提取成功
public static final int PROCESS_STATUS_SUBTITLE_GENERATING = 3; // 字幕生成中
public static final int PROCESS_STATUS_SUBTITLE_FAILED = 30;  // 字幕生成失败
public static final int PROCESS_STATUS_SUBTITLE_SUCCESS = 31; // 字幕生成成功
public static final int PROCESS_STATUS_AI_SUMMARIZING = 4;    // AI总结中
public static final int PROCESS_STATUS_AI_FAILED = 40;     // AI总结失败
public static final int PROCESS_STATUS_AI_SUCCESS = 41;     // AI总结成功
public static final int PROCESS_STATUS_COMPLETED = 5;        // 全部完成
```

这个状态码设计非常合理：
- 十位数表示阶段（1=转码, 2=音频, 3=字幕, 4=AI）
- 个位数0=失败，- 个位数1=成功

### 1.3 稿件状态（Manuscript.java）
```java
public static final int STATUS_PENDING_REVIEW = 0;  // 待审核
public static final int STATUS_PROCESSING = 1;        // 处理中
public static final int STATUS_READY_TO_PUBLISH = 2;  // 待上架
public static final int STATUS_PUBLISHED = 3;        // 已发布
public static final int STATUS_REJECTED = 4;        // 已拒绝
public static final int STATUS_PROCESS_FAILED = 5;  // 处理失败
```

---

## 二、问题分析

### 2.1 当前问题
1. **前端状态文本映射不完整**：VideoProcessView.vue 中的 `getProcessStatusText` 方法没有覆盖所有状态
2. **稿件处理状态计算缺失**：没有根据稿件下所有视频的处理状态来计算稿件整体状态

3. **审核后才能处理的逻辑缺失**：前端没有判断稿件是否审核通过

### 2.2 需要修复的内容
1. **完善前端状态文本映射**
2. **添加稿件整体状态计算逻辑**
3. **添加审核状态判断**

---

## 三、实现步骤

### 3.1 后端修改

#### 步骤1：添加稿件整体状态计算接口
**文件**：`VideoProcessAdminController.java`

添加接口：`GET /ai/admin/process/manuscript/{manuscriptId}/status`

返回稿件的整体处理状态：
- 如果稿件下所有视频都处理完成（process_status = 5），返回"completed"
- 如果有视频正在处理中，返回"processing"
- 如果所有视频都待处理，返回"pending"
- 如果有视频处理失败，返回"failed"

#### 步骤2：修改消费者逻辑
**文件**：`VideoProcessConsumer.java`

确保状态更新逻辑正确

### 3.2 前端修改

#### 步骤1：完善状态文本映射
**文件**：`VideoProcessView.vue`

修改 `getProcessStatusText` 方法，添加所有状态的文本：
```javascript
const getProcessStatusText = (processStatus) => {
  const statusMap = {
    0: '待处理',
    1: '视频转码中',
    10: '转码失败',
    11: '转码成功',
    2: '音频提取中',
    20: '音频提取失败',
    21: '音频提取成功',
    3: '字幕生成中',
    30: '字幕生成失败',
    31: '字幕生成成功',
    4: 'AI总结中',
    40: 'AI总结失败',
    41: 'AI总结成功',
    5: '处理完成'
  }
  return statusMap[processStatus] || '未知(' + processStatus + ')'
}
```

#### 步骤2：添加稿件状态显示
在视频列表中添加稿件状态列，显示稿件的整体处理状态

#### 步骤3：添加审核状态判断
只有 `review_status = 1`（审核通过）的稿件才能触发处理

---

## 四、详细任务清单

| 序号 | 任务 | 文件 | 优先级 |
|------|------|------|--------|
| 1 | 添加稿件整体状态计算接口 | VideoProcessAdminController.java | 高 |
| 2 | 完善前端状态文本映射 | VideoProcessView.vue | 高 |
| 3 | 添加稿件状态显示列 | VideoProcessView.vue | 高 |
| 4 | 添加审核状态判断 | VideoProcessView.vue | 中 |

---

## 五、接口设计

### 5.1 获取稿件处理状态
```
GET /api/ai/admin/process/manuscript/{manuscriptId}/status

Response:
{
  "code": 200,
  "data": {
    "manuscriptId": 123,
    "status": "completed",  // completed, processing, pending, failed
    "totalVideos": 3,
    "processedVideos": 3,
    "processingVideos": 0,
    "failedVideos": 0,
    "pendingVideos": 0
  }
}
```
