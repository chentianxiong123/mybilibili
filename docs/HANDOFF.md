# mybilibili-cloud 项目交接文档

> 最后更新: 2026-05-31 | 提交: 79c65dd | JDK: Corretto 17

---

## 一、项目概览

Spring Cloud 微服务仿 B 站平台，13 个后端模块 + 2 个前端（用户端 + 管理端）。

| 模块 | 端口 | 职责 |
|------|------|------|
| gateway | 8080 | WebFlux 网关，JWT 鉴权 + 路由 |
| user | 8081 | 用户/管��员/角色/权限 |
| video | 8082 | 稿件/视频/分类/轮播图/统计 |
| comment | 8083 | 评论/举报/违禁词/内容审核 |
| interaction | 8084 | 动态/收藏/合集/观看历史 |
| danmaku | 8085 | 弹幕 |
| search | 8086 | ES 搜索/推荐/热词/索引管理 |
| message | 8087 | 私信/系统通知 |
| ai | 8088 | AI 字幕/总结/客服/技能/渠道 |
| live | 8089 | 直播间/连麦/会议/WebSocket |
| common | - | 共享实体/工具/Security/Storage |
| mq | - | RocketMQ 消息定义 |

前端: `mybilibili-web`(用户端 Vite+Vue3) / `mybilibili-admin-web`(管理端)

---

## 二、本次提交内容 (79c65dd)

### 2.1 Spring Security RBAC 权限体系

**网关层 (WebFlux)**
- 删除旧 `AuthFilter`，引入 `spring-boot-starter-security`
- 新增: `GatewaySecurityConfig` / `JwtReactiveAuthenticationManager` / `BearerTokenServerAuthenticationConverter` / `GatewayAuthorizationManager` / `SecurityUserHeaderFilter`
- `GatewayRequestPolicy`: fail-closed 路径分类（公开 / 需登录 / 需管理员 / 需超管）
- 权限码校验: 管理员 token 携带 `permissions` claim，网关解析为 `PERMISSION_xxx` authority，按路径前缀匹配所需权限

**服务层 (Servlet)**
- `SecurityConfig`: 注册 `SecurityFilterChain`，`JwtFilter` 在 `UsernamePasswordAuthenticationFilter` 前执行
- `JwtFilter`: 解析 JWT claims → 设置 `SecurityContextHolder`，authority 包括 `ROLE_USER` / `ROLE_ADMIN` / `ROLE_SUPER_ADMIN` / `PERMISSION_xxx`
- `JwtRequestPolicy`: 方法+路径级权限规则，`/category` POST 需 `category:manage`，`/banner-images/` POST 需 `banner:manage` 等

**登录签发**
- `AdminUserServiceImpl.login()`: 查询角色 → 查询权限码 → `JwtUtils.generateAdminToken(id, username, role, permissionCodes)`
- `JwtUtils.generateAdminToken`: 新增 4 参数重载，将 `permissionCodes` 写入 JWT `permissions` claim

**前端管理端**
- `stores/admin.js`: 登录后保存 `permissions` 到 localStorage，提供 `hasPermission(code)` 方法
- `App.vue`: 菜单按 `permission` 字段过滤可见性
- `router/index.js`: 路由守卫检查 `meta.permission`，无权限时重定向

### 2.2 17 个权限码

| ID | 名称 | code | 分配角色 |
|----|------|------|----------|
| 1 | 用户管理 | user:manage | 超管 + 普管 |
| 2 | 视频管理 | video:manage | 超管 + 普管 |
| 3 | 评论管理 | comment:manage | 超管 + 普管 |
| 4 | 分类管理 | category:manage | 超管 |
| 5 | 标签管理 | tag:manage | 超管 |
| 6 | 内容审核 | review:manage | 超管 + 普管 |
| 7 | 统计分析 | statistics:manage | 超管 + 普管 |
| 8 | 角色管理 | role:manage | 超管 |
| 9 | 管理员管理 | admin:manage | 超管 |
| 10 | 安全设置 | security:manage | 超管 |
| 11 | 直播管理 | live:manage | 超管 + 普管 |
| 12 | 会议管理 | meeting:manage | 超管 + 普管 |
| 13 | 存储管理 | storage:manage | 超管 |
| 14 | 轮播图管理 | banner:manage | 超管 |
| 15 | 搜索索引管理 | search:manage | 超管 |
| 16 | AI管理 | ai:manage | 超管 + 普管 |
| 17 | 消息管理 | message:manage | 超管 |

种子数据: `init/mybilibili-mysql.sql` (fresh install)
迁移脚本: `init/V9__admin_permission_hardening.sql` (existing DB)

### 2.3 MinIO 对象存储迁移

- `StorageService` 接口 + `MinioStorageService` 实现，统一上传/下载 URL 生成
- `StorageKeys`: 集中定义所有存储路径 key（稿件视频/封面/头像/轮播图/弹幕等）
- `StorageMigrationTool`: 本地文件到 MinIO
- `StorageMigrationController`: `/admin/storage/migrate` 触发迁移（需 `storage:manage` 权限）
- 前端播放/上传全部走 MinIO URL，不再有本地 `/uploads/` 兜底

