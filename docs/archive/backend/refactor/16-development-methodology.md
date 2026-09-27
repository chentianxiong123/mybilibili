# 开发方法论

## 1. 核心理念

### 1.1 技术栈

| 语言 | 适用服务 | 原因 |
|------|---------|------|
| **Go** | gateway、core、realtime、search、ads | 标准库即可，不需要框架 |
| **Java GraalVM** | media、store | 支付链路需要事务保障，FFmpeg 生态 |

### 1.2 Go 不需要框架

Go 标准库 + gRPC 足够，不引入 Gin、Kratos、go-zero 等框架。

```
gRPC 负责通信（替代 HTTP 框架）
标准库 net/http 只做健康检查 / metrics 暴露
database/sql 操作数据库
```

### 1.3 契约先行

1. 先定义 proto 文件（接口契约）
2. 生成 gRPC 代码
3. 服务端和客户端各写各的，互不阻塞

---

## 2. 目录结构

```
mybilibili/
├── cmd/server/          ← 启动入口（每个服务一个 main.go）
├── internal/
│   ├── core/            ← 核心业务（用户/内容/评论/点赞）
│   │   ├── server.go       ← gRPC 服务注册 + 启动
│   │   ├── handler.go      ← proto 接口实现（收请求，调 service）
│   │   ├── service.go      ← 业务逻辑
│   │   └── repository.go   ← 数据访问（MySQL）
│   ├── realtime/         ← 实时服务（弹幕/消息/通知）
│   ├── search/           ← 搜索服务（Bleve 嵌入式）
│   └── ads/              ← 广告服务（计数）
├── pkg/
│   ├── abstraction/      ← 抽象层接口（已定义）
│   ├── config/           ← 配置加载
│   ├── logger/           ← 日志
│   └── middleware/        ← gRPC 拦截器（JWT 解析、日志、恢复）
├── proto/                ← proto 定义 + 生成的 Go 代码
├── deploy/               ← 部署配置
├── docs/                 ← 文档
└── go.mod
```

---

## 3. 分层职责

```
handler（接收请求，校验参数，返回响应）
  → service（业务逻辑，组合各种数据源）
    → repository（数据库操作，只做 CRUD）
    → abstraction（抽象层接口，可切换实现）
```

**规则：**
- handler 不调数据库，只调 service
- service 不直接调数据库，通过 repository 或 abstraction 访问
- repository 只做 SQL 操作，不做业务判断
- 同层之间不互相调用（避免循环依赖）

---

## 4. 命名规范

### 4.1 Go

| 项 | 规范 | 示例 |
|----|------|------|
| 文件命名 | 蛇形命名 | `user_service.go` |
| 函数/方法 | 驼峰，导出大写 | `CreateUser()` |
| 变量 | 驼峰，简短 | `userID`, `db` |
| 接口 | 结尾 er | `Storager`, `Cacher` |
| 错误 | 以小写开头 | `errors.New("user not found")` |
| 包名 | 小写，单数 | `package core` |

### 4.2 proto

| 项 | 规范 | 示例 |
|----|------|------|
| 服务名 | 大驼峰 | `UserService` |
| RPC 方法 | 动词 + 名词 | `CreateUser`, `GetUser` |
| 消息名 | 大驼峰 | `CreateUserRequest` |
| 字段名 | 下划线 | `user_id` |

---

## 5. 错误处理

### 5.1 统一错误码

```go
// pkg/abstraction/errors.go
var (
    ErrNotFound       = status.Error(codes.NotFound, "resource not found")
    ErrInvalidInput   = status.Error(codes.InvalidArgument, "invalid input")
    ErrUnauthenticated = status.Error(codes.Unauthenticated, "unauthorized")
    ErrInternal       = status.Error(codes.Internal, "internal error")
)
```

### 5.2 错误处理规则

- handler 层：统一捕获错误，返回 gRPC 错误码
- service 层：返回业务错误，不处理 HTTP 状态码
- repository 层：返回数据库错误，不包装业务逻辑

