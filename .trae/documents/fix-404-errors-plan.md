# 修复 404 错误 - 接口对齐计划

## 问题分析

多个接口返回 404 Not Found，主要涉及：
- 评论接口：`/api/comment/manuscript/31`
- 弹幕接口：`/api/manuscript/31/danmakus`
- 相关推荐接口：`/api/recommend/related/31`
- 字幕接口：`/api/subtitle/video/31`

## 关键概念
- **稿件（manuscript）**：互动、评论、消息的单位
- **视频（video）**：字幕、弹幕的单位

## 修复计划

### [ ] 任务 1：检查 Gateway 路由配置
- **优先级**：P0
- **依赖**：无
- **描述**：检查 Gateway 中相关接口的路由配置
- **成功标准**：确认所有 404 接口在 Gateway 中有正确的路由配置
- **测试要求**：
  - `programmatic`：Gateway 路由配置中存在对应路径的路由规则

### [ ] 任务 2：检查 Comment 服务控制器
- **优先级**：P0
- **依赖**：任务 1
- **描述**：检查 Comment 服务是否有 `/comment/manuscript/{id}` 接口
- **成功标准**：Comment 服务提供正确的评论接口
- **测试要求**：
  - `programmatic`：Comment 服务控制器存在对应接口

### [ ] 任务 3：检查 Interaction 服务弹幕接口
- **优先级**：P0
- **依赖**：任务 1
- **描述**：检查 Interaction 服务是否有 `/manuscript/{id}/danmakus` 接口
- **成功标准**：Interaction 服务提供正确的弹幕接口
- **测试要求**：
  - `programmatic`：Interaction 服务控制器存在对应接口

### [ ] 任务 4：检查 Recommend 服务相关推荐接口
- **优先级**：P1
- **依赖**：任务 1
- **描述**：检查是否有推荐服务或相关推荐接口
- **成功标准**：相关推荐接口存在
- **测试要求**：
  - `programmatic`：相关推荐接口存在

### [ ] 任务 5：检查 Subtitle 服务字幕接口
- **优先级**：P1
- **依赖**：任务 1
- **描述**：检查 Subtitle 服务是否有 `/subtitle/video/{id}` 接口
- **成功标准**：Subtitle 服务提供正确的字幕接口
- **测试要求**：
  - `programmatic`：Subtitle 服务控制器存在对应接口

### [ ] 任务 6：修复路由配置
- **优先级**：P0
- **依赖**：任务 1-5
- **描述**：根据检查结果修复 Gateway 路由配置
- **成功标准**：所有 404 接口路由配置正确
- **测试要求**：
  - `programmatic`：Gateway 路由配置正确

### [ ] 任务 7：验证修复结果
- **优先级**：P0
- **依赖**：任务 6
- **描述**：测试所有 404 接口是否恢复正常
- **成功标准**：所有接口返回 200 或正确的业务响应
- **测试要求**：
  - `programmatic`：所有接口返回 200 状态码
