# 稿件审核通过后自动处理视频流水线 Spec

## Why

当前视频处理需要管理员在视频处理页面手动点击每个步骤按钮来触发任务。稿件审核通过后，管理员需要额外操作才能处理视频，效率较低。需要实现稿件审核通过时可选自动触发全流程处理的功能。

## What Changes

* 稿件管理页面"审核通过"按钮改为弹出选项，让管理员选择是否自动处理视频

* 后端新增"一键全流程处理"接口，支持自动按顺序执行转码→音频→字幕→AI

* SSE实时更新流水线步骤按钮颜色（灰色→蓝色→绿色/红色）

* 稿件管理页面显示"处理状态"而非"转码状态"

## Impact

* Affected specs: admin-features-implementation

* Affected code:

  * mybilibili-admin-web/src/views/ManuscriptsView\.vue

  * mybilibili-admin-web/src/views/VideoProcessView\.vue

  * mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/VideoTranscodeServiceImpl.java

  * mybilibili-ai/src/main/java/com/mybilibili/ai/service/impl/AiSubtitleServiceImpl.java

  * mybilibili-ai/src/main/java/com/mybilibili/ai/consumer/VideoProcessConsumer.java

## ADDED Requirements

### Requirement: 审核通过时可选自动处理视频

系统应在稿件审核通过时，让管理员选择是否自动处理该稿件下所有视频的所有处理步骤。

#### Scenario: 审核通过时选择自动处理

* **WHEN** 管理员点击"审核通过"按钮

* **THEN** 弹出确认对话框，包含选项：

  * "仅审核通过"：只更新稿件状态，不处理视频

  * "审核通过并处理视频"：更新稿件状态，并自动将该稿件下所有视频的所有处理步骤加入队列

#### Scenario: 选择自动处理后

* **WHEN** 管理员选择"审核通过并处理视频"

* **THEN** 系统执行以下操作：

  1. 更新稿件状态为"已通过"
  2. 获取该稿件下所有视频
  3. 为每个视频创建全流程任务队列（转码→音频→字幕→AI）
  4. 按单线程顺序执行任务

### Requirement: 全流程自动执行

系统应支持自动按顺序执行视频的所有处理步骤，无需人工干预。

#### Scenario: 全流程执行顺序

* **WHEN** 视频开始全流程处理

* **THEN** 系统按以下顺序自动执行：

  1. 转码（processStatus=1）
  2. 转码完成后自动触发音频提取（processStatus=2）
  3. 音频提取完成后自动触发字幕生成（processStatus=3）
  4. 字幕生成完成后自动触发AI总结（processStatus=4）
  5. 所有步骤完成后状态变为"处理完成"（processStatus=5）

#### Scenario: 步骤失败处理

* **WHEN** 某个步骤执行失败

* **THEN** 系统停止该视频的后续步骤，状态变为对应失败状态（10/20/30/40）

* **AND** 继续处理队列中的下一个视频

### Requirement: 流水线按钮颜色实时更新

系统应通过SSE实时更新流水线步骤按钮的颜色，反映当前任务状态。

#### Scenario: 按钮颜色规则

* **WHEN** 任务状态变化

* **THEN** 流水线步骤按钮颜色按以下规则更新：

  * 灰色（pending）：该步骤未开始

  * 蓝色（current）：该步骤正在执行中

  * 绿色（completed）：该步骤已完成

  * 红色（error）：该步骤执行失败

#### Scenario: SSE推送状态更新

* **WHEN** 后端完成某个步骤

* **THEN** 通过SSE推送状态更新消息

* **AND** 前端实时更新对应步骤的按钮颜色

### Requirement: 稿件管理页面状态显示优化

稿件管理页面应显示"处理状态"而非"转码状态"，更准确地反映视频处理进度。

#### Scenario: 处理状态显示

* **WHEN** 稿件下有视频正在处理

* **THEN** 稿件管理页面显示"处理状态"列

* **AND** 显示当前处理进度（如"转码中"、"音频提取中"等）

#### Scenario: 全部完成显示

* **WHEN** 稿件下所有视频的所有步骤都完成

* **THEN** 稿件管理页面显示"已完成"

## MODIFIED Requirements

### Requirement: 审核通过按钮交互

原有的"审核通过"直接确认改为弹出选项对话框。

**原实现**：

```javascript
await ElMessageBox.confirm('确定审核通过该稿件吗？', '审核通过', { type: 'warning' })
```

**新实现**：

```javascript
// 弹出选项对话框，让管理员选择是否自动处理视频
```

## Technical Design

### 前端改动

1. **ManuscriptsView\.vue**

   * 修改`handleApprove`方法，弹出选项对话框

   * 新增"审核通过并处理视频"选项

   * 调用新的后端接口

2. **VideoProcessView\.vue**

   * 确保SSE正确更新流水线按钮颜色

   * 优化`getStepClass`函数的颜色映射

### 后端改动

1. **新增接口**

   * `POST /api/manuscript/admin/{id}/approve-with-process`

   * 参数：`autoProcess: boolean`

   * 功能：审核通过稿件，可选自动处理视频

2. **全流程处理服务**

   * 创建任务队列，按顺序执行

   * 每个步骤完成后自动触发下一步

   * 通过SSE推送状态更新

### SSE消息格式

```json
{
  "event": "progress",
  "data": {
    "videoId": 123,
    "stage": "transcode",
    "progress": 100,
    "status": "转码完成",
    "nextStage": "audio"
  }
}
```

