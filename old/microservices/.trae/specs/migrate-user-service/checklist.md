# 用户接口微服务移植检查清单

## 实体类检查
- [ ] User.java 字段完整，包含 pointCount
- [ ] User.java 使用 MyBatis-Plus 注解
- [ ] UserVO.java 包含所有返回字段
- [ ] UserDTO.java 包含注册所需字段
- [ ] LoginDTO.java 包含登录所需字段
- [ ] UserUpdateDTO.java 包含更新所需字段

## Service 层检查
- [ ] register 方法正确处理用户名重复
- [ ] login 方法使用 BCrypt 验证密码
- [ ] login 方法返回 JWT Token
- [ ] getUserById 返回完整统计数据
- [ ] updateUser 正确处理日期格式转换
- [ ] uploadAvatar 验证文件大小和格式
- [ ] getUserList 支持分页和搜索
- [ ] updateUserStatus 正确更新状态
- [ ] resetPassword 使用 BCrypt 加密

## Controller 层检查
- [ ] /user/register 注册接口正常
- [ ] /user/login 登录接口正常
- [ ] /user/{id} 获取用户信息正常
- [ ] /user/{id} PUT 更新用户信息正常
- [ ] /user/{id}/avatar POST 上传头像正常
- [ ] /user/admin/list GET 获取用户列表正常
- [ ] /user/admin/{id} GET 获取管理员用户详情正常
- [ ] /user/admin/{id}/status PUT 更新用户状态正常
- [ ] /user/admin/{id}/password PUT 重置密码正常

## 依赖检查
- [ ] pom.xml 包含 spring-security-crypto 依赖
- [ ] pom.xml 包含 mybatis-plus 依赖
- [ ] pom.xml 包含 redis 依赖

## 配置检查
- [ ] application.yml 数据库连接正确
- [ ] application.yml Redis 连接正确
- [ ] application.yml Nacos 配置正确
- [ ] 文件上传路径配置存在

## 编译检查
- [ ] mvn clean compile 成功
- [ ] 没有编译错误

## 运行检查
- [ ] mybilibili-user 服务启动成功
- [ ] 端口 8081 正常监听
- [ ] Nacos 注册成功
- [ ] 数据库连接正常
