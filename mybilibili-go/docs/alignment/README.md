# docs/alignment — 源码级对齐清单（中央索引）

> 目标：让 Go 单体在**行为层面**与 Java 微服务逐业务域**源码级对齐**（状态机/业务规则/事件/接口结构/错误语义一致）。
> 每个域一份 `tasks/XX-*.md`，方法级逐个打勾。**没锁死就是没完成**（文档 3.2 强制规则）。
> 旧版仓库：`mybilibili-cloud`（只读参照）。新版：本仓库 `internal/`。

## 首部清单

| ✅ | 文件 | 覆盖旧版类 | 方法数 | 状态 |
|---|---|---|---|---|
| ☑ | [state-machines.md](state-machines.md) | Manuscript/Video/EditVersion/Collection/Interaction/Analytics | — | ✅ 已建 |
| ☑ | [tasks/01-稿件域.md](tasks/01-稿件域.md) | ManuscriptService | 39 | 已核对 ✅26/🟡10/❌3 |
| ☑ | [tasks/02-互动域.md](tasks/02-互动域.md) | VideoInteraction/WatchHistory/Collection | 32 | 已核对 ✅23/🟡6/❌0 |
| ☑ | [tasks/03-评论域.md](tasks/03-评论域.md) | Comment×26/CreatorComment×5/DynamicComment×6/ProhibitedWord ×6/Report×3/Spam×3 | 49 | 已核对 ✅27/🟡6/❌16 |
| ☑ | [tasks/04-动态弹幕域.md](tasks/04-动态弹幕域.md) | DynamicServiceImpl/DanmakuServiceImpl | 20 | 已核对 ✅17/🟡3/❌0 |
| ☑ | [tasks/05-消息域.md](tasks/05-消息域.md) | MessageServiceImpl/ConversationServiceImpl/MessageSettingServiceImpl | 32 | 已核对 ✅22/🟡6/❌0 |
| ☑ | [tasks/06-用户账号域.md](tasks/06-用户账号域.md) | AdminUser/AuditLog/Captcha/EmailCode/Follow/LoginLog/OperationTask/Privacy | 29 | 已核对 ✅17/🟡6/❌0 |
| ☑ | [tasks/07-AI域.md](tasks/07-AI域.md) | AiSummary/AiSubtitle/CustomerService/CustomerSession/AiApiConfig/AiSkill/ContentReview/AdminAi/SkillRouting/VideoProcessState | ~40 | 已核对 ✅24/🟡8/❌8 |
| ☑ | [tasks/08-搜索推荐域.md](tasks/08-搜索推荐域.md) | VideoSearch×2/VideoRecommend/HotSearch/ManuscriptIndex/UserProfile | ~21 | 已核对 ✅14/🟡7/❌0 |
| ☑ | [tasks/09-创作者分析域.md](tasks/09-创作者分析域.md) | CreatorStatsServiceImpl/SupportTicketServiceImpl | 15 | 已核对 ✅14/🟡1/❌0 |
| ☑ | [tasks/10-字幕直播域.md](tasks/10-字幕直播域.md) | SubtitleServiceImpl/LiveRoomServiceImpl | 27 | 已核对 ✅26/🟡1/❌0 |
| ☑ | [tasks/11-管理域.md](tasks/11-管理域.md) | AdminUserService/AuditLogService/AdminAiService/ContentReviewService | 14 | 已核对 ✅15/🟡0/❌0 |

## 汇总

| 指标 | 值 |
|---|---|
| 旧版方法总数 | ~318（含 ~62 个近似/汇总方法） |
| ✅ 完整等效 | 223 |
| 🟡 部分等效 | 54 |
| ❌ 缺失 | 39（AI 域 8 + 评论审核流 4 = 12，非 AI 域零） |
| 覆盖域 | 11/11（全部落盘） |

## 状态图例

- ✅ 完整等效（Go 有实现且行为对齐）
- 🟡 部分等效（有路由/函数但简化：stub/空数组/缺副作用）
- ❌ 缺失（Go 无对应或仅占位）
- 附注：强制规则违反（事件不落库 / TODO / `not yet implemented` / 空数组占位）单独标 ❌

## 强制规则（来自总文档 3.2）

1. 旧版状态常量新版必须同名同值，不得改值
2. 旧版事件副作用（通知/索引/状态流水）新版必须落库
3. 旧版有该分支新版缺该分支 = 未完成
4. 禁止 TODO / `not yet implemented` / 空数组占位通过验收
5. 新增任何表用 `sql/0xx_*.sql` 迁移文件，别手写建表
6. 一任务一 commit，message 前缀 `align(T批次)任务名`

## 运行验收

```bash
cd /tmp/mybilibili/mybilibili-go && go build ./... && go test ./... -run Contract -vet=off
```