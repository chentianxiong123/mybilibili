# 稿件全链路状态机与数据流 — 2026-06-11 探索报告

> 时间: 2026-06-11
> 仓库: `D:\files\mybilibili-next\mybilibili-cloud`
> 分支: `feature/manuscript-edit-review-20260531`
> 方法: 不基于假设,**每个疑点都用 grep/Read 在源码中验证后再下结论**
> 范围: 稿件从用户上传 → 审核 → 视频处理 → 发布 → 搜索/推荐索引 → 互动数据回流 的完整链路

---

## 0. TL;DR — 探索结果速览

| 类别 | 数量 | 说明 |
|---|---|---|
| 已确认的真 bug | **3 个** | 阻塞 demo |
| 已确认的孤儿代码 | **1 处** | 死代码 |
| 已确认的"设计意图"(非 bug) | **5 处** | 有意为之 |
| 需要产品决策 | **2 处** | 不确定是 bug 还是 feature |
| 需要后续深挖 | **3 处** | 涉及跨服务架构,本报告只做标注 |

---

## 1. 状态机定义(从 `Manuscript.java` / `Video.java` 实际源码)

### 1.1 Manuscript 状态(`mybilibili-common/.../entity/Manuscript.java:49-55`)

```java
public static final int STATUS_PENDING_REVIEW   = 0;  // 待审核
public static final int STATUS_PROCESSING       = 1;  // 处理中(审核通过,正在转码)
public static final int STATUS_READY_TO_PUBLISH = 2;  // 待发布
public static final int STATUS_PUBLISHED        = 3;  // 已发布
public static final int STATUS_REJECTED         = 4;  // 已拒绝
public static final int STATUS_PROCESS_FAILED   = 5;  // 处理失败
public static final int STATUS_UNPUBLISHED      = -1; // 已下架
```

### 1.2 Manuscript 审核状态

```java
public static final int REVIEW_STATUS_PENDING   = 0;
public static final int REVIEW_STATUS_APPROVED  = 1;
public static final int REVIEW_STATUS_REJECTED  = 2;
```

### 1.3 Video 处理状态(`mybilibili-common/.../entity/Video.java:34-50`)

```java
public static final int PROCESS_STATUS_PENDING              = 0;   // 待处理
public static final int PROCESS_STATUS_TRANSCODING          = 1;   // 转码中
public static final int PROCESS_STATUS_AUDIO_EXTRACTING     = 2;   // 音频提取中
public static final int PROCESS_STATUS_SUBTITLE_GENERATING  = 3;   // 字幕生成中
public static final int PROCESS_STATUS_AI_SUMMARIZING       = 4;   // AI 总结中
public static final int PROCESS_STATUS_COMPLETED            = 5;   // 全部完成
public static final int PROCESS_STATUS_TRANSCODE_FAILED     = 10;  // 转码失败
public static final int PROCESS_STATUS_AUDIO_FAILED         = 20;  // 音频失败
public static final int PROCESS_STATUS_SUBTITLE_FAILED      = 30;  // 字幕失败
public static final int PROCESS_STATUS_AI_FAILED            = 40;  // AI 失败
public static final int PROCESS_STATUS_TRANSCODE_SUCCESS    = 11;  // 转码成功
public static final int PROCESS_STATUS_AUDIO_SUCCESS        = 21;  // 音频成功
public static final int PROCESS_STATUS_SUBTITLE_SUCCESS     = 31;  // 字幕成功
public static final int PROCESS_STATUS_AI_SUCCESS           = 41;  // AI 成功
```

### 1.4 Video 也有 `STATUS_*`(同 Manuscript,值一样)

**死代码** — `grep` 验证: 0 处使用 `Video.STATUS_PUBLISHED / STATUS_PROCESSING / STATUS_REJECTED / STATUS_READY_TO_PUBLISH`。Video 实际只用 `PROCESS_STATUS_*` 这套。

