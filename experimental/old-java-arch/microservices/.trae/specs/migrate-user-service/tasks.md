# 用户接口微服务移植任务清单

## 阶段一: 基础实体和工具类迁移

- [ ] 1. 更新 mybilibili-common 中的 User 实体类
  - [ ] 添加 MyBatis-Plus 注解
  - [ ] 确保字段与原实体一致
  - [ ] 添加 pointCount 字段

- [ ] 2. 更新 mybilibili-common 中的 UserVO
  - [ ] 确保包含所有需要返回给前端的字段

- [ ] 3. 更新 mybilibili-common 中的 UserDTO
  - [ ] 用户注册时使用

- [ ] 4. 更新 mybilibili-common 中的 LoginDTO
  - [ ] 登录参数

- [ ] 5. 更新 mybilibili-common 中的 UserUpdateDTO
  - [ ] 用户信息更新参数

- [ ] 6. 复制 UploadFilePathUtils 到 common 模块
  - [ ] 用于头像上传路径处理

## 阶段二: UserService 实现

- [ ] 7. 实现用户注册功能 (register)
  - [ ] 用户名唯一性检查
  - [ ] BCrypt 密码加密
  - [ ] 默认头像设置

- [ ] 8. 实现用户登录功能 (login)
  - [ ] 用户名密码验证
  - [ ] 生成 JWT Token
  - [ ] 返回 Token 和用户信息

- [ ] 9. 实现获取用户信息 (getUserById)
  - [ ] 基本信息查询
  - [ ] 统计数据查询（关注数、粉丝数等）
  - [ ] 稿件统计（播放数、点赞数）

- [ ] 10. 实现更新用户信息 (updateUser)
  - [ ] 字段更新
  - [ ] 日期格式处理

- [ ] 11. 实现头像上传 (uploadAvatar)
  - [ ] 文件验证（大小、格式）
  - [ ] 文件保存
  - [ ] URL 更新

## 阶段三: 管理员接口实现

- [ ] 12. 实现获取用户列表 (getUserList)
  - [ ] 分页查询
  - [ ] 关键词搜索
  - [ ] 返回总数和分页信息

- [ ] 13. 实现获取用户详情 (getAdminUserById)
  - [ ] 基本信息查询

- [ ] 14. 实现更新用户状态 (updateUserStatus)
  - [ ] 禁用/启用用户

- [ ] 15. 实现重置密码 (resetPassword)
  - [ ] BCrypt 加密新密码

## 阶段四: Controller 完善

- [ ] 16. 更新 UserController
  - [ ] 添加用户端接口
  - [ ] 添加管理员接口到 /admin 子路径
  - [ ] 完善异常处理

## 阶段五: 配置和测试

- [ ] 17. 确保 mybilibili-user 的 application.yml 配置正确
  - [ ] 数据库连接
  - [ ] Redis 连接
  - [ ] Nacos 配置

- [ ] 18. 添加 security 依赖到 mybilibili-user/pom.xml
  - [ ] spring-security-crypto

- [ ] 19. 编译测试
  - [ ] mvn clean compile
  - [ ] 修复编译错误

- [ ] 20. 启动服务测试
  - [ ] 启动 mybilibili-user
  - [ ] 测试各接口
