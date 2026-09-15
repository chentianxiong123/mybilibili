# 状态机对照表（旧版常量 → Go 版）

> 依据旧版 `mybilibili-cloud` 实体类/枚举逐条提取。规则：**旧版状态常量新版必须同名同值，不得改值**（文档 3.2 强制规则 #1）。
> 核对日期：2026-08-14

## 1. Manuscript（稿件）

来源：`mybilibili-common/src/main/java/com/mybilibili/common/entity/Manuscript.java:49-58`

| 常量 | 值 | 语义 | Go 侧 |
|---|---|---|---|
| `STATUS_PENDING_REVIEW` | 0 | 待审核 | `internal/admin/admin_manuscript_handler.go:168` ✅ |
| `STATUS_PROCESSING` | 1 | 处理中 | `:169` ✅ |
| `STATUS_PUBLISHED` | 3 | 已发布 | `:170` ✅ |
| `STATUS_REJECTED` | 4 | 已拒绝 | `:171` ✅ |
| `STATUS_PROCESS_FAILED` | 5 | 处理失败 | `:172` ✅ |
| `STATUS_UNPUBLISHED` | -1 | 已下架 | `:173` ✅ |
| `REVIEW_STATUS_PENDING` | 0 | 审核待处理 | `:175` ✅ |
| `REVIEW_STATUS_APPROVED` | 1 | 审核通过 | `:176` ✅ |
| `REVIEW_STATUS_REJECTED` | 2 | 审核拒绝 | `:177` ✅ |

## 2. Video（视频处理进度）

来源：`mybilibili-common/src/main/java/com/mybilibili/common/entity/Video.java:32-56`

| 常量 | 值 | Go 侧 |
|---|---|---|
| `STATUS_PENDING_REVIEW` | 0 | —（视频经稿件状态控制） |
| `STATUS_PROCESSING` | 1 | — |
| `STATUS_PUBLISHED` | 3 | — |
| `STATUS_REJECTED` | 4 | — |
| `STATUS_PROCESS_FAILED` | 5 | — |
| `STATUS_UNPUBLISHED` | -1 | — |
| `REVIEW_STATUS_PENDING/APPROVED/REJECTED` | 0/1/2 | — |
| `PROCESS_STATUS_PENDING` | 0 | `admin/admin_manuscript_handler.go:179` ✅ |
| `PROCESS_STATUS_TRANSCODING` | 1 | `:180` ✅ |
| `PROCESS_STATUS_TRANSCODE_FAILED` | 10 | `:185` ✅（`TranscodeEnd=11`，10 未命名常量） |
| `PROCESS_STATUS_TRANSCODE_SUCCESS` | 11 | `:185` ✅ |
| `PROCESS_STATUS_AUDIO_EXTRACTING` | 2 | `:181` ✅ |
| `PROCESS_STATUS_AUDIO_FAILED` | 20 | 未定义常量（视频失败统一记 process_error） |
| `PROCESS_STATUS_AUDIO_SUCCESS` | 21 | 未定义常量 |
| `PROCESS_STATUS_SUBTITLE_GENERATING` | 3 | `:182` ✅ |
| `PROCESS_STATUS_SUBTITLE_FAILED` | 30 | 未定义常量 |
| `PROCESS_STATUS_SUBTITLE_SUCCESS` | 31 | 未定义常量 |
| `PROCESS_STATUS_AI_SUMMARIZING` | 4 | `:183` ✅ |
| `PROCESS_STATUS_AI_FAILED` | 40 | 未定义常量 |
| `PROCESS_STATUS_AI_SUCCESS` | 41 | 未定义常量 |
| `PROCESS_STATUS_COMPLETED` | 5 | `:184` ✅ |

> ⚠️ Go 侧常量定义于 `internal/admin/admin_manuscript_handler.go:179-185`。**失败态（10/20/30/40）与 SUCCESS 态（21/31/41）无命名常量、无写路径**；仅 `triggerVideoProcess`/`resetVideo`/`triggerAllVideoProcess` 写 PROCESSING/TRANSCODING/PENDING，完成态无自动推进。

## 3. ManuscriptEditVersion（编辑版本审核结果）