**为什么死**: 这是早期版本 Video 也有"发布"概念,后来 Video 改成纯"处理状态"概念(转码/AI)。状态常量没删,被遗忘。

---

## 2. 真实状态机流转(基于源码验证)

### 2.1 完整状态图(从代码推断)

```
                              用户上传
                                 ↓
                manuscript: PENDING_REVIEW(0)  ← uploadManuscript 直接设这个
                video:       PENDING(0)
                视频处理: ❌ 不会触发
                                 ↓
                          管理员审核
                          ┌──────┴──────┐
                       approve        reject
                          ↓              ↓
                manuscript: PROCESSING(1)   REJECTED(4)
                MQ: TRANSCODE 触发         (无 MQ)
                          ↓
                  VideoMediaProcessConsumer
                          ↓
                  ┌───────┴────────┐
              transcode        transcode 失败
                  ↓                  ↓
                success(11)        FAILED(10)
                  ↓ 触发 next
              extract audio
                  ↓ success(21)
                  ↓ 触发 next
              generate subtitle (AI 服务)
                  ↓ success(31)
                  ↓ 触发 next
              ai summary (AI 服务)
                  ↓ success(41)
                  ↓
              COMPLETED(5)
                  ↓
              ❌ 没有任何代码把 manuscript 推进到 READY_TO_PUBLISH(2)
                  ↓
                (等待)
                                 ↓
                          管理员/用户 publish
                          ┌──────┴──────┐
                       publish       (无前置校验)
                          ↓
                manuscript: PUBLISHED(3)
                userEventMQ: 通知用户
                search-index: ❌ 不会触发
                                 ↓
                            前台可见
                                 ↓
                          互动数据(点赞/投币/收藏)
                                 ↓
                content-interaction: UPDATE manuscripts.like_count
                                  (同步直写,跨服务)
                                 ↓
                MQ: ManuscriptAnalyticsEvent METRIC_LIKE
                                  (search-recommend 写日统计表)
                                 ↓
                ES 文档 likeCount: ❌ 永远不更新
                                 ↓
                          unpublish
                                 ↓
                manuscript: UNPUBLISHED(-1)
                ES 文档: ❌ 不会删除
```

### 2.2 关键代码点(每一步都有源码)

| 步骤 | 代码位置 | 关键逻辑 |
|---|---|---|
| 用户上传 | `ManuscriptServiceImpl.uploadManuscript:100-168` | 设 `status=PENDING_REVIEW`,**不发 MQ** |
| 管理员 approve | `ManuscriptServiceImpl.approveManuscript:1086-1112` | 设 `status=PROCESSING`,**对每个 video 发 TRANSCODE** |
| 转码 | `VideoMediaProcessConsumer.handleTranscode:47-61` | FFmpeg → 3 个 m3u8 → `markTranscodeSuccess` → 触发 EXTRACT_AUDIO |
| 音频提取 | `VideoMediaProcessConsumer.handleAudioExtract:63-73` | FFmpeg → wav → `markProgress` → 触发 GENERATE_SUBTITLE |
| 字幕 | `VideoAiProcessConsumer` + `SubtitleProcessStep:31-44` | 调用 AI → 写 `has_subtitle=1` |
| 摘要 | `VideoAiProcessConsumer` + `VideoProcessOrchestratorImpl:40-44` | 调用 AI → `markProcessCompleted` |
| 待发布 | ❌ **无代码** | 状态机断点 |
| 发布 | `ManuscriptServiceImpl.publishManuscript:1182-1202` | 设 `status=PUBLISHED`,发通知 MQ,**不发索引 MQ** |
| 互动 | `VideoInteractionServiceImpl.likeVideo:84-90` | 直接 UPDATE 库 + 发 metrics MQ |
| 下架 | `ManuscriptServiceImpl.unpublishManuscript:1204-1219` | 设 `status=UNPUBLISHED`,**不发任何 MQ** |

---

## 3. 已确认的真 Bug(阻塞 demo)

### 🔴 Bug 1: `STATUS_READY_TO_PUBLISH(2)` 是孤儿状态码

