# 管理后台核心功能实现 Spec

## Why
管理后台的稿件管理、视频处理、索引管理、内容审核中心四个核心页面需要完整的前后端实现，目前部分功能缺少后端接口或实现不完整。

## What Changes
- 实现索引管理后端接口（IndexAdminController）
- 验证并完善稿件管理、视频处理、内容审核的后端服务实现
- 确保网关路由正确配置

## Impact
- Affected specs: mybilibili-search, mybilibili-video, mybilibili-comment
- Affected code: 
  - mybilibili-search: 新增 IndexAdminController
  - mybilibili-video: ManuscriptServiceImpl 需要完善
  - mybilibili-comment: ContentReviewController 需要验证

## Current Status Analysis

### 1. 稿件管理
| 层级 | 状态 | 文件 |
|------|------|------|
| 前端 | ✅ 已实现 | ManuscriptsView.vue |
| 后端 Controller | ✅ 已实现 | ManuscriptAdminController.java |
| 后端 Service | ⚠️ 需验证 | ManuscriptServiceImpl.java |
| 网关路由 | ✅ 已配置 | /api/manuscript/admin/** |

### 2. 视频处理
| 层级 | 状态 | 文件 |
|------|------|------|
| 前端 | ✅ 已实现 | VideoProcessView.vue |
| 后端 Controller | ✅ 已实现 | ManuscriptAdminController.java |
| 后端 Service | ⚠️ 需验证 | ManuscriptServiceImpl.java |
| 网关路由 | ✅ 已配置 | /api/manuscript/admin/** |

### 3. 索引管理
| 层级 | 状态 | 文件 |
|------|------|------|
| 前端 | ✅ 已实现 | IndexManagerView.vue |
| 后端 Controller | ❌ 缺失 | 需创建 IndexAdminController.java |
| 后端 Service | ⚠️ 需验证 | VideoSearchService.java |
| 网关路由 | ✅ 已配置 | /api/search/** |

### 4. 内容审核中心
| 层级 | 状态 | 文件 |
|------|------|------|
| 前端 | ✅ 已实现 | ContentReviewView.vue |
| 后端 Controller | ✅ 已实现 | ContentReviewController.java |
| 后端 Service | ⚠️ 需验证 | CommentServiceImpl.java |
| 网关路由 | ✅ 已配置 | /api/admin/content-review/** |

## ADDED Requirements

### Requirement: 索引管理接口
系统应提供 Elasticsearch 索引管理接口，支持管理员查看索引状态、批量索引稿件、重建索引等操作。

#### Scenario: 获取索引状态
- **WHEN** 管理员访问 /api/search/admin/index/status
- **THEN** 返回当前索引的文档数量、索引名称、健康状态

#### Scenario: 批量索引稿件
- **WHEN** 管理员调用 /api/search/admin/index/bulk
- **THEN** 系统将所有已上架稿件批量索引到 Elasticsearch

#### Scenario: 重建索引
- **WHEN** 管理员调用 /api/search/admin/index/rebuild
- **THEN** 系统删除旧索引并重新创建，然后批量导入数据

### Requirement: 稿件管理服务完善
稿件管理服务需要完整实现以下方法：
- getPendingManuscripts(): 获取待审核稿件
- getProcessingManuscripts(): 获取处理中稿件
- getReadyManuscripts(): 获取待上架稿件
- getAllManuscripts(): 获取所有稿件
- approveManuscript(): 审核通过
- rejectManuscript(): 审核拒绝
- publishManuscript(): 上架稿件
- unpublishManuscript(): 下架稿件
- getManuscriptStatistics(): 获取统计数据

### Requirement: 视频处理服务完善
视频处理服务需要完整实现以下方法：
- manualTranscode(): 手动触发转码
- manualExtractAudio(): 手动提取音频
- manualGenerateSubtitle(): 手动生成字幕
- manualAiSummary(): 手动AI总结
- manualProcessAll(): 一键全流程处理
- resetVideoStatus(): 重置视频状态
- getVideoProcessStatus(): 获取处理状态

### Requirement: 内容审核服务完善
内容审核服务需要完整实现以下方法：
- getPendingList(): 获取待审核内容列表
- getAllContent(): 获取所有内容
- restoreContent(): 恢复内容
- deleteContent(): 删除内容
- batchProcess(): 批量处理