### 2.4 分片上传

- 后端: `ManuscriptUploadSessionService` 管理上传会话（Redis 存储），`/manuscript/upload-session` / `/manuscript/upload-chunk` / `/manuscript/upload-complete`
- 前端: `useChunkedManuscriptUpload.js` composable，串行分片 + 进度回调
- 设计: 串行上传（非并发），每个分片独立确认

### 2.5 服务扫描修复

- `CommentApplication`: `@ComponentScan` 新增 `com.mybilibili.common`
- `MessageApplication`: `@ComponentScan` 新增 `com.mybilibili.common`
- `SearchApplication`: `@ComponentScan` 新增 `com.mybilibili.common`
- 效果: 这三个服务现在也启用 `SecurityConfig` + `JwtFilter`

### 2.6 网关路径补全

新增到 `AUTH_REQUIRED_PATH_PREFIXES`:
- `/api/user/privacy` (隐私设置)
- `/api/user/pinned-video` (置顶视频)
- `/api/manuscript/me/` (我的稿件)
- `/api/live/room/create`, `/api/live/room/my`

新增到 `PUBLIC_EXACT_PATHS`:
- `/api/live/room/srs/hook` (SRS 回调)

---

## 三、测试状态

```
JDK: Corretto 17.0.16
mvn test → BUILD SUCCESS
Tests run: 56, Failures: 0, Errors: 0
```

测试分布:
- `mybilibili-common`: JwtFilterTest(4) + JwtRequestPolicyTest(3)
- `mybilibili-gateway`: GatewayRequestPolicyTest(3) + GatewayAuthorizationManagerTest(6) + JwtReactiveAuthenticationManagerTest(3)
- `mybilibili-video`: ManuscriptControllerUploadTest(1) + ManuscriptUploadSessionServiceTest(3)
- `mybilibili-live`: LiveConstantsTest(1) + RequestUserResolverTest(4) + LiveRoomServiceImplTest(8) + LiveLinkmicServiceTest(8) + MeetingServiceTest(10) + MeetingSessionRegistryTest(3) + MeetingWebSocketHandlerTest(8) + MeetingWsMessageTypeTest(1)
- 其余模块: 无测试或仅编译验证

---

## 四、构建命令

```powershell
# 设置 JDK 17
$env:JAVA_HOME='D:\Program Files (x86)\corretto-17'
$env:Path="$env:JAVA_HOME\bin;$env:Path"

# 全量编译
mvn compile

# 全量测试
mvn test

# 单模块测试
mvn -pl mybilibili-gateway -am test
```

---

## 五、已知状态与待办

### 已完成
- Spring Security 双层鉴权（网关 + 服务）
- 17 个权限码 RBAC，前后端对齐
- MinIO 存储统一化
- 分片上传 + 进度条
- 直播/会议/WebSocket 功能
- AI 字幕/总结/客服/技能管理
- 数据库备份 (backups/)

### 用户侧权限
- 普通用户不需要 RBAC，当前用 `ROLE_USER` + 业务层状态检查即可
- 如需用户等级限制（如 Lv1 才能评论），在 Controller 层按 `user.getLevel()` 判断

### 可优化项
- 全文搜索 `IndexAdminController` 目前无权限保护（路径 `/search/admin/index` 已在网关层保护）
- `/api/statistics/` 路径在网关层标记为 admin，但前端用户端 Dashboard 也调用了统计接口，需确认是否需要拆分
- `StorageMigrationController` 仅在配置了 `storage.migration.local-path` 时激活
- WebSocket 会议信令 (`/ws/meeting`) 当前走网关透传，未经过 Spring Security 过滤器链
- 前端管理端部分 API 调用路径用了 `/manuscript/admin/` 而非 `/api/manuscript/admin/`，依赖 Vite proxy 转发

---

## 六、关键文件索引

### 安全体系
- `mybilibili-gateway/src/main/java/.../gateway/security/GatewaySecurityConfig.java`
- `mybilibili-gateway/src/main/java/.../gateway/security/GatewayAuthorizationManager.java`
- `mybilibili-gateway/src/main/java/.../gateway/filter/GatewayRequestPolicy.java`
- `mybilibili-common/src/main/java/.../common/config/SecurityConfig.java`
- `mybilibili-common/src/main/java/.../Filter.java`
- `mybilibili-common/src/main/java/.../common/config/JwtRequestPolicy.java`
- `mybilibili-common/src/main/java/.../common/utils/JwtUtils.java`

### 存储
- `mybilibili-common/src/main/java/.../common/storage/StorageService.java`
- `mybilibili-common/src/main/java/.../common/storage/MinioStorageService.java`
- `mybilibili-common/src/main/java/.../common/storage/StorageKeys.java`

### 数据库
- `init/mybilibili-mysql.sql` (种子数据)
- `init/V6__meeting_reserve.sql` ~ `init/V9__admin_permission_hardening.sql` (迁移)
