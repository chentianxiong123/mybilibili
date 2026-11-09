# MyBilibili Cloud 系统安装与使用说明书

## 第一章 系统概述

MyBilibili Cloud 是一个仿 B 站（哔哩哔哩）的微服务架构视频平台，基于 Spring Cloud + Vue 3 构建，涵盖视频投稿、弹幕、评论、私信、动态、搜索、AI 字幕/摘要等核心功能。

### 1.1 系统架构图

```
客户端（浏览器）
    │
    ▼
Nginx (:80) ─── 静态资源 + API 转发
    │
    ▼
Spring Cloud Gateway (:8080) ─── 统一入口，JWT 鉴权
    │
    ├── mybilibili-user     (:8081) ─── 用户/登录/关注/管理员
    ├── mybilibili-video    (:8082) ─── 视频/稿件/字幕/创作者数据
    ├── mybilibili-danmaku  (:8083) ─── 弹幕（MongoDB）
    ├── mybilibili-search   (:8084) ─── ES 全文搜索
    ├── mybilibili-comment (:8085) ─── 评论/举报/违禁词
    ├── mybilibili-interaction (:8086) ─── 点赞/投币/收藏/动态
    ├── mybilibili-message (:8087) ─── 私信/会话
    ├── mybilibili-ai      (:8088) ─── 视频转码/AI字幕/AI摘要
    └── mybilibili-live     (:8090) ─── 直播房间/连麦/WebRTC会议
```

### 1.2 核心技术栈

**后端：**
- Spring Boot 2.7.18 + Spring Cloud 2021.0.8
- Spring Cloud Alibaba（Nacos 注册/配置中心）
- Spring Cloud Gateway（API 网关）
- MyBatis Plus 3.5.3（ORM）
- MySQL 8.0 + Redis + MongoDB + Elasticsearch
- RocketMQ（消息队列，异步视频处理）

**前端：**
- Vue 3.4 + Vite 4.5
- Element Plus 2.7（UI 组件库）
- Artplayer 5.1 + HLS.js（视频播放器）
- ECharts 5.6（数据可视化）

---

## 第二章 系统功能说明

### 2.1 用户端功能

| 功能模块 | 说明 |
|---------|------|
| 用户注册/登录 | 用户名密码注册、登录、JWT 鉴权 |
| 视频播放 | HLS 多清晰度播放、弹幕展示、评论展示 |
| 视频投稿 | 视频上传、多 P 管理、封面设置、分类选择 |
| 弹幕系统 | 实时弹幕发送、展示 |
| 评论系统 | 多级评论/回复、违禁词自动过滤 |
| 互动系统 | 点赞、投币、收藏、分享、关注 |
| 个人中心 | 个人信息编辑、头像上传、视频列表 |
| 动态系统 | 发布图文动态、引用视频、好友互动 |
| 私信系统 | 一对一私信、会话管理、消息通知 |
| 搜索系统 | 全文搜索视频、用户 |
| 创作中心 | 稿件管理、数据统计、收藏夹/合集管理 |
| 直播系统 | 直播推流、直播房间、观众等级、定时开播提醒 |
| 会议系统 | WebRTC 视频会议、连麦、屏幕共享 |

### 2.2 管理端功能

| 功能模块 | 说明 |
|---------|------|
| 稿件审核 | 视频稿件审核、上下架管理 |
| 用户管理 | 用户列表、状态管理、密码重置 |
| 轮播图管理 | 首页轮播图配置 |
| 违禁词管理 | 敏感词过滤维护 |
| 数据统计 | 平台数据看板、趋势图表 |

---

## 第三章 系统安装与部署

### 3.1 环境依赖

| 依赖 | 版本 | 说明 | 端口 |
|------|------|------|------|
| JDK | 1.8+ | 后端运行环境 | - |
| Maven | 3.6+ | 后端编译构建 | - |
| Node.js | 16+ | 前端运行环境 | - |
| npm | 8+ | 前端包管理 | - |
| MySQL | 8.0+ | 关系型数据库 | 3306 |
| Redis | 6+ | 缓存/Session | 6379 |
| Nacos | 2.x | 服务注册/配置中心 | 8848 |
| RocketMQ | 4.9+ | 消息队列（视频异步处理流水线） | 9876 |
| Elasticsearch | 7.x | 全文搜索 | 9200 |
| MongoDB | 5+ | 弹幕存储 | 27017 |
| FFmpeg | 最新 | 视频 HLS 转码（可选） | - |
| Nginx | 1.20+ | 反向代理（生产环境） | 80 |

### 3.2 端口分配