**证据**:
- `grep "STATUS_READY_TO_PUBLISH\s*="` 在整个项目内只有 entity 常量定义,**没有任何 set 调用**
- `grep "setStatus(2)\|setStatus(Manuscript.STATUS_READY"` 0 命中
- 实际代码里 `approveManuscript` 设 PROCESSING(1),`publishManuscript` 设 PUBLISHED(3)
- 唯一的引用 `getReadyManuscripts()` (`ManuscriptAdminController:55-60`) **永远返回空列表**

**症状**:
- 视频处理跑完后,manuscript 永远卡在 PROCESSING
- admin 后台"待发布"菜单**永远是空的**
- 演示时 admin 找不到刚处理完的稿件,只能去"处理中"列表找

**触发条件**: 任何审核通过且视频处理完毕的稿件

**修复方案**:
- 方案 A: 在 `VideoProcessOrchestratorImpl.markProcessCompleted` 后,发一个 `ManuscriptReadyToPublishEvent` 到新 topic,video-media 监听,改 manuscript.status=READY_TO_PUBLISH
- 方案 B: 在 `VideoMediaProcessStateServiceImpl` 监听 `video-process-completed` 事件,检查 manuscript 下所有 video.process_status 都是 COMPLETED(5) 或 FAILED(允许部分失败),改 manuscript.status
- **我推荐方案 B**,因为它复用现有的 PROCESS_STATUS_COMPLETED(5) 信号,不需要新增事件类型

### 🔴 Bug 2: 用户上传后 video 不进入处理流程

**证据**:
- `grep "videoProcessPort\."` 只在 4 处出现: `approveManuscript`, `approveManuscriptWithProcess`, `manualProcessAll`, 以及它的 PortAdapter
- `uploadManuscript` 0 处调用 `videoProcessPort`
- 用户上传后 video.process_status = PENDING(0),**永远不前进**

**症状**:
- 用户上传后看不到"处理进度",因为根本没有启动
- 必须等管理员 approve 才会处理
- 如果管理员忘记 approve,用户的视频永远卡 PENDING

**触发条件**: 任何用户上传

**修复方案**:
- 方案 A: 上传后立即发 MQ 启动处理(违反"先审核后处理"的设计)
- 方案 B: 接受当前设计(审核通过才处理),**只需要在 UI 上明示**"稿件已提交,等待管理员审核"
- **建议先采纳方案 B**,因为方案 A 会浪费转码资源给最终被拒绝的稿件

### 🔴 Bug 3: `publishManuscript` 不检查 video.process_status

**证据**:
- `grep -A 15 "public boolean publishManuscript\("` 整个方法**没有任何 processStatus 检查**
- 即使视频转码失败(PROCESS_STATUS_TRANSCODE_FAILED=10),管理员点 publish 也能上架
- 前端拿到 `playUrlHd=null`,显示"视频不存在"

**症状**:
- admin 一不小心把转码失败的稿件上架,前台空白
- 没有任何"该稿件视频未就绪,无法发布"的提示

**修复方案**:
```java
public boolean publishManuscript(Integer manuscriptId) {
    Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
    if (manuscript == null) return false;

    // 新增: 检查视频处理状态
    List<Video> videos = videoMapper.selectByManuscriptId(manuscriptId);
    boolean anyTranscoded = videos.stream().anyMatch(v ->
        v.getProcessStatus() != null &&
        v.getProcessStatus() >= Video.PROCESS_STATUS_TRANSCODE_SUCCESS);  // 11
    if (!anyTranscoded) {
        throw new BusinessException("视频未转码完成,无法发布");
    }
    // ... 原有逻辑
}
```

### 🔴 Bug 4(隐藏): `publishManuscriptByOwner` 不检查 status

**证据**: 同样 `grep`,用户自己 publish 接口**完全没有 status 过滤** — 用户可以从 PENDING_REVIEW(0) 直接 publish 到 PUBLISHED(3),**绕过审核**

