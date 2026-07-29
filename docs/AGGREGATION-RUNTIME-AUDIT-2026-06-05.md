# Aggregation Runtime Audit - 2026-06-05

当前默认运行态只允许 6 个后端服务：

- `mybilibili-gateway`
- `mybilibili-account-social`
- `mybilibili-video-media`
- `mybilibili-content-interaction`
- `mybilibili-search-recommend`
- `mybilibili-ai`

旧后端模块已归位到 `legacy-services/`，仍保留为源码来源和参考材料，但不能被父 `pom.xml`、默认启动脚本、网关路由或内部调用重新拉回运行时。

## 自动检查

本地检查命令：

```powershell
.\scripts\check-architecture.ps1
```

该命令会执行：

- `scripts/check-feign-boundaries.ps1`
- `scripts/check-aggregation-runtime.ps1`

聚合运行态检查规则：

- 网关 `lb://` 只能指向聚合后的运行服务。
- 父 `pom.xml` 不能再把旧后端列为 Maven module。
- 旧后端源码必须位于 `legacy-services/`。
- `start-all.ps1`、`start-all.bat`、`test-services.sh` 必须包含 6 个运行后端。
- 默认启动脚本不得启动旧后端模块。
- `stop-all.bat` 可以继续清理旧端口，作为本地兜底，不视为运行态违规。

## 旧模块边界

以下旧模块暂时保留，但不属于默认运行态：

- `legacy-services/mybilibili-user`
- `legacy-services/mybilibili-video`
- `legacy-services/mybilibili-danmaku`
- `legacy-services/mybilibili-search`
- `legacy-services/mybilibili-comment`
- `legacy-services/mybilibili-interaction`
- `legacy-services/mybilibili-message`
- `legacy-services/mybilibili-live`
- `legacy-services/mybilibili-analytics`
