# 解决 RocketMQTemplate Bean 找不到的问题

## 问题分析

启动 mybilibili-ai 服务时报错：
```
Field rocketMQTemplate in com.mybilibili.mq.VideoMQProducer required a bean of type 'org.apache.rocketmq.spring.core.RocketMQTemplate' that could not be found.
```

**根本原因**：
1. `AiApplication.java` 中有 `@ComponentScan(basePackages = {"com.mybilibili.ai", "com.mybilibili.mq"})`
2. 这会扫描到 `mybilibili-mq` 模块中的 `VideoMQProducer` 类
3. `VideoMQProducer` 注入了 `RocketMQTemplate`
4. 但 `RocketMQTemplate` 的自动配置需要特定条件，当前未满足

**现状**：
- `VideoProcessConsumer` 已经改用 `DefaultMQProducer` 直接发送消息
- 不再依赖 `VideoMQProducer` 和 `RocketMQTemplate`
- `VideoProcessMessage` 和 `MQConstants` 是纯数据/常量类，不需要 Spring 管理

## 解决方案

移除 `AiApplication.java` 中对 `com.mybilibili.mq` 包的扫描。

## 实施步骤

### 步骤 1：修改 AiApplication.java

**文件**：`d:\files\mybilibili-next\mybilibili-cloud\mybilibili-ai\src\main\java\com\mybilibili\ai\AiApplication.java`

**修改内容**：
- 移除 `@ComponentScan` 注解（或从中移除 `com.mybilibili.mq`）

**修改前**：
```java
@SpringBootApplication
@EnableDiscoveryClient
@EnableFeignClients
@ComponentScan(basePackages = {"com.mybilibili.ai", "com.mybilibili.mq"})
public class AiApplication {
```

**修改后**：
```java
@SpringBootApplication
@EnableDiscoveryClient
@EnableFeignClients
public class AiApplication {
```

> 说明：`@SpringBootApplication` 默认会扫描当前包及其子包（`com.mybilibili.ai`），所以不需要额外的 `@ComponentScan`。`VideoProcessMessage` 和 `MQConstants` 是普通类，直接 import 使用即可，不需要 Spring 管理。

### 步骤 2：重新编译并启动服务

运行 Maven 编译，然后启动 mybilibili-ai 服务验证问题已解决。

## 验证方法

1. 编译项目：`mvn clean compile -pl mybilibili-ai -am`
2. 启动服务，确认无 `RocketMQTemplate` 相关错误
3. 检查日志中是否有 "RocketMQ Producer启动成功" 输出
