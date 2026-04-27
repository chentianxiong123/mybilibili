# 用户接口微服务移植规格文档

## Why
原单体架构包含用户端(mybilibili-web)和管理员端(mybilibili-admin)两套用户相关接口，移植到微服务架构后，需要合并这些接口到用户服务(mybilibili-user)中，避免重复代码，实现统一管理。

## What Changes
- 将用户端和管理员端的用户接口合并到 mybilibili-user 服务
- 实体类迁移到 mybilibili-common 模块
- DTO/VO 类迁移到 mybilibili-common 模块
- UserMapper 迁移到 mybilibili-user 模块，使用 MyBatis-Plus
- UserService 重新实现用户注册、登录、查询、更新等功能
- Gateway 路由配置支持所有用户相关接口
- 添加密码加密支持（BCrypt）
- 添加头像上传功能（复用 UploadFilePathUtils）

## Impact
- Affected modules: mybilibili-common, mybilibili-user, mybilibili-gateway
- 数据库表: users

## 移植接口清单

### 用户端接口 (合并到 /user/**)
| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 用户注册 | POST | /user/register | 新用户注册 |
| 用户登录 | POST | /user/login | 返回Token和用户信息 |
| 获取用户信息 | GET | /user/{id} | 获取指定用户详情 |
| 更新用户信息 | PUT | /user/{id} | 更新用户个人资料 |
| 上传头像 | POST | /user/{id}/avatar | 上传用户头像 |

### 管理员接口 (合并到 /user/admin/**)
| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 获取用户列表 | GET | /user/admin/list | 分页获取用户列表 |
| 获取用户详情 | GET | /user/admin/{id} | 获取指定用户详情 |
| 更新用户状态 | PUT | /user/admin/{id}/status | 禁用/启用用户 |
| 重置用户密码 | PUT | /user/admin/{id}/password | 重置用户密码 |

## 原有代码分析

### 实体类 (User.java)
```
id, username, password, nickname, avatar, email, phone,
gender, birthdate(Date), signature, level, followingCount,
followerCount, videoCount, likedCount, coinCount, pointCount,
experience, bio, announcement, status, createdAt, updatedAt
```

### UserVO 字段
```
id, username, nickname, avatar, level, followingCount,
followerCount, dynamicCount, videoCount, experience, bio,
signature, announcement, email, phone, gender, birthdate,
status, totalViewCount, totalLikeCount
```

### UserDTO 字段
```
username, password, nickname, email
```

### LoginDTO 字段
```
username, password
```

### UserUpdateDTO 字段
```
nickname, avatar, email, phone, gender, birthdate,
signature, bio, announcement
```

## 技术要求
- 密码加密: BCryptPasswordEncoder
- JWT Token: 使用 common 模块中的 JwtUtils
- 文件上传: 复用 UploadFilePathUtils
- ORM: MyBatis-Plus
- 数据库: MySQL (mybilibili_cloud.users)