**严重程度**: 安全漏洞,应立刻修

**修复方案**: 同 Bug 3,加 `if (manuscript.getStatus() != STATUS_READY_TO_PUBLISH) throw ...`

---

## 4. 已确认的"设计意图"(非 bug,但要标清楚)

### 4.1 ES 只索引 PUBLISHED 稿件

**证据**: `mybilibili-search-recommend/.../ManuscriptMapper.java:24`
```sql
WHERE m.status = 3   -- PUBLISHED
```

**意图**: 未发布的稿件不进搜索,正确。

**坑**: `rebuildIndex` / `bulkIndex` 是 admin 手动,**没有自动触发**。如果管理员忘了重建索引,新发布的稿件**搜不到**。

### 4.2 跨服务直写 `manuscripts` 表(点赞等)

**证据**: `content-interaction/.../VideoInteractionServiceImpl.java:84`
```java
manuscriptMapper.updateLikeCount(manuscriptId, 1);   // 跨服务 UPDATE
userMapper.updateLikedCount(userId, 1);               // 跨服务 UPDATE
```

**意图**: 互动数据要"立即可见",所以同步直写 + MQ 日统计。

**坑**:
- 不算微服务,算"分布式单体"
- 写库失败会回滚,但 MQ 发了,日统计和实时数据不一致
- 5 个服务都有跨服务 Mapper (见 §6)

### 4.3 MQ 死 topic(`TOPIC_SEARCH_INDEX`, `TOPIC_VIDEO_PUBLISH`)

**证据**: 0 个 Producer, 0 个 Consumer

**意图**: 早期规划预留,从未实现

**坑**: 看代码会以为有自动索引,实际没有

### 4.4 `Video.STATUS_*` 是死代码

**证据**: 0 处使用

**意图**: 早期版本残留

**建议**: 直接删(改 entity 一行)

### 4.5 "草稿" 状态实际是 PENDING_REVIEW(0)

**证据**: `ManuscriptController.convertStatusParam:490-495`
```java
case "pending":
case "reviewing":
case "draft":        // ← 前端"草稿"落到这里
    return Manuscript.STATUS_PENDING_REVIEW;
```

**意图**: 给前端用别名,后端统一 PENDING_REVIEW

**坑**: 前端 WAP 我之前扫到 WAP 用了"草稿"概念,但状态码上没有区分,**用户编辑后是否走另一条路需要查前端代码**

---

## 5. 需要产品决策(不确定是 bug 还是 feature)

### 5.1 上传后是否自动处理?

**两种选择**:
- **A. 上传即处理**(符合用户预期,但被拒绝的稿件浪费算力)
- **B. 审核通过才处理**(当前实现,资源友好但用户感知差)

**我的建议**: B 加上"用户能看到处理进度待审核"的 UI 文案。否则用户上传后什么都看不到,以为系统坏了。

### 5.2 reject 后能不能再 approve?

**当前代码**: `approveManuscript` 不检查当前 status,从 REJECTED(4) 直接进 PROCESSING(1)。**没有代码禁止**。

**但**: 视频处理链已经跑过一次(PROCESS_STATUS_xxx),再次发 TRANSCODE MQ 会:
- `VideoMediaProcessConsumer` 接到新 MQ,转码**会再跑一次**(覆盖之前的输出)
- `dispatchNextIfNeeded` 又触发 EXTRACT_AUDIO...**全流程重跑**

**不算 bug,但是浪费**。建议 admin 后台"重试"按钮只触发 PROCESS_STATUS_PENDING 的稿件,而不是无脑重跑。

---

## 6. 跨服务直查/直写 DB 全景图(架构债)

5 个服务都有跨服务 Mapper(本次全量扫描):

