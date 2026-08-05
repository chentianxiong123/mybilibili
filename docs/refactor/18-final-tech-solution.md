# 最终技术方案：PG + Redis + SQLite 混合架构

## 1. 最终架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        每个业务实例（Pod/盒子）                        │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  内存缓存（L0） ← 极热数据，零延迟，重启丢                       │  │
│  │  SQLite mmap（L1） ← 热数据，持久化，~0.01ms                    │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        归档中心（老电脑/服务器）                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Redis（L2） ← 分布式锁、全局计数器、热数据共享                   │  │
│  │  PostgreSQL（L3） ← 唯一真相源，一库三用                        │  │
│  │    ├── 关系表（替代 MySQL）                                     │  │
│  │    ├── JSONB + GIN 索引（替代 MongoDB）                         │  │
│  │    └── tsvector 全文检索（替代 ES 基础功能）                      │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. PG 一库三用详解

### 2.1 关系表（替代 MySQL）

```sql
-- 用户表，和 MySQL 一样
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    nickname VARCHAR(50),
    email VARCHAR(100),
    avatar TEXT,
    level INT DEFAULT 1,
    status INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 视频表
CREATE TABLE videos (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    video_url TEXT,
    cover_url TEXT,
    category_id BIGINT,
    view_count BIGINT DEFAULT 0,
    like_count BIGINT DEFAULT 0,
    comment_count BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### 2.2 JSONB 文档（替代 MongoDB）

```sql
-- 弹幕：JSONB 灵活存储
CREATE TABLE danmaku (
    id BIGSERIAL PRIMARY KEY,
    video_id BIGINT NOT NULL,
    data JSONB NOT NULL,  -- { content, timestamp, color, position, user_id }
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 查询：GIN 索引加速
CREATE INDEX idx_danmaku_video ON danmaku(video_id);
CREATE INDEX idx_danmaku_data ON danmaku USING GIN(data);

-- 弹幕查询
SELECT data->>'content' AS content,
       data->>'timestamp' AS ts
FROM danmaku
WHERE video_id = 1001
  AND (data->>'timestamp')::INT BETWEEN 1000 AND 2000;

-- 用户画像：JSONB 灵活存储
CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY,
    tags JSONB DEFAULT '[]',       -- ["游戏", "科技", "音乐"]
    behaviors JSONB DEFAULT '{}',  -- { "avg_watch_time": 120, "categories": [...] }
    preferences JSONB DEFAULT '{}',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 查询画像：GIN 索引加速
SELECT * FROM user_profiles
WHERE tags ? '游戏';
```

### 2.3 全文检索（替代 ES 基础功能）

```sql
-- 视频表加 tsvector 列
ALTER TABLE videos ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(description, ''))
    ) STORED;

-- GIN 索引加速搜索
CREATE INDEX idx_video_search ON videos USING GIN(search_vector);

-- 搜索视频
SELECT id, title, description,
       ts_rank(search_vector, query) AS rank
FROM videos, plainto_tsquery('simple', '猫') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 20;
```

**PG 全文检索的边界：**

| 功能 | PG 表现 | ES 表现 |
|------|--------|--------|
| 标题/描述搜索 | ✅ 一样快 | ✅ 一样快 |
| 中文分词 | ✅ 需要插件（zhparser） | ✅ 内置 |
| 相关度排序 | ✅ 够用 | ✅ 更强 |
| 聚合分析 | ⚠️ 能做但慢 | ✅ 快 10 倍 |
| 海量日志 | ❌ 扛不住 | ✅ 专长 |

**结论：100 人规模，PG 全文检索完全够用。到 10 万用户再考虑 ES。**

---

## 3. 各功能模块实现方案

### 3.1 弹幕

| 环节 | 实现 |
|------|------|
| 发送 | Go 接收 → NATS 广播 → 各实例实时收到 → MQ 异步写 PG |
| 存储 | PG JSONB，`video_id + data->>timestamp` 索引 |
| 拉取 | 播放器请求 → 本地 SQLite 命中返回 → 未命中查 PG 回填 |
| 实时 | WebSocket 直连，不经过 PG |

### 3.2 字幕

| 环节 | 实现 |
|------|------|
| 存储 | MinIO 存字幕文件，PG 存 `video_id → subtitle_url` 映射 |
| 检索 | SQLite 缓存热点字幕，PG 查映射 |
| 生成 | FFmpeg 提取 / Whisper 生成，走 MQ 任务队列 |

### 3.3 推荐

| 环节 | 实现 |
|------|------|
| 算法 | 简单规则：热度 × 时间衰减 + 分类匹配 + 协同过滤（基础版） |
| 计算 | Go 定时任务，PG 查数据 → 计算 → 结果写 Redis |
| 缓存 | 推荐结果缓存到本地 SQLite，定时刷新（~5 分钟） |
| 画像 | PG JSONB 存用户标签 + 行为，定时聚合 |

### 3.4 用户画像

| 环节 | 实现 |
|------|------|
| 采集 | 用户行为 → MQ → 异步写入 PG JSONB |
| 标签 | 定时任务聚合（Go 定时器），计算标签权重 |
| 查询 | 本地 SQLite 缓存画像数据，命中直接返回 |
| 更新 | 画像变更时失效本地缓存，下次请求重新拉取 |

### 3.5 搜索

| 环节 | 实现 |
|------|------|
| 视频搜索 | PG tsvector 全文检索，基础排序 |
| 用户搜索 | PG `ILIKE '%keyword%'`，用户名/昵称搜索 |
| 热榜 | PG 定时聚合，结果写 Redis，各实例从 Redis 拉 |
| 搜索建议 | 本地 SQLite 缓存热门搜索词，定时更新 |

---

## 4. 组件清单

| 组件 | 用途 | 内存 | 部署位置 |
|------|------|------|---------|
| **PostgreSQL** | 归档层，一库三用 | ~200MB | 老电脑 |
| **Redis** | 分布式锁、全局计数器 | ~50MB | 老电脑 |
| **NATS** | 消息队列 | ~20MB | 老电脑 |
| **etcd** | 服务发现 | ~50MB | 老电脑 |
| **MinIO** | 对象存储 | ~100MB | 主力机 |
| **Traefik** | 网关 | ~30MB | 盒子入口 |
| **Go 业务** | 业务逻辑 + SQLite 缓存 | ~30MB/实例 | 盒子 |
| **SQLite** | 本地热数据缓存（mmap） | 0（文件） | 每实例 |

**老电脑合计：~420MB，1GB 盒子：~60MB**

---

## 5. 对比总结

| 维度 | 旧方案（MySQL+ES+Mongo） | 新方案（PG+Redis+SQLite） |
|------|------------------------|-------------------------|
| 归档层组件数 | 3 个 | **2 个（PG+Redis）** |
| 搜索能力 | ES 强但重 | PG 够用且轻 |
| 文档存储 | MongoDB | PG JSONB 差不多 |
| 运维复杂度 | 高 | **低** |
| 盒子友好 | 不友好 | **友好** |
| 从小到大的扩展 | 起步就要上全套 | **起步轻量，大了再切** |

---

## 6. 核心原则

1. **PG 一库三用**：关系表 + JSONB 文档 + 全文检索，一个数据库覆盖三个需求
2. **本地缓存优先**：内存 L0 → SQLite L1，挡掉 99% 读请求
3. **Redis 只做协调**：分布式锁、全局计数器，不做缓存主力
4. **抽象层兜底**：未来需要 ES/MongoDB，改配置不改代码
5. **从小到大的扩展**：一台盒子能跑，一百台机器能扛，中间不改架构