| 服务 | 端口 | 依赖 |
|------|------|------|
| mybilibili-gateway | 8080 | Nacos |
| mybilibili-user | 8081 | MySQL, Redis, Nacos |
| mybilibili-video | 8082 | MySQL, Redis, MongoDB, Nacos, RocketMQ |
| mybilibili-danmaku | 8083 | MongoDB, Redis |
| mybilibili-search | 8084 | MySQL, Elasticsearch, Redis |
| mybilibili-comment | 8085 | MySQL, Redis, Nacos |
| mybilibili-interaction | 8086 | MySQL, Redis, Nacos |
| mybilibili-message | 8087 | MySQL, Redis, Nacos |
|| mybilibili-ai | 8088 | MySQL, MongoDB, Nacos, RocketMQ |
|| mybilibili-live | 8090 | MySQL, Redis, Nacos |
|| 用户端前端 | 5173 | - |
|| 管理端前端 | 3002 | - |

### 3.3 环境准备

#### 3.3.1 JDK 1.8
下载地址：https://adoptium.net/temurin/releases/?version=8

配置环境变量：
```bash
setx JAVA_HOME "C:\Program Files\Eclipse Adoptium\jdk-8.0.412.8-hotspot"
setx PATH "%PATH%;%JAVA_HOME%\bin"
```

#### 3.3.2 Maven
下载地址：https://maven.apache.org/download.cgi

项目根目录的 `settings.xml` 配置了阿里云镜像，复制到 `C:\Users\<用户名>\.m2\settings.xml` 可加速依赖下载。

#### 3.3.3 Node.js
下载地址：https://nodejs.org/zh-cn/（选择 LTS 版本）

#### 3.3.4 MySQL
安装 MySQL 8.0+，端口 3306，root 密码设为 `123456`。创建数据库：
```sql
CREATE DATABASE mybilibili CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```
执行数据库初始化脚本（`init/mybilibili-mysql.sql`）完成建表，需要 MongoDB 则执行 `init/mybilibili-mongodb.sql`。

#### 3.3.5 Redis
下载地址：https://github.com/tporadowski/redis/releases
```bash
redis-server.exe redis.windows.conf
```

#### 3.3.6 Nacos（单节点）
下载地址：https://github.com/alibaba/nacos/releases（选择 nacos-server-2.1.0.zip）
```bash
startup.cmd -m standalone
```
启动后访问 http://127.0.0.1:8848/nacos，默认账号密码 `nacos/nacos`。

#### 3.3.7 RocketMQ
```bash
docker run -d --name rocketmq -p 9876:9876 -p 10911:10911 -v d:/rocketmq/store:/root/store foshaned/rocketmq:4.9.4 ./mqnamesrv
docker run -d --name rocketmq-broker -p 10909:10909 -p 10911:10911 --link rocketmq -v d:/rocketmq/brokerstore:/root/store -e NAMESRV_ADDR=rocketmq:9876 foshaned/rocketmq:4.9.4 ./mqbroker -c ../conf/broker.conf
```

RocketMQ 是视频上传后异步处理的核心依赖：video 服务上传完成后发送 MQ 消息，ai 服务消费消息执行 HLS 转码、AI 字幕生成等步骤。

#### 3.3.8 Elasticsearch
```bash
docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "ES_JAVA_OPTS=-Xms512m -Xmx512m" elasticsearch:7.17.0
```

#### 3.3.9 MongoDB
```bash
docker run -d --name mongodb -p 27017:27017 -v d:/mongodb/data:/data/db mongo:5
```

#### 3.3.10 FFmpeg（可选）
下载地址：https://ffmpeg.org/download.html，将 `ffmpeg.exe` 加入系统 PATH。

### 3.4 后端部署

#### 3.4.1 编译项目
```bash
cd mybilibili-cloud
mvn clean install -DskipTests
```

#### 3.4.2 启动顺序

按以下顺序依次启动，每个服务成功注册到 Nacos 后再启动下一个：

1. **mybilibili-gateway**（端口 8080，必须先于其他服务启动）
2. **mybilibili-user**（端口 8081）
3. **mybilibili-video**（端口 8082）
4. **mybilibili-danmaku**（端口 8083）
5. **mybilibili-search**（端口 8084）
6. **mybilibili-comment**（端口 8085）
7. **mybilibili-interaction**（端口 8086）
8. **mybilibili-message**（端口 8087）
9. **mybilibili-ai**（端口 8088）

#### 3.4.3 启动方式

**方式一：IDEA**
在 IDEA 中打开项目，右键根目录 `pom.xml` → Maven → Reload Project，依次启动各模块的 `Application` 主类。

**方式二：jar 启动**
```bash
java -jar mybilibili-gateway/target/mybilibili-gateway-1.0.0.jar
# 其他服务同理
```

