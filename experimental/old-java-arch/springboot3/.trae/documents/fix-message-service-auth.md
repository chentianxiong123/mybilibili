# 消息服务三层匹配逻辑修复计划

## 问题分析

### 当前问题
用户访问 `http://localhost:5173/message/private` 时出现500错误。

### 根本原因
经过分析，问题有两个层面：

#### 1. Gateway 认证层面
- `/api/message/` 路径需要用户认证（需要 `X-User-Id` 头）
- 当前 AuthFilter.java 中没有将 `/api/message/` 相关路径添加到 `AUTH_REQUIRED_PATHS`
- 这可能导致消息接口被错误地放入白名单，跳过认证

#### 2. 数据库层面（主要问题）
- **MessageMapper.xml** 中多处使用了 `status = 0` 条件
- **数据库 `messages` 表实际没有 `status` 列**
- 这导致所有涉及 `status` 的 SQL 查询都会报错：`Unknown column 'status' in 'where clause'`

### 数据库实际结构
```
messages 表字段：
id, sender_id, receiver_id, content, is_read, created_at, updated_at, 
message_type, target_id, media_url, conversation_id
（没有 status 字段）
```

### MessageMapper.xml 中需要修复的位置
1. 第18行：resultMap 中的 `<result column="status" property="status"/>`
2. 第38-39行：INSERT 语句中的 `status` 字段
3. 第59行：`AND m.status = 0`
4. 第69行：`AND m.status = 0`
5. 第81行：`AND m.status = 0`
6. 第106行：`AND status = 0`
7. 第129行：`AND m.status = 0`
8. 第135行：`AND status = 0`
9. 第158行：`AND m.status = 0`
10. 第164行：`AND status = 0`
11. 第187行：`AND m.status = 0`
12. 第193行：`AND status = 0`
13. 第209行：`AND m.status = 0`
14. 第215行：`AND status = 0`

---

## 修复计划

### 步骤1：修复 Gateway AuthFilter.java
将消息相关路径添加到 `AUTH_REQUIRED_PATHS`：
```java
"/api/message/conversations",
"/api/message/send",
"/api/message/unread",
"/api/message/replies",
"/api/message/at",
"/api/message/likes",
"/api/message/system",
"/api/message/settings"
```

### 步骤2：修复 MessageMapper.xml
移除所有 `status` 相关的代码：
- 从 resultMap 中移除 `status` 映射
- 从 INSERT 语句中移除 `status` 字段
- 从所有 SELECT 查询中移除 `AND status = 0` 条件

### 步骤3：检查 Message 实体类
确保 `Message.java` 实体类中没有 `status` 字段（如果有则移除）

### 步骤4：重启服务并测试
- 重启 gateway 服务
- 重启 message 服务
- 测试消息相关接口

---

## 预期结果
- 消息服务接口正常返回数据
- 前端 `/message/private` 页面正常显示
- 用户可以查看会话列表、发送消息、查看通知等
