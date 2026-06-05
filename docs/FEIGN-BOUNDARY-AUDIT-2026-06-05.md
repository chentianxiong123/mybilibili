# Feign Boundary Audit - 2026-06-05

本清单用于服务合并后的 Feign 边界治理。当前运行态目标是 6 个后端服务：

- `mybilibili-gateway`
- `mybilibili-account-social`
- `mybilibili-video-media`
- `mybilibili-content-interaction`
- `mybilibili-search-recommend`
- `mybilibili-ai`

原则：

- 跨聚合服务的 Feign 是真实分布式边界，保留。
- 同一聚合服务内部的 Feign 不再新增，已本地化的接口先标记废弃。
- 旧独立模块仍作为参考、回滚和对照材料，暂不物理删除。
- 后续删除旧模块或旧接口必须按 `.trash/YYYYMMDD-HHMMSS/` 逻辑删除规则执行。

## 当前清单

| Feign 接口 | 原始消费模块 | 目标运行服务 | 聚合后关系 | 当前动作 |
| --- | --- | --- | --- | --- |
| `com.mybilibili.analytics.feign.DanmakuClient` | `mybilibili-analytics` | `mybilibili-content-interaction` | `search-recommend -> content-interaction` | keep |
| `com.mybilibili.danmaku.feign.VideoClient` | `mybilibili-danmaku` | `mybilibili-video-media` | `content-interaction -> video-media` | keep |
| `com.mybilibili.comment.feign.UserClient` | `mybilibili-comment` | `mybilibili-account-social` | `content-interaction -> account-social` | keep |
| `com.mybilibili.comment.feign.MessageClient` | `mybilibili-comment` | `mybilibili-account-social` | `content-interaction -> account-social` | keep |
| `com.mybilibili.comment.feign.ManuscriptClient` | `mybilibili-comment` | `mybilibili-video-media` | `content-interaction -> video-media` | keep |
| `com.mybilibili.comment.feign.LikeClient` | `mybilibili-comment` | `mybilibili-content-interaction` | `content-interaction` 内部调用 | local-adapter, deprecated |
| `com.mybilibili.comment.feign.DynamicClient` | `mybilibili-comment` | `mybilibili-content-interaction` | `content-interaction` 内部调用 | local-adapter, deprecated |
| `com.mybilibili.comment.feign.ContentReviewClient` | `mybilibili-comment` | `mybilibili-ai` | `content-interaction -> ai` | keep |
| `com.mybilibili.interaction.feign.VideoClient` | `mybilibili-interaction` | `mybilibili-video-media` | `content-interaction -> video-media` | keep |
| `com.mybilibili.interaction.feign.MessageClient` | `mybilibili-interaction` | `mybilibili-account-social` | `content-interaction -> account-social` | keep |
| `com.mybilibili.ai.feign.ReportCallbackClient` | `mybilibili-ai` | `mybilibili-content-interaction` | `ai -> content-interaction` | keep |
| `com.mybilibili.user.feign.UserProfileClient` | `mybilibili-user` | `mybilibili-content-interaction` | `account-social -> content-interaction` | keep |
| `com.mybilibili.user.feign.UserClient` | `mybilibili-user` | `mybilibili-account-social` | `account-social` 内部调用 | local-adapter, deprecated |
| `com.mybilibili.user.feign.ManuscriptClient` | `mybilibili-user` | `mybilibili-video-media` | `account-social -> video-media` | keep |
| `com.mybilibili.message.feign.UserClient` | `mybilibili-message` | `mybilibili-account-social` | `account-social` 内部调用 | local-adapter, deprecated |
| `com.mybilibili.search.feign.UserProfileClient` | `mybilibili-search` | `mybilibili-content-interaction` | `search-recommend -> content-interaction` | keep |
| `com.mybilibili.video.feign.VideoProcessClient` | `mybilibili-video` | `mybilibili-ai` | `video-media -> ai` | keep |
| `com.mybilibili.video.feign.VideoPipelineClient` | `mybilibili-video` | `mybilibili-ai` | `video-media -> ai` | keep |
| `com.mybilibili.video.feign.UserClient` | `mybilibili-video` | `mybilibili-account-social` | `video-media -> account-social` | keep |
| `com.mybilibili.video.feign.MessageClient` | `mybilibili-video` | `mybilibili-account-social` | `video-media -> account-social` | keep |

## 已冻结的内部 Feign

这些接口已经由聚合服务内的本地 Adapter 接管，后续不要再新增依赖：

| 废弃接口 | 本地 Adapter | 说明 |
| --- | --- | --- |
| `com.mybilibili.message.feign.UserClient` | `MessageUserClientLocalAdapter` | 私信模块查用户，聚合后改成本地 `UserService` 调用 |
| `com.mybilibili.user.feign.UserClient` | `UserSelfClientLocalAdapter` | 用户模块自调用，聚合后改成本地 `UserService` 调用 |
| `com.mybilibili.comment.feign.LikeClient` | `CommentLikeClientLocalAdapter` | 评论查点赞状态/数量，聚合后改成本地 `InteractionService` 调用 |
| `com.mybilibili.comment.feign.DynamicClient` | `CommentDynamicClientLocalAdapter` | 动态评论数更新，聚合后改成本地 `DynamicService` 调用 |

## 后续删除节奏

1. 当前阶段：保留文件，接口标记 `@Deprecated`，聚合服务启动时只注册跨聚合 Feign。
2. 6 服务全链路测试稳定后：禁止新代码注入上述内部 Feign，新增内部调用应直接依赖本地 Service 或明确的本地门面。
3. CI、Docker、K8s 都不再启动旧独立服务后：把不再需要的旧模块或接口按 `.trash/YYYYMMDD-HHMMSS/` 规则逻辑删除。

## 自动检查

本地检查命令：

```powershell
.\scripts\check-feign-boundaries.ps1
```

`scripts/build-all.bat` 已在 Maven 构建前通过 `scripts/check-architecture.ps1` 执行该检查。检查规则：

- 新增同聚合服务内部 Feign 直接失败。
- 当前 4 个历史内部 Feign 必须保持 `@Deprecated`。
- 当前 4 个历史内部 Feign 必须有对应本地 Adapter。
- 聚合启动类不得把这些内部 Feign 注册为远程 Feign Client。