```go
// handler 层：接收请求，校验参数，返回响应
func (s *UserHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    if req.Username == "" || req.Password == "" {
        return nil, ErrInvalidInput
    }
    return s.service.Login(ctx, req)
}

// service 层：业务逻辑
func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    user, err := s.repo.GetByUsername(ctx, req.Username)
    if err != nil {
        return nil, ErrNotFound
    }
    if !bcrypt.CompareHashAndPassword(user.Password, req.Password) {
        return nil, ErrUnauthenticated
    }
    return &pb.LoginResponse{UserId: user.ID}, nil
}

// repository 层：只做 SQL
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*User, error) {
    row := r.db.QueryRowContext(ctx, "SELECT id, password FROM users WHERE username = ?", username)
    // ...
}
```

---

## 6. 配置管理

### 6.1 配置来源

| 优先级 | 来源 | 说明 |
|--------|------|------|
| 最高 | 环境变量 | 运行时覆盖，适合敏感信息 |
| 中 | 配置文件 | `config.yaml`，适合非敏感配置 |
| 最低 | 默认值 | 代码里写死，开发环境用 |

### 6.2 配置结构

```yaml
server:
  port: 50051
  name: core

database:
  driver: mysql
  dsn: "root:password@tcp(localhost:3306)/mybilibili"

redis:
  addr: "localhost:6379"

etcd:
  endpoints: ["localhost:2379"]

nats:
  url: "nats://localhost:4222"

log:
  level: info
  format: json
```

---

## 7. 抽象层使用规则

### 7.1 什么时候用抽象层

| 场景 | 用抽象层？ | 原因 |
|------|-----------|------|
| 缓存 | ✅ CacheStore | 可在 Redis / 本地内存间切换 |
| 消息队列 | ✅ MessageQueue | 可在 NATS / 内存间切换 |
| 服务发现 | ✅ ServiceDiscovery | 可在 etcd / 配置间切换 |
| 搜索 | ✅ SearchEngine | 可在 Bleve / ES 间切换 |
| 文件存储 | ✅ StorageService | 可在 MinIO / 本地文件间切换 |
| 数据库 | ❌ 直接 MySQL | 不会换，不需要抽象 |
| JWT | ❌ 直接 golang-jwt | 不会换，不需要抽象 |

### 7.2 配置驱动

```yaml
infrastructure:
  cachestore:
    type: redis    # redis / memory
    endpoint: localhost:6379
  messagequeue:
    type: nats     # nats / memory
    url: nats://localhost:4222
  servicediscovery:
    type: etcd     # etcd / static
    endpoints: ["localhost:2379"]
```

---

## 8. 开发流程

### 8.1 每个服务的开发步骤

```
1. 定义 proto 接口（契约先行）
2. 生成 gRPC 代码
3. 写 repository（数据访问）
4. 写 service（业务逻辑）
5. 写 handler（gRPC 接口实现）
6. 写 server.go（注册服务）
7. 写 main.go（启动入口）
8. 跑通一条链路验证
9. 写测试（对拍测试）
```

### 8.2 验证方式

```
启动服务 → 用 grpcurl 或写测试客户端调接口
不对旧 Java 服务做依赖，直接写新服务
跑通后，再写对拍测试对比旧服务结果
```

---

## 9. AI 代码规范

### 9.1 写给 AI 的代码

| 原则 | 说明 |
|------|------|
| **类型优先** | 类型定义就是最好的文档，AI 读类型比读注释快 10 倍 |
| **接口清晰** | 函数签名精准，不要 `interface{}`，不要 `any` |
| **少写注释，多写类型** | 类型自文档化，注释只写"为什么"不写"是什么" |
| **模块化** | 每个文件只做一件事，AI 理解小文件更准 |
| **统一模式** | 所有 handler/service/repository 写法统一 |

### 9.2 不好的写法 ×

```go
// 处理用户请求
func process(data interface{}) interface{} {
    // 这里做了很多事情
    return data
}
```

### 9.3 好的写法 ✓

```go
func (h *UserHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    return h.service.CreateUser(ctx, req)
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    id, err := s.repo.Insert(ctx, &User{
        Username: req.Username,
        Password: hashPassword(req.Password),
    })
    if err != nil {
        return nil, err
    }
    return &pb.CreateUserResponse{UserId: id}, nil
}
```

---

## 10. 开发顺序

```
Phase 0（现在）：
  core 服务跑通 → 验证用户注册 + 登录 + 获取用户

Phase 1（下一步）：
  realtime 服务 → 弹幕发送 + 消息通知

Phase 2：
  search 服务 → 搜索 + 热榜

Phase 3：
  ads 服务 → 广告计数

Phase 4：
  media + store（Java GraalVM）
```