**方式三：一键脚本**
```bash
.\scripts\start-all.bat    # CMD
.\scripts\start-all.ps1    # PowerShell
```
脚本启动顺序：Gateway → User → Video → Danmaku → Search → Comment → Interaction → Message → AI → 用户端前端 → 管理端前端。

#### 3.4.4 配置文件说明

各服务的 `src/main/resources/application.yml` 中包含主要配置项：
```yaml
spring:
  datasource:
    url: jdbc:mysql://127.0.0.1:3306/mybilibili
    username: root
    password: ${MYSQL_PASSWORD:}
  redis:
    host: 127.0.0.1
    port: 6379
  cloud:
    nacos:
      discovery:
        server-addr: 127.0.0.1:8848
upload:
  base-path: d:/files/mybilibili/uploads
rocketmq:
  name-server: 127.0.0.1:9876
```

#### 3.4.5 文件上传目录

上传文件根目录为 `d:/files/mybilibili/uploads`，需提前创建：
```bash
mkdir d:\files\mybilibili\uploads\videos
mkdir d:\files\mybilibili\uploads\covers
mkdir d:\files\mybilibili\uploads\images
mkdir d:\files\mybilibili\uploads\subtitles
```

### 3.5 前端部署

#### 3.5.1 用户端
```bash
cd mybilibili-web
npm install
npm run dev
```
访问 http://localhost:5173

#### 3.5.2 管理端
```bash
cd mybilibili-admin-web
npm install
npm run dev
```
访问 http://localhost:3002

### 3.6 生产环境部署（Nginx）

#### 3.6.1 编译前端
```bash
cd mybilibili-web && npm install && npm run build
cd mybilibili-admin-web && npm install && npm run build
```
构建产物位于 `mybilibili-web/dist` 和 `mybilibili-admin-web/dist`。

#### 3.6.2 Nginx 配置
参考项目 `nginx/conf/nginx.conf`，关键配置：
```nginx
# 用户端静态资源
root d:/files/mybilibili-next/mybilibili-web/dist;

# 管理端
location /admin/ {
    alias d:/files/mybilibili-next/mybilibili-cloud/mybilibili-admin-web/dist/;
    try_files $uri $uri/ /admin/index.html;
}

# 上传文件访问
location /covers/ { alias d:/files/mybilibili/uploads/covers/; }
location /videos/ { alias d:/files/mybilibili/uploads/videos/; }
location /images/ { alias d:/files/mybilibili/uploads/images/; }

# API 代理
location /api/ {
    proxy_pass http://127.0.0.1:8080;
}
```

#### 3.6.3 启动 Nginx
```bash
nginx.exe -c D:\nginx\conf\nginx.conf
```

### 3.7 验证部署

#### 3.7.1 健康检查
- Gateway: http://localhost:8080
- Nacos: http://localhost:8848/nacos
- Elasticsearch: http://localhost:9200

#### 3.7.2 注册测试
```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test123456","email":"test@test.com"}'
```

#### 3.7.3 登录测试
```bash
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"Test123456"}'
```

---

## 第四章 系统使用指南

### 4.1 用户注册与登录

**注册：**
1. 打开用户端首页 http://localhost:5173
2. 点击右上角"注册"按钮
3. 输入用户名、密码、邮箱，点击"注册"
4. 注册成功自动跳转到首页并登录

**登录：**
1. 点击右上角"登录"按钮
2. 输入用户名和密码
3. 登录成功后页面右上角显示用户头像和昵称

### 4.2 视频播放

1. 首页展示推荐视频列表
2. 点击视频卡片进入视频详情页
3. 视频播放器支持：播放/暂停、进度拖拽、音量调节、全屏、画中画
4. 右侧显示弹幕列表，视频画面实时飘过弹幕
5. 下方显示评论区和视频详情

### 4.3 视频投稿

1. 点击页面顶部"投稿"按钮
2. 填写视频标题、简介、选择分类
3. 选择视频文件上传（支持 MP4、AVI 等常见格式）
4. 上传完成后后台自动进行 HLS 转码处理
5. 可设置封面、添加标签
6. 提交后进入审核流程，审核通过后在前台展示

### 4.4 弹幕功能

1. 在视频播放页面，右侧弹幕输入框输入内容
2. 点击"发送"即可在视频画面上发送弹幕飘过
3. 弹幕随机颜色显示，支持实时刷新
4. 弹幕内容会经过违禁词过滤

### 4.5 评论与回复

1. 在视频详情页下方评论区发表评论
2. 支持对其他用户的评论进行回复
3. 评论支持点赞
4. 用户可查看自己的评论记录

### 4.6 互动功能

