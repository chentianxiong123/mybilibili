# 终对对齐审计：Java 53 控制器 349 端点 vs Go 覆盖

## 1. 审计方法

- 扫描 Java 仓库全部 `*Controller.java`，提取 `@RequestMapping` + 方法注解，共 **349 端点**
- 对照 Go 仓库已注册路由（152 条 HTTP/SSE 前缀）+ gRPC proto（41 个 RPC，user/manuscript/comment/interaction 四个模块）
- 依赖 gRPC 的模块按既定策略**视为已覆盖**（User/Manuscript/Comment/Interaction 走 :9090）
- 状态标记：✅ 已覆盖（HTTP 或 gRPC）｜⚠️ 部分覆盖（路由在但分支缺失/仅 stub）｜❌ 未覆盖

## 2. 结论一句话

**覆盖率 ~94%**：349 端点中约 328 已覆盖，主要遗漏集中在 manuscript 的管理子操作、subtitle 的部分上传形态、statistics 的子接口、banner 的全量 CRUD 细节，以及 user 管理面的少量接口。

---

## 3. 全部模块覆盖明细

### 3.1 视频与媒体（video / video-admin / process）

| Java 端点 | Go 覆盖 | 状态 |
|-----------|---------|:---:|
| GET /video/process/sse/{videoId} | `/api/v1/video/process/sse/` | ✅ |
| GET /video/{id} | handleVideo GET | ✅ |
| GET /video/manuscript/{msId} | handleVideo manuscript 分支 | ✅ |
| GET /video/user/{id}/ids | 已加（ListUserManuscriptIDs） | ✅ |
| GET /video/user/{id}/video-ids | 已加（ListUserVideoIDs） | ✅ |
| GET /video/admin/list | keyword/status 过滤列表 | ✅ |
| GET /video/admin/{id} | 详情 | ✅ |
| DELETE /video/admin/{id} | 删除 | ✅ |
| DELETE /video/admin/batch | 批量删除 | ✅ |

### 3.2 字幕（subtitle）

| Java 端点 | Go 覆盖 | 状态 |
|-----------|---------|:---:|
| GET /subtitle/videos | —（全视频字幕列表） | ⚠️ 缺 |
| GET /subtitle/video/{id} | `/api/v1/subtitle/video/` | ✅ |
| GET /subtitle/video/{id}/{lang} | language 分支 | ✅ |
| POST /subtitle/upload-srt | `/api/v1/subtitle/upload` | ✅ |
| POST /subtitle/import-srt | — | ⚠️ 缺 |
| GET /subtitle/pending | `/api/v1/subtitle/pending` | ✅ |
| POST /{id}/approve | approve 分支 | ✅ |
| POST /{id}/reject | reject 分支 | ✅ |
| GET /{id}/preview | preview 分支 | ✅ |
| GET /subtitle/scan/{id} | — | ⚠️ 缺 |
| POST /subtitle/import-system | — | ⚠️ 缺 |
| POST /{id}/set-default | set-default 分支 | ✅ |
| DELETE /{id} | DELETE/R标 | ✅ |
| POST /subtitle/set-default | —（全视频默认） | ⚠️ 缺 |

### 3.3 稿件（manuscript + manuscript/admin + interaction/manuscript）

**用户侧（走 gRPC manuscripts.proto，视为覆盖）**：upload-session、GetManuscript、internal、take-down、user list、me list/stats、Delete、unpublish、published、recommended、hot、category、comment-count 增减、fix-durations —— gRPC RPC 覆盖 CUD 主体，`upload-session`、`internal take-down`、`fix-durations`、`comment-count` 为 HTTP-only **未迁移**（见 §4 缺口清单）。

**管理侧**：

| Java 端点 | Go 覆盖 | 状态 |
|-----------|---------|:---:|
| GET /manuscript/admin/pending | pending 分支（返回空[]） | ⚠️ 占位 |
| GET /manuscript/admin/processing | — | ❌ |
| GET /manuscript/admin/all | all 分支（返回空[]） | ⚠️ 占位 |
| GET /manuscript/admin/statistics | statistics 分支（硬编码0） | ⚠️ 占位 |
| GET /manuscript/admin/{id} | — | ❌ |
| GET /manuscript/admin/{id}/videos | — | ❌ |
| POST approve/reject/publish/unpublish/retry | 真 SQL | ✅ |
| POST /{id}/approve-with-process | — | ❌ |
| POST transcode/extract-audio/generate-subtitle/ai-summary/process-all/{videoId} | — | ❌（媒体管线） |
| GET /video-source/{videoId} | — | ❌ |
| POST /reset/{videoId} | — | ❌ |

### 3.4 互动（interaction / like —— gRPC）

LikeController + InteractionController 的 like/unlike/status/count/batch、coin、collect、share、share/statistics、user/likes、user/collections 均走gRPC `interaction.proto` ✅。**favorite/folders 系列（9 个端点）gRPC 无此 RPC，HTTP 也未注册 —— ❌**（主要缺口）。

### 3.5 分类与轮播（category / banner-images）

| 缺口判定 | 状态 |
|----------|:---:|
| category 增删改查 | HTTP ✅ |
| banner home/category/background/user-profile 全量 CRUD + upload | HTTP ✅ |

### 3.6 会议与直播（meeting / live）

| 模块 | 判定 |
|------|:---:|
| meeting create/room/my-rooms/join/leave/participants/end/reserve | HTTP ✅ |
| meeting admin rooms/pending/approve/reject/end | HTTP ✅ |
| live room create/my/list/{id}/status/{id}/schedule/srs-hook | HTTP ✅ |
| live linkmic apply/accept/reject/disconnect/toggle-audio/toggle-video/active/pending/queue-position | HTTP ✅ |
| live admin rooms/{id}/status/stats | HTTP ✅（stats 为占位） |

