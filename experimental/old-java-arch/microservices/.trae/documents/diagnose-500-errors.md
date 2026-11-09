# 500 错误诊断计划

## 当前错误列表
所有接口都返回 500 Internal Server Error：
- `/api/category` - 分类接口 → 路由到 `lb://mybilibili-video`
- `/api/dynamic/list` - 动态接口 → 路由到 `lb://mybilibili-interaction`
- `/api/watch-history` - 历史记录接口 → 路由到 `lb://mybilibili-interaction`
- `/api/manuscript/favorite/folders` - 收藏夹接口 → 路由到 `lb://mybilibili-interaction`
- `/api/message/unread/counts` - 消息接口 → 路由到 `lb://mybilibili-message`

## 已确认正常的部分
1. **路由配置正确** - Gateway 的路由配置与各服务的接口路径匹配
2. **服务命名正确** - 各服务的 `spring.application.name` 与 Gateway 路由中的服务名匹配
3. **Gateway 依赖正确** - 包含了 `spring-cloud-starter-loadbalancer` 依赖

## 可能的问题原因

### 1. Nacos 服务发现问题
Gateway 使用 `lb://service-name` 进行负载均衡调用，需要通过 Nacos 发现服务实例。
- 如果服务没有注册到 Nacos，Gateway 会返回 500 错误
- 如果 Nacos 连接失败，服务发现会失败

### 2. 负载均衡器问题
Spring Cloud 2020+ 版本使用 Spring Cloud LoadBalancer 替代了 Ribbon。
- 需要确保 LoadBalancer 正确配置
- 需要确保服务实例正确注册

### 3. 服务内部错误
服务虽然启动了，但内部有错误导致请求处理失败。

## 诊断步骤

### 第一步：检查 Nacos 服务注册状态
请打开 Nacos 控制台 http://localhost:8848/nacos
- 用户名/密码：nacos/nacos
- 进入"服务管理" -> "服务列表"
- 确认以下服务是否都在列表中：
  - mybilibili-gateway
  - mybilibili-user
  - mybilibili-video
  - mybilibili-interaction
  - mybilibili-message

### 第二步：检查 Gateway 日志
当访问 `/api/category` 时，Gateway 控制台应该输出类似：
```
AuthFilter - Path: /api/category, UserId: xxx, Username: xxx
```
如果没有输出，说明请求没有到达 Gateway 或者 Gateway 有启动问题。

### 第三步：检查目标服务日志
当访问 `/api/category` 时，mybilibili-video 服务应该有日志输出。
如果没有输出，说明请求没有转发到该服务。

### 第四步：检查服务端口
确认各服务实际运行的端口：
- mybilibili-gateway: 8080
- mybilibili-user: 8081
- mybilibili-video: 8082
- mybilibili-interaction: 8086
- mybilibili-message: 8088

## 常见问题排查

### 问题1：服务未注册到 Nacos
**症状**：Nacos 控制台看不到服务
**原因**：
- Nacos 未启动
- 服务配置的 Nacos 地址错误
- 服务启动失败

**解决**：
1. 确保 Nacos 已启动
2. 检查各服务的 `spring.cloud.nacos.discovery.server-addr` 配置
3. 检查服务启动日志是否有错误

### 问题2：Gateway 无法发现服务
**症状**：Gateway 日志显示 "No instances available"
**原因**：
- 服务未注册到 Nacos
- Gateway 和服务不在同一个 Nacos 命名空间
- 网络问题

**解决**：
1. 确认服务已注册到 Nacos
2. 检查 Gateway 和服务的 Nacos 配置是否一致

### 问题3：服务内部错误
**症状**：请求到达服务但返回 500
**原因**：
- 数据库连接失败
- 代码有 bug
- 配置错误

**解决**：
1. 查看服务日志的具体错误堆栈
2. 检查数据库连接
3. 检查相关配置

## 请提供以下信息

1. **Nacos 控制台截图** - 服务列表页面
2. **Gateway 控制台日志** - 访问接口时的输出
3. **目标服务控制台日志** - 访问接口时的输出
4. **各服务的启动日志** - 是否有错误信息