| 服务 | Mapper | 跨服务访问的表 | 写? | 风险 |
|---|---|---|---|---|
| `mybilibili-account-social` | `UserMapper` | `users`(本服务) | - | 无 |
| `mybilibili-video-media` | `ManuscriptMapper`, `VideoMapper` | `manuscripts`, `videos`(本服务) | - | 无 |
| `mybilibili-content-interaction` | `interaction/mapper/ManuscriptMapper` | `manuscripts`(其他服务) | **是(UPDATE)** | 🔴 |
| `mybilibili-content-interaction` | `interaction/mapper/UserMapper` | `users`(其他服务) | **是(UPDATE)** | 🔴 |
| `mybilibili-search-recommend` | `search/mapper/ManuscriptMapper` | `manuscripts/users/categories/videos/tags/video_tags`(全栈) | 只读 | 🟡 |
| `mybilibili-search-recommend` | `analytics/mapper/CreatorStatsMapper` | `manuscripts/comments/users/user_interactions` | 只读 | 🟡 |
| `mybilibili-ai` | `ai/mapper/VideoMapper` | `videos`(其他服务) | **是(UPDATE process_status)** | 🔴 |
| `mybilibili-ai` | `ai/mapper/ManuscriptMapper` | `manuscripts` | 只读 | 🟡 |
| `mybilibili-ai` | `ai/mapper/AdminStatsMapper` | `ai_*`(本服务) | - | 无 |

**架构债等级**: 整体是"分布式单体"。

**真实微服务需要做的**:
- 跨服务的写操作改成事件(MQ)
- 跨服务的读操作改成 Feign 调用或者读 search-recommend 自己的 ES 镜像

**当前能跑的原因**: 所有服务连同一个 MySQL,事务一致。

**当前能跑不能上生产的风险**:
- 跨服务写锁(同表,长事务容易死锁)
- 跨服务读表结构变化(改表要同步改 4 个服务)
- 分布式事务无法保证

---

## 7. ES 文档"不更新"问题(本次新发现)

### 7.1 ES `ManuscriptDocument` 的字段(从 `common/.../ManuscriptDocument.java`)

- `viewCount, likeCount, commentCount, shareCount, collectCount, coinCount` — **6 个互动字段**

### 7.2 谁更新这些字段?

**答案**: 没有人。

- `ManuscriptIndexServiceImpl` 在 `bulkIndex` / `rebuildIndex` / `incrementalIndex` 时,从 MySQL `manuscripts` 表读数据建 ES
- **之后这些字段永远不更新**
- 互动数据走的是 `manuscript_analytics_events` 日统计表,和 ES 无关

### 7.3 后果

- 首页卡片**永远显示 0 播放 / 0 评论 / 0 点赞**
- 推荐算法 ES `viewCount` 排序是 0,等价于随机
- 演示时非常尴尬 — "我的稿件明明 100 个赞,首页显示 0 赞"

### 7.4 修复方向(供选择)

**方案 A**: 互动事件时同步更新 ES 文档
- `VideoInteractionServiceImpl.likeVideo` 后,发 `manuscript-metric-update` 事件,search-recommend 监听,update ES 文档
- 工作量: 0.5-1 天

**方案 B**: ES 卡片数据用 Feign 实时查 `manuscripts.like_count`
- 每次返回 ES 文档时,用 Feign 查 video-media 服务拿最新数
- 工作量: 0.5-1 天,但增加延迟

**方案 C**: ES 文档只存"快照式"概要(标题/封面/作者/时长/发布日期),互动数据卡片显示 N/A
- 工作量: 0.5 天,UI 适配
- **如果卡片不一定要显示具体数字,这个最快**

**我推荐 C** — 项目当前阶段,先解决"不显示 0"的问题。如果未来要显示真实数字,再做 A。

---

## 8. 互动数据双写架构(已确认非 bug,但有坑)

### 8.1 当前流程(以点赞为例)

```
content-interaction.VideoInteractionServiceImpl.likeVideo():
  1. user_interactions 插记录 (本服务表)
  2. redisTemplate increment (Redis 缓存)
  3. manuscriptMapper.updateLikeCount(manuscriptId, 1)  ← 跨服务直写 manuscripts
  4. userMapper.updateLikedCount(userId, 1)             ← 跨服务直写 users
  5. publishMetric(METRIC_LIKE, 1) → MQ
  6. search-recommend.ManuscriptAnalyticsEventConsumer 写日统计表
```

