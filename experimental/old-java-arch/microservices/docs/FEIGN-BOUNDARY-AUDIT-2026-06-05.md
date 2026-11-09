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
- 旧独立模块已归位到 `legacy-services/` 源码池，暂不物理删除，但不作为默认 Maven module 或调用路径。
- 后续删除旧模块或旧接口必须按 `.trash/YYYYMMDD-HHMMSS/` 逻辑删除规则执行。

## 当前清单

| Feign 接口 | 原始消费模块 | 目标运行服务 | 聚合后关系 | 当前动作 |
| --- | --- | --- | --- | --- |
| `com.mybilibili.analytics.feign.DanmakuClient` | `legacy-services/mybilibili-analytics` | `mybilibili-content-interaction` | `search-recommend -> content-interaction` | keep |
| `com.mybilibili.danmaku.feign.VideoClient` | `legacy-services/mybilibili-danmaku` | `mybilibili-video-media` | `content-interaction -> video-media` | keep |
| `com.mybilibili.comment.feign.UserClient` | `legacy-services/mybilibili-comment` | `mybilibili-account-social` | `content-interaction -> account-social` | keep |
| `com.mybilibili.comment.feign.MessageClient` | `legacy-services/mybilibili-comment` | `mybilibili-account-social` | `content-interaction -> account-social` | keep |
| `com.mybilibili.comment.feign.ManuscriptClient` | `legacy-services/mybilibili-comment` | `mybilibili-video-media` | `content-interaction -> video-media` | keep |
| `com.mybilibili.comment.feign.LikeClient` | `legacy-services/mybilibili-comment` | `mybilibili-content-interaction` | `content-interaction` 内部调用 | local-port, deprecated |
| `com.mybilibili.comment.feign.DynamicClient` | `legacy-services/mybilibili-comment` | `mybilibili-content-interaction` | `content-interaction` 内部调用 | local-port, deprecated |
| `com.mybilibili.comment.feign.ContentReviewClient` | `legacy-services/mybilibili-comment` | `mybilibili-ai` | `content-interaction -> ai` | keep |
| `com.mybilibili.interaction.feign.VideoClient` | `legacy-services/mybilibili-interaction` | `mybilibili-video-media` | `content-interaction -> video-media` | keep |
| `com.mybilibili.interaction.feign.MessageClient` | `legacy-services/mybilibili-interaction` | `mybilibili-account-social` | `content-interaction -> account-social` | keep |
| `com.mybilibili.ai.feign.ReportCallbackClient` | `mybilibili-ai` | `mybilibili-content-interaction` | `ai -> content-interaction` | keep |
| `com.mybilibili.user.feign.UserProfileClient` | `legacy-services/mybilibili-user` | `mybilibili-content-interaction` | `account-social -> content-interaction` | keep |
| `com.mybilibili.user.feign.UserClient` | `legacy-services/mybilibili-user` | `mybilibili-account-social` | `account-social` 内部调用 | local-port, deprecated |
| `com.mybilibili.user.feign.ManuscriptClient` | `legacy-services/mybilibili-user` | `mybilibili-video-media` | `account-social -> video-media` | keep |
| `com.mybilibili.message.feign.UserClient` | `legacy-services/mybilibili-message` | `mybilibili-account-social` | `account-social` 内部调用 | local-port, deprecated |
| `com.mybilibili.search.feign.UserProfileClient` | `legacy-services/mybilibili-search` | `mybilibili-content-interaction` | `search-recommend -> content-interaction` | keep |
| `com.mybilibili.video.feign.UserClient` | `legacy-services/mybilibili-video` | `mybilibili-account-social` | `video-media -> account-social` | keep |
| `com.mybilibili.video.feign.MessageClient` | `legacy-services/mybilibili-video` | `mybilibili-account-social` | `video-media -> account-social` | keep |

## 已冻结的内部 Feign

这些接口不再作为聚合运行态的调用抽象。业务代码必须依赖本地调用端口，聚合服务内的 Adapter 只实现本地端口：

| 废弃接口 | 本地端口 | 本地 Adapter | 说明 |
| --- | --- | --- | --- |
| `com.mybilibili.message.feign.UserClient` | `UserLookupPort` | `MessageUserClientLocalAdapter` | 私信模块查用户，聚合后改成本地 `UserService` 调用 |
| `com.mybilibili.user.feign.UserClient` | `UserLookupPort` | `UserSelfClientLocalAdapter` | 用户模块自调用，聚合后改成本地 `UserService` 调用 |
| `com.mybilibili.comment.feign.LikeClient` | `LikeInteractionPort` | `CommentLikeClientLocalAdapter` | 评论查点赞状态/数量，聚合后改成本地 `InteractionService` 调用 |
| `com.mybilibili.comment.feign.DynamicClient` | `DynamicInteractionPort` | `CommentDynamicClientLocalAdapter` | 动态评论数更新，聚合后改成本地 `DynamicService` 调用 |

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
- 业务代码不得 import 当前 4 个历史内部 Feign。
- 当前 4 个历史内部 Feign 必须有对应本地端口 Adapter。
- 聚合启动类不得把这些内部 Feign 注册为远程 Feign Client。