**点赞：** 视频详情页点击点赞按钮，点赞数实时更新
**投币：** 点击投币按钮可为视频投币，每日投币有上限
**收藏：** 将视频添加至收藏夹，可创建和管理多个收藏夹
**关注：** 点击作者头像旁的关注按钮即可关注

### 4.7 个人中心

1. 点击右上角头像进入个人中心
2. **个人信息**：编辑昵称、签名、头像、性别、生日等
3. **视频列表**：查看用户投稿的所有视频
4. **收藏夹**：管理收藏夹及收藏的视频
5. **关注/粉丝**：查看关注列表和粉丝列表

### 4.8 动态系统

1. 在动态页发布图文动态
2. 动态可引用已投稿的视频
3. 关注用户的动态会出现在时间线
4. 支持动态评论互动

### 4.9 私信系统

1. 进入消息页面，查看会话列表
2. 点击用户发起私信对话
3. 支持发送文字消息
4. 未读消息有角标提醒

### 4.10 搜索功能

1. 首页顶部搜索框输入关键词
2. 支持搜索视频标题和用户
3. 搜索结果按相关性排序展示
4. 基于 Elasticsearch 全文搜索

### 4.11 创作中心

**入口：** 点击右上角头像 → "创作中心"

**功能：**
1. **稿件管理**：查看所有投稿视频，编辑信息、上下架管理
2. **数据统计**：查看视频播放量、点赞数、弹幕数等趋势图
3. **收藏夹管理**：创建编辑收藏夹
4. **合集管理**：将多个视频组合成合集，方便用户连续观看

### 4.12 管理后台

**入口：** 访问 http://localhost:3002

**功能：**
1. **数据看板**：平台用户数、稿件数、评论数等概览统计
2. **稿件审核**：管理员审核用户投稿，通过/拒绝操作
3. **用户管理**：查看用户列表、封禁/解封、重置密码
4. **轮播图管理**：配置首页轮播图片和跳转链接
5. **违禁词管理**：维护敏感词库，新增/删除违禁词
6. **内容审核**：审核评论内容

---

## 第五章 常见问题

### 5.1 服务无法注册到 Nacos
- 检查 Nacos 是否启动：http://localhost:8848/nacos
- 检查防火墙是否开放 8848 端口
- 检查各服务 `application.yml` 中 Nacos 地址配置

### 5.2 视频上传后不自动处理
- 检查 RocketMQ 是否启动，端口 9876
- 检查 ai 服务是否启动（8088）
- 查看 ai 服务日志是否有 "视频处理消息发送成功" 输出

### 5.3 前端请求 500 错误
- 检查对应后端服务是否启动，端口是否被占用
- 检查 MySQL、Redis 连接配置
- 查看后端日志排查具体错误

### 5.4 视频上传失败
- 检查 `d:/files/mybilibili/uploads` 目录是否存在
- 检查 Nginx 上传大小限制（默认 4096M）
- 检查 FFmpeg 是否安装并加入 PATH

### 5.5 搜索无结果
- 检查 Elasticsearch 是否启动
- 搜索依赖 ES 索引，需先有数据才会建立索引

### 5.6 AI 功能不可用
- AI 字幕需要 FFmpeg + Whisper 模型
- AI 摘要需要配置 DeepSeek API Key（`mybilibili-ai/application.yml`）
- AI 服务为可选，不影响视频上传、播放等核心功能

---

## 附录 项目目录结构

```
mybilibili-cloud/
├── mybilibili-gateway/          # API 网关（端口 8080）
├── mybilibili-user/             # 用户服务（端口 8081）
├── mybilibili-video/            # 视频服务（端口 8082）
├── mybilibili-comment/          # 评论服务（端口 8085）
├── mybilibili-danmaku/          # 弹幕服务（端口 8083）
├── mybilibili-interaction/      # 互动服务（端口 8086）
├── mybilibili-message/          # 消息服务（端口 8087）
├── mybilibili-search/           # 搜索服务（端口 8084）
├── mybilibili-ai/               # AI 服务（端口 8088）
├── mybilibili-live/              # 直播服务（端口 8090）
├── mybilibili-common/           # 公共模块（JWT、VO、工具类）
├── mybilibili-mq/               # 消息队列模块（RocketMQ 生产者）
├── mybilibili-web/               # 用户端前端（Vue 3）
├── mybilibili-admin-web/        # 管理端前端（Vue 3）
├── nginx/                       # Nginx 配置文件
├── init/                        # 数据库初始化 SQL（mybilibili-mysql.sql / mybilibili-mongodb.sql）
├── scripts/                     # 启动/停止/编译脚本（start-all.bat 等）
├── pom.xml                      # Maven 父 POM
└── settings.xml                  # Maven 镜像配置
```