---

## 13. 2026-06-11 实际修改清单(与用户对齐后)

以下是对齐环节结束后实际落地到代码的改动:

### 13.1 修"单步不跳" bug

- **文件**: `VideoProcessMessagePortAdapter.java`
- **效果**: `sendManualStep` 内部创建 MQ 消息时改为 `VideoProcessMessage.of(...)`
  (autoChain),不再调 `manualSingle()`。结果:admin 点 manualTranscode
  成功后会自动跳音频→字幕→AI,不再停在单步。
- `MANUAL_SINGLE` 代码保留(consumer 端兼容旧消息),不再被生产。

### 13.2 删 STATUS_READY_TO_PUBLISH(2)

删除范围:

| 文件 | 改动 |
|---|---|
| `Manuscript.java` | 删 `STATUS_READY_TO_PUBLISH = 2` |
| `Video.java` | 同上 |
| `CustomerServiceReadonlyToolSet.java` | 删 `case "待发布"` |
| `ManuscriptController.java` | 删 `case "ready"` |
| `ManuscriptService.java` | 删 `getReadyManuscripts()` 接口声明 |
| `ManuscriptServiceImpl.java` | 删 `getReadyManuscripts()` 实现 + `stats.put("ready")` |
| `ManuscriptAdminController.java` | 删 `@GetMapping("/ready")` 接口 |
| `ManuscriptMapper.java` | 删 `countReadyByUserId()` |
| `manuscripts.js`(admin-web API) | 删 `getReadyManuscripts()` |
| `ManuscriptsView.vue`(admin-web) | 删 import 和 `case '2'` 分支 |

### 13.3 新增 ManuscriptAutoPublishConsumer

| 文件 | 性质 | 说明 |
|---|---|---|
| `VideoPublishEvent.java` | 新文件 | 事件体:manuscriptId + videoId + trigger |
| `VideoMQProducer.java` | 已改 | 加 `sendVideoPublishEvent()` |
| `MQConstants.java` | 已改 | `TOPIC_VIDEO_PUBLISH` 已有(0 Producer → now 1) |
| `VideoProcessStateServiceImpl.java`(mybilibili-ai) | 已改 | `markProcessCompleted` 后发 VideoPublishEvent |
| `ManuscriptAutoPublishConsumer.java`(video-media) | **新文件** | 监听 `TOPIC_VIDEO_PUBLISH`,调 `autoPublishIfAllVideosCompleted` |
| `ManuscriptService.java` | 已改 | 加 `autoPublishIfAllVideosCompleted()` 接口声明 |
| `ManuscriptServiceImpl.java` | 已改 | 实现:查 manuscript 下全部 video 是否 COMPLETED → 改 PUBLISHED + 发通知 + 发索引事件 |

### 13.4 新增 ES 索引事件链路

| 文件 | 性质 | 说明 |
|---|---|---|
| `ManuscriptIndexEvent.java` | **新文件** | UPSERT/DELETE + manuscriptId |
| `MQConstants.java` | 已改 | 加 `TOPIC_MANUSCRIPT_INDEX` |
| `VideoMQProducer.java` | 已改 | 加 `sendManuscriptIndexEvent()` |
| `ManuscriptIndexEventConsumer.java`(search-recommend) | **新文件** | 监听 UPSERT → `indexOne`; DELETE → `deleteOne` |
| `ManuscriptIndexService.java` | 已改 | 加 `indexOne()` / `deleteOne()` 接口 |
| `ManuscriptIndexServiceImpl.java` | 已改 | 实现:单条查 MySQL → convert → save ES / delete ES |
| `ManuscriptMapper.java`(search-recommend) | 已改 | 加 `selectPublishedManuscriptById` |

