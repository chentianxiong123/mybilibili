# MyBilibili Java Backend

基于 Spring Boot 的类 B 站视频平台后端服务。

## 项目结构

```
mybilibili-java/
├── mybilibili-common/     # 公共模块 - 实体类、工具类、VO等
├── mybilibili-admin/      # 管理后台服务 - 管理员相关接口
├── mybilibili-web/        # 用户端服务 - 用户相关接口
└── pom.xml               # Maven 父项目配置
```

## 技术栈

- **Spring Boot 2.7.18** - 基础框架
- **MyBatis 1.3.2** - ORM 框架
- **MySQL 8.0** - 关系型数据库
- **Redis** - 缓存服务
- **Elasticsearch 7.17** - 搜索引擎
- **JWT** - 用户认证
- **Lombok** - 代码简化
- **Swagger/OpenAPI** - API 文档

## 功能模块

### mybilibili-common
公共模块，包含：
- 实体类 (Entity): User, Video, Comment, Danmaku, Follow, Like 等
- 数据传输对象 (DTO): LoginDTO, UserDTO 等
- 视图对象 (VO): VideoVO, UserVO, CommentVO 等
- 工具类: JwtUtils 等
- 配置: JwtFilter 等

### mybilibili-admin
管理后台服务，提供：
- 管理员登录认证
- 用户管理
- 视频稿件审核
- 评论管理
- 分类管理
- 字幕管理
- 轮播图管理
- 敏感词管理
- 数据统计

### mybilibili-web
用户端服务，提供：
- 用户注册/登录
- 视频上传/播放
- 弹幕系统
- 评论系统
- 点赞/投币/收藏
- 关注系统
- 消息通知
- 动态发布
- 搜索功能
- 个人中心

## 快速开始

### 环境要求

- JDK 1.8+
- Maven 3.6+
- MySQL 8.0+
- Redis 6.0+
- Elasticsearch 7.17+ (可选，用于搜索功能)

### 配置说明

1. 修改数据库配置 (application.yml):
```yaml
spring:
  datasource:
    url: jdbc:mysql://localhost:3306/mybilibili?useSSL=false&serverTimezone=Asia/Shanghai
    username: root
    password: your_password
```

2. 修改 Redis 配置:
```yaml
spring:
  redis:
    host: localhost
    port: 6379
```

### 运行项目

```bash
# 编译项目
mvn clean install

# 运行用户端服务
cd mybilibili-web
mvn spring-boot:run

# 运行管理后台服务
cd mybilibili-admin
mvn spring-boot:run
```

## API 文档

启动服务后访问:
- 用户端 API: http://localhost:8080/swagger-ui.html
- 管理端 API: http://localhost:8081/swagger-ui.html

## 相关项目

- [mybilibili-web](https://gitee.com/dllm7tou/mybilibili-web) - 用户端前端
- [mybilibili-admin-web](https://gitee.com/dllm7tou/mybilibili-admin-web) - 管理后台前端

## License

MIT