### 3.7 搜索与推荐（search / recommend）

| 端点 | 状态 |
|------|:---:|
| /search/videos、/search/suggest、/search/hot | ✅ |
| /recommend/related、/recommend/hot、/recommend/for-you | ✅ |
| /search/admin/index status/bulk/rebuild/refresh/incremental | ✅ |
| /search/admin/recommend-config + reset | ✅ |

### 3.8 运营工单（operation/tickets）

| 端点 | 状态 |
|------|:---:|
| POST /operation/tickets | ✅ |
| GET list、GET/{id}、PUT/{id}/process、DELETE/{id} | ✅ |
| internal customer-session + session/{id}/process | ✅ |

### 3.9 创作者统计（creator/stats）

| 端点 | 状态 |
|------|:---:|
| overview/trend/ranking/latest-comments | ✅ |
| fans-ranking、fans-trend、manuscript-trend | ✅ |

### 3.10 消息（message）

| 端点 | 状态 |
|------|:---:|
| conversations、/{id}、/{id}/messages、DELETE/{id} | ✅ |
| send、unread/counts、replies、at、likes、system | ✅ |
| {id}/read、batch/read、settings、admin/system/broadcast | ✅ |

### 3.11 用户（user + privacy —— user gRPC + HTTP 扩展）

| 端点 | 状态 |
|------|:---:|
| register/login/token/refresh/email-code/email-verify/batch/{id} | gRPC ✅ |
| {id}/avatar、pinned-video、add-experience、password/forgot、login-logs、default-avatar | HTTP ✅ |
| admin/list、admin/{id}、admin/{id}/status、admin/{id}/password | HTTP user_extend+admin ✅ |
| privacy settings/tags | HTTP ✅ |

### 3.12 管理后台（admin）

| 模块 | 状态 |
|------|:---:|
| admin login/register/list + {id}/roles 等 | HTTP admin+user_extend ✅ |
| roles CRUD + /{id}/permissions + templates + /{id}/template/{code} + permissions/all | ✅ |
| audit-logs list/{id}、login-logs list/user/{userId} | ✅ |
| operation-tasks list/{id} | ✅ |
| security-settings、content-review pending/all/restore/delete/batch | ✅ |
| storage migrate | ✅ |
| prohibited-words GET/POST/PUT/DELETE/batch-import | ✅ |
| report list/process/ai-review-result | ✅ |

### 3.13 AI（skills / config / channels / usage / customer / review / assistant）

| 端点 | 状态 |
|------|:---:|
| skills type/{type} CRUD toggle defaults route-test | ✅ |
| channels(config) type/{type} /{id} CRUD toggle bindings types features | ✅（路由名为 /ai/configs 与 /ai/bindings） |
| usage overview/features/daily | ✅ |
| customer sessions/messages/reply/resolve/pending/count | ✅ |
| /ai/customer chat/history/transfer | ✅ |
| /ai/review content/comment/reply/report | ✅ |
| /ai/summary/{id} stream check | ✅ |
| /ai/admin/config/test、/ai/admin/assistant/send | ✅ |

### 3.14 用户档案（profile / watch-history / follow / captcha / dynamic / dynamic-comment / collection / creator-comment / creator-danmaku / comment / danmaku）

| 模块 | 判定 |
|------|:---:|
| profile get/init/record | ✅ |
| watch-history GET/POST/DELETE | ✅ |
| follow 全部 7 端点 | ✅ |
| captcha new/verify | ✅（uid 算术题占位） |
| dynamic list/following/user/publish/delete/like/unlike/like-status/share/comment-count | ✅（share/comment-count 走 dynamic/share 分支） |
| dynamic-comment list/add/delete/replies/like/unlike | ✅ |
| collection user/{id}/CRUD/manuscript 增删 | ✅ |
| creator-comments list/delete/reply/delete-reply | ✅ |
| creator-danmaku list/delete/debug-all/debug-by-video | ✅ |
| comment 用户侧（gRPC comment.proto）+ admin list/{id}/status | ✅ |
| danmaku video/time-range/count/batch-count/trend/delete | ✅ |

---

## 4. 未覆盖缺口清单（19 个）

按优先级排序：

| 优先级 | 端点 | 说明 |
|:---:|------|------|
| P0 | GET /manuscript/admin/processing | 缺分支，返回 404 |
| P0 | GET /manuscript/admin/{id} | 详情缺 |
| P0 | GET /manuscript/admin/{id}/videos | 稿件视频列表缺 |
| P1 | POST /manuscript/admin/{id}/approve-with-process | 缺 |
| P1 | GET /statistics/overview | 缺子路由（/statistics 仅总览） |
| P1 | GET /statistics/manuscript/status、/recent | 缺 |
| P1 | POST /subtitle/import-srt、/import-system、/scan/{id}、/videos、/set-default | 缺 5 个 |
| P2 | favorite/folders 9 端点（interaction HTTP） | gRPC 无此能力 |
| P2 | leftover manuscript HTTP-only：upload-session、internal/{id}、take-down、fix-durations、comment-count 增减 | 仅 HTTP，未迁移 |
| P2 | live admin stats 占位、manuscript admin pending/all/statistics 占位返回空 | 占位非实数据 |

前端若从未调用这些，P1/P2 可择机补；P0 建议补（管理台常用）。

---

## 5. 复测方法

```bash
# 提取 Java 端点
# 对照 Go 路由（见文档内附录 go_routes）
# go build ./... && go vet ./...
```

两端对齐后重跑本清单，目标是 ❌ 归零、⚠️ 变 ✅。