来源：`mybilibili-video-media/src/main/java/com/mybilibili/video/entity/ManuscriptEditVersion.java:11-13`

| 常量 | 值 |
|---|---|
| `STATUS_PENDING` | `"PENDING"` |
| `STATUS_APPROVED` | `"APPROVED"` |
| `STATUS_REJECTED` | `"REJECTED"` |

> ❌ Go 侧**无任何 `manuscript_edit_versions` 写入代码**（表在 `sql/011_analytics.sql:1` 已建）。

## 4. ManuscriptCollection（合集）

来源：`mybilibili-common/src/main/java/com/mybilibili/common/entity/ManuscriptCollection.java:31-32`

| 常量 | 值 |
|---|---|
| `STATUS_PRIVATE` | 0（私密） |
| `STATUS_PUBLIC` | 1（公开） |

> ✅ Go `CollectionService.Create` 支持 status 入参。

## 5. InteractionType（用户互动类型）

来源：`mybilibili-common/src/main/java/com/mybilibili/common/enums/InteractionType.java:4-8`

| 枚举 | code |
|---|---|
| `LIKE` | `"LIKE"` |
| `COLLECT` | `"COLLECT"` |
| `COIN` | `"COIN"` |
| `SHARE` | `"SHARE"` |

> ✅ Go `interaction_repository.go` 使用同字符串。旧版另有 `TargetType`（`MANUSCRIPT`），Go 用 `"MANUSCRIPT"` 一致。

## 6. ManuscriptAnalyticsEvent（分析事件）

来源：`ManuscriptAnalyticsEvent.java`（搜索/推荐模块）

| 常量 | 值 |
|---|---|
| `TYPE_VIEW_INCREMENT` | `"VIEW_INCREMENT"` |
| `TYPE_STATUS_CHANGE` | `"STATUS_CHANGE"` |
| `TYPE_METRIC_INCREMENT` | `"METRIC_INCREMENT"` |
| `METRIC_LIKE` | `"LIKE"` |
| `METRIC_COIN` | `"COIN"` |
| `METRIC_COLLECT` | `"COLLECT"` |
| `METRIC_SHARE` | `"SHARE"` |
| `METRIC_COMMENT` | `"COMMENT"` |
| `METRIC_DANMAKU` | `"DANMAKU"` |

> ❌ Go `internal/core/event_publisher.go:30 PublishAnalytics` 已定义但**无任何调用方**（互动/评论/稿件都不发）。

## 7. 状态流转动作（Manuscript 状态事件 action 值）

来源：`ManuscriptServiceImpl.java` 内 `recordManuscriptStatusEvent` 调用点

| action | from→to | 触发方 |
|---|---|---|
| `UPLOAD_SUBMITTED` | null→0(pending) | 用户投稿 |
| `OWNER_EDIT_SUBMITTED` | 任意→0(pending) | 用户编辑重提 |
| `ADMIN_APPROVE` | →1(processing) | 管理员审核通过 |
| `ADMIN_APPROVE_WITH_PROCESS` | →1(processing) | 管理员通过+处理 |
| `ADMIN_REJECT` | →4(rejected) | 管理员拒绝 |
| `ADMIN_PUBLISH` | →3(published) | 管理员发布 |
| `ADMIN_UNPUBLISH` | →-1(unpublished) | 管理员下架 |
| `OWNER_REPUBLISH` | -1→3(published) | 用户重新上架 |
| `OWNER_UNPUBLISH` | →-1(unpublished) | 用户下架 |
| `TAKE_DOWN_VIOLATION` | →-1(review=2) | 违规下架 |
| `ADMIN_RETRY` | →1(processing) | 管理员重试 |
| `AUTO_PUBLISH` | →3(published) | 系统（视频处理完成自动上架） |

> ❌ Go 侧**全部缺失**：无 `manuscript_status_events` 写入（`sql/011_analytics.sql:19` 表已建）。

## 8. Message（消息）

> ⚠️ 旧版常量内联于 service/controller，未集中实体。核对到的字段：`messageType`(int)、`isRead`(bool)。Go `message_handler.go` 需逐方法核对类型拆分。