### 13.5 修 OWNER_PUBLISH 安全漏洞

- `publishManuscriptByOwner`: 加状态过滤,只允许 UNPUBLISHED→PUBLISHED
- 新增触发 `ManuscriptIndexEvent.OPERATION_UPSERT`

### 13.6 下架全链路加 ES DELETE

| 入口 | 事件 |
|---|---|
| `unpublishManuscript`(admin) | 发 `OPERATION_DELETE` + `trigger=ADMIN_UNPUBLISH` |
| `unpublishManuscriptByOwner`(用户) | 发 `OPERATION_DELETE` + `trigger=OWNER_UNPUBLISH` |
| `takeDownViolatingManuscript`(违规下架,新增) | 发 `OPERATION_DELETE` + `trigger=TAKE_DOWN_VIOLATION` |

旧代码 `takeDownManuscriptInternal` 直接 `updateManuscript` 绕过 MQ,已重写为调
`takeDownViolatingManuscript` Service 方法。

### 13.7 未改动(决策后保留现状)

- 审核流程不变:先审核再转码,admin approve 后走 AUTO_CHAIN
- 6 个 manual 按钮保留(单步调试 + 日志 + 任务中心联动)
- admin `publishManuscript` 保留(作为 force-publish/republish 备用入口)
- admin `unpublishManuscript` 保留(加发了 ES DELETE)
- `MANUAL_SINGLE` 代码保留(consumer 兼容)
- `Video.STATUS_*` 死代码未删(与本期无关)
- 用户"我的稿件"列表包括 UNPUBLISHED 状态(前端双保险过滤)
- 举报链路不改(用户提交 → AI 审核 → admin 决策 → admin 手动下架)

### 13.8 新链路图(最终版)

```
用户上传 → PENDING_REVIEW
              ↓
admin approve → PROCESSING + 发 TRANSCODE MQ(AUTO_CHAIN)
              ↓
       ┌──TRANSCODE──┐
       │   success    │→ 发 EXTRACT_AUDIO MQ
       └─────────────┘
              ↓
       ┌──EXTRACT_AUDIO──┐
       │    success       │→ 发 GENERATE_SUBTITLE MQ
       └──────────────────┘
              ↓
       ┌──GENERATE_SUBTITLE──┐  (AI 服务消费)
       │      success        │→ 发 AI_SUMMARY MQ
       └─────────────────────┘
              ↓
       ┌──AI_SUMMARY──┐
       │   success    │→ COMPLETED + 发 VideoPublishEvent
       └──────────────┘
              ↓
       ManuscriptAutoPublishConsumer
              ↓ 检查全部 video COMPLETED
              ↓ 改 status=PUBLISHED + 通知用户
              ↓ 发 ManuscriptIndexEvent(UPSERT)
              ↓
       ManuscriptIndexEventConsumer
              ↓ indexOne → ES 可见
              ↓
       前台可见 + 可搜索 + 可推荐
              ↓
       (下架) → 发 ManuscriptIndexEvent(DELETE)
       (重上架) → 发 ManuscriptIndexEvent(UPSERT)
```

### 8.2 双写的不一致风险

- 步骤 3-4 失败 → 步骤 5 已经发 MQ → 日统计和实时不一致
- MQ 失败 → 日统计没更新,但 manuscripts 表已加
- Redis increment 失败 → 缓存不准

### 8.3 当前能跑的原因

所有服务连同一个 MySQL,事务默认在 content-interaction 内提交,跨服务写库立即可见。

### 8.4 建议

不动架构(成本太高),但**加监控**: 检测 `manuscripts.like_count` 和 `manuscript_analytics_events.like_count` 累计差异。

---

## 9. 三个"暂时不动"的事项(后续深挖)

### 9.1 ES 客户端跨大版本

- 服务端 ES 7.17.18 (Spring Data ES 期望 8.x)
- Spring Data ES 5.2 → 客户端 `co.elastic.clients` 8.x 兼容 7.17+
- **能跑**,但 ingest pipeline / search template 可能 silently 失败

