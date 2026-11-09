# MyBilibili Admin Web

基于 Vue 3 的类 B 站视频平台管理后台前端。

## 技术栈

- **Vue 3.3** - 前端框架
- **Vite 4.4** - 构建工具
- **Vue Router 4** - 路由管理
- **Pinia 2** - 状态管理
- **Element Plus 2.3** - UI 组件库
- **Axios** - HTTP 请求

## 功能模块

### 仪表盘
- 数据统计概览
- 用户增长趋势
- 视频发布统计

### 用户管理
- 用户列表
- 用户状态管理
- 用户信息查看

### 管理员管理
- 管理员列表
- 权限分配
- 角色管理

### 稿件管理
- 视频稿件列表
- 稿件审核
- 稿件状态管理

### 视频处理
- 视频转码状态
- 处理进度监控

### 评论管理
- 评论列表
- 评论审核
- 敏感内容过滤

### 分类管理
- 视频分类
- 分类增删改查

### 字幕管理
- 字幕列表
- 字幕审核

### 轮播图管理
- 首页轮播图配置

### 敏感词管理
- 敏感词库
- 敏感词增删改查

## 项目结构

```
mybilibili-admin-web/
├── src/
│   ├── api/           # API 接口
│   ├── router/        # 路由配置
│   ├── stores/        # Pinia 状态
│   ├── views/         # 页面组件
│   ├── App.vue        # 根组件
│   └── main.js        # 入口文件
├── public/            # 静态资源
├── dist/              # 构建输出
└── package.json       # 项目配置
```

## 快速开始

### 环境要求

- Node.js 16.0+
- npm 或 pnpm

### 安装依赖

```bash
npm install
```

### 开发运行

```bash
npm run dev
```

### 生产构建

```bash
npm run build
```

## 页面路由

| 路径 | 页面 | 说明 |
|------|------|------|
| `/login` | LoginView | 登录页 |
| `/dashboard` | DashboardView | 仪表盘 |
| `/users` | UsersView | 用户管理 |
| `/admins` | AdminsView | 管理员管理 |
| `/roles` | RolesView | 角色管理 |
| `/manuscripts` | ManuscriptsView | 稿件管理 |
| `/video-process` | VideoProcessView | 视频处理 |
| `/comments` | CommentsView | 评论管理 |
| `/categories` | CategoriesView | 分类管理 |
| `/subtitle` | SubtitleManagementView | 字幕管理 |
| `/banner-images` | BannerImagesView | 轮播图管理 |
| `/prohibited-words` | ProhibitedWordsView | 敏感词管理 |

## 相关项目

- [mybilibili](https://gitee.com/dllm7tou/mybilibili) - Java 后端服务
- [mybilibili-web](https://gitee.com/dllm7tou/mybilibili-web) - 用户端前端

## License

MIT
