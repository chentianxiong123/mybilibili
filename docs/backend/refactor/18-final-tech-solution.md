# 最终技术方案：PG + Redis + SQLite 混合架构

## 1. 最终架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        每个业务实例（Pod/盒子）                        │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  SQLite mmap（L1） ← 只读缓存，mmap 零拷贝，~0.01ms           │  │
│  │  职责：扛 99% 读请求，不参与写入                                 │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                         读走 SQLite，写走 Redis
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        归档中心（老电脑/服务器）                       │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Redis（L2） ← 写入入口，扛并发写，透传归档到 PG                │  │
│  │  职责：                                                      │  │
│  │  ├── 写入缓冲区：所有写请求先到 Redis，高并发无压力              │  │
│  │  ├── 分布式锁：全局计数器、原子操作                             │  │
│  │  └── 异步刷 PG：Redis 异步将数据持久化到 PostgreSQL             │  │
│  ├──────────────────────────────────────────────────────────────┤  │
│  │  PostgreSQL（L3） ← 唯一真相源，归档层                         │  │
│  │  职责：                                                      │  │
│  │  ├── 关系表（替代 MySQL）                                     │  │
│  │  ├── JSONB + GIN 索引（替代 MongoDB）                         │  │
│  │  ├── tsvector 全文检索（替代 ES 基础功能）                      │  │
│  │  └── 从 Redis 异步刷入，不直接扛并发写                          │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. 读写分离：读走 SQLite，写走 Redis → PG

### 2.1 读路径（本地缓存，99% 不落网）

```
读请求 → SQLite（L1，本地 mmap，~0.01ms）
       → 命中？返回
       → 未命中 → PostgreSQL（L3，~5ms）
                → 回填 SQLite → 返回
```

**不经过 Redis 读，因为 SQLite 本地 mmap 比 Redis 网络请求快 50 倍。**

### 2.2 写路径（先 Redis，后 PG，异步持久化）

```
写请求 → Redis（L2，高并发写入，~0.5ms）
       → 返回成功给用户
       → 后台 Consumer 从 Redis 消费
       → 批量写入 PostgreSQL（削峰填谷）
       → 更新本地 SQLite 缓存（下次读命中）
```

### 2.3 为什么写走 Redis 而不是直接写 PG

| 方案 | 延迟 | 并发能力 | 数据安全 |
|------|------|---------|---------|
| 直接写 PG | ~5ms | 受 PG 连接池限制 | ✅ 强一致 |
| **先写 Redis，再刷 PG** | **~0.5ms** | **Redis 单机 10w QPS** | ✅ Redis AOF + 异步刷 PG |

### 2.4 强一致场景（支付/订单）

```
写请求 → 标记为「强一致」
       → 直接写 PostgreSQL（事务，~5ms）
       → 返回成功
       → 同时写入 Redis 缓存
       → 更新本地 SQLite
```

**抽象层保证：强一致操作直连 PG，非关键操作走 Redis → PG。**

---

## 3. PG 一库三用详解

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

## 4. 各功能模块实现方案

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

## 5. 组件清单

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

## 6. 智能路由：按资源 × 用户 × 地点

### 5.1 核心思想

不是"一人一核"，而是**智能调度 + 本地缓存**——一组实例共同服务所有用户，路由层根据三个维度动态调度：

```
用户请求
  ↓
路由层（Traefik + 自研调度器）
  ├── 按地点：就近接入，减少延迟
  ├── 按资源：选最空闲的实例，负载均衡
  └── 按用户：同用户尽量打同实例，保 SQLite 缓存命中
  ↓
实例池（Go + SQLite 本地缓存）
  ├── 实例A（华东，负载 30%）
  ├── 实例B（华东，负载 70%）
  ├── 实例C（华南，负载 20%）
  └── 实例D（华南，负载 50%）
  ↓
PG + Redis（归档中心）
```

### 5.2 三层路由策略

| 维度 | 策略 | 效果 |
|------|------|------|
| **地点** | 用户就近接入最近实例 | 延迟从 50ms 降到 5ms |
| **资源** | 路由到负载最低的实例 | 资源利用率从 30% 升到 80% |
| **用户** | 同用户长期绑定同实例（粘性） | SQLite 本地缓存命中率 > 95% |

### 5.3 为什么不是"一人一核"

| 方案 | 资源利用率 | 运维复杂度 | 适用规模 |
|------|-----------|-----------|---------|
| 一人一核 | 低（有用户才占资源） | 高（1000 个 Pod 管不过来） | 小 |
| **智能路由（推荐）** | **高（共享实例池）** | **低（一组实例统一管理）** | **大/小都行** |

**智能路由方案 = 粘性的好处（本地缓存命中率高）+ 池化的好处（资源利用率高）。**

### 5.4 资源计算

```
1000 用户，峰值并发 200
→ 10 个 Go 实例 × 20MB = 200MB（智能路由，共享实例池）
vs
→ 200 个 Go 实例 × 20MB = 4GB（一人一核，浪费）
```

**智能路由比一人一核省 20 倍资源。**

### 5.5 实例扩缩容

```
流量低（凌晨 3 点）：
  2 个实例运行，所有用户路由到这 2 个实例
  本地 SQLite 缓存命中率仍 > 90%（因为路由固定）

流量高（晚 8 点）：
  自动扩容到 10 个实例
  新实例从 PG 拉取热点数据预热 SQLite
  路由层逐步迁移用户，避免缓存雪崩

实例挂了：
  路由层自动把用户切到其他实例
  新实例从 PG 拉取数据，降级为无缓存模式
  后台同步重建 SQLite 缓存
```

## 7. 对比总结

| 维度 | 旧方案（MySQL+ES+Mongo） | 新方案（PG+Redis+SQLite） |
|------|------------------------|-------------------------|
| 归档层组件数 | 3 个 | **2 个（PG+Redis）** |
| 搜索能力 | ES 强但重 | PG 够用且轻 |
| 文档存储 | MongoDB | PG JSONB 差不多 |
| 运维复杂度 | 高 | **低** |
| 盒子友好 | 不友好 | **友好** |
| 从小到大的扩展 | 起步就要上全套 | **起步轻量，大了再切** |

---

## 8. 核心原则

1. **PG 一库三用**：关系表 + JSONB 文档 + 全文检索，一个数据库覆盖三个需求
2. **本地缓存优先**：内存 L0 → SQLite L1，挡掉 99% 读请求
3. **Redis 只做协调**：分布式锁、全局计数器，不做缓存主力
4. **抽象层兜底**：未来需要 ES/MongoDB，改配置不改代码
5. **从小到大的扩展**：一台盒子能跑，一百台机器能扛，中间不改架构