### 9.2 WAP 直播弹幕协议不一致

- WAP `live/Room.vue:82-103` 用 B 站协议 `DANMU_MSG` + `info[1]`
- 后端 WS 端点 `/ws/danmaku` 是自研协议
- **前端假数据**,目前不修

### 9.3 HLS 播放地址 + MinIO CORS

- `videos.play_url_hd/sd/ld` 存的是 MinIO 路径
- MinIO 默认不开 CORS / Range 友好
- Artplayer 播 HLS 可能播不出来
- **没真测过**,要 demo 时才知道

---

## 10. 优先级矩阵(下一步建议)

| 优先级 | 事项 | 工作量 | 影响 | 必要性 |
|---|---|---|---|---|
| **P0** | 修 Bug 1(READY_TO_PUBLISH 孤儿) | 1-1.5 天 | admin "待发布"菜单永远是空 | 🔴 demo 翻车 |
| **P0** | 修 Bug 4(publish 不检查状态) | 0.5 天 | 安全漏洞 + 转码失败也能上架 | 🔴 安全 |
| **P1** | 修 ES 文档互动数显示(方案 C) | 0.5 天 | 首页卡片显示 0 数字 | 🟡 不致命但尴尬 |
| **P1** | 修 publish→索引自动同步 | 1 天 | 修完才能"全自动 demo" | 🟡 P0 修完后必做 |
| **P2** | 修 Bug 3(publish 检查 process_status) | 0.5 天 | 防止转码失败上架 | 🟢 P0 修完附带 |
| **P2** | 写"上传即处理 vs 审核后处理"产品决策 | 0.5 天 | 用户感知 | 🟢 |
| **P3** | 跨服务写库改事件 | 5+ 天 | 真正"上生产"才需要 | ⚪ 暂缓 |

**完成 P0+P1 后,项目才能"端到端 demo 不翻车"**。

---

## 11. 附:本次探索用的全部 grep 命令清单

```bash
# 疑点 1: 孤儿状态码
grep -rn "STATUS_READY_TO_PUBLISH\s*=" --include="*.java"
grep -rn "setStatus(2)\|setStatus(Manuscript.STATUS_READY" --include="*.java"

# 疑点 2: 上传触发
grep -B 1 -A 5 "videoProcessPort\." ManuscriptServiceImpl.java

# 疑点 3: DLQ
cat VideoMediaProcessDLQConsumer.java

# 疑点 4: reject 后再 approve
grep -A 15 "approveManuscript.*Integer manuscriptId" ManuscriptServiceImpl.java

# 疑点 5+6: edit 和 owner publish
grep -A 20 "updateManuscriptByOwner" ManuscriptServiceImpl.java
grep -A 15 "publishManuscriptByOwner" ManuscriptServiceImpl.java

# 疑点 7+11: publish 过滤
grep -A 15 "public boolean publishManuscript" ManuscriptServiceImpl.java

# 疑点 9: ES 过滤
cat search/mapper/ManuscriptMapper.java  # WHERE m.status = 3

# 疑点 12: dispatchNextIfNeeded
grep -A 20 "private void dispatchNextIfNeeded" VideoMediaProcessConsumer.java

# 疑点 13: draft
grep -rn "\"draft\"" --include="*.java"

# 疑点 14: Video.STATUS 死代码
grep -rn "Video.STATUS_PUBLISHED\|Video.STATUS_PROCESSING" --include="*.java"

# 跨服务直查 DB 全景
grep -rln "from manuscripts\|from videos\|from users" {service}/src/main/java

# 互动直写库
grep -n "manuscriptMapper\." VideoInteractionServiceImpl.java
```

---

## 12. 文档维护说明

- 报告生成: 2026-06-11
- 验证方法: 所有结论都有对应 grep/Read 引用,**不接受任何"我觉得"**
- 重新验证: 跑 §11 命令清单
- 状态变更: 每次改 manuscript 状态机前,要更新 §2
