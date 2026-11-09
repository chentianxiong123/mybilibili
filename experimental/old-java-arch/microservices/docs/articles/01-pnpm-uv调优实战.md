# pnpm + uv 调优实战：从 npm/conda 到极致省空间的包管理方案

> 发布于 CSDN · 2026-06-05  
> 关键词：pnpm, uv, npm, conda, 包管理, 空间优化, 微服务, 工程化

---

## 一、问题背景

我维护一个微服务架构的直播平台项目，包含 5 个前端服务（用户端 web、管理后台 admin、移动端 wap、剪辑工作室 studio-web、OBS 桌面端 live-desktop）和多个 Java 后端模块。

在一次仓库瘦身中，我把 2.16GB 的工作区砍到了 32MB 源码态。包管理优化是这次瘦身的关键一环——光是把 npm 换成 pnpm，前端依赖就省了 500MB+。

这篇文章记录完整的调优过程和踩坑经验。

---

## 二、npm 的空间浪费到底在哪

### 2.1 扁平化 + 全量复制

npm 从 v3 开始采用扁平化 node_modules 结构。每个项目执行 `npm install`，都会在自己的 `node_modules` 下**完整复制**一份依赖。

在我们的 monorepo 里，5 个前端项目各自跑一次 npm install：

| 项目 | node_modules 体积 |
|------|-------------------|
| mybilibili-web | ~180MB |
| mybilibili-admin | ~120MB |
| mybilibili-wap | ~95MB |
| mybilibili-studio-web | ~85MB |
| mybilibili-live-desktop | ~420MB |
| **合计** | **~900MB** |

其中 vue、echarts、vite、sass、element-plus 等核心依赖在 5 个项目里各存一份。实际去重后，唯一内容大概只有 ~200MB，剩下 ~700MB 都是重复。

### 2.2 幽灵依赖

npm 扁平化还会导致"幽灵依赖"——你的 package.json 里没声明的包，因为被提升到了顶层 node_modules，代码里 import 也能用。一旦上游依赖版本变了，幽灵依赖就可能消失，导致线上构建突然失败。

---

## 三、pnpm：硬链接 + 严格隔离

### 3.1 核心原理

pnpm 采用 **content-addressable store + 硬链接**：

1. 所有包的实际文件只存一份在全局 store（默认 `~/.pnpm-store`）
2. 每个项目的 `node_modules` 里是指向 store 的**硬链接**，不额外占磁盘空间
3. 依赖树**严格隔离**（非扁平化），彻底消灭幽灵依赖

### 3.2 实测效果

同一台机器，同一个项目，npm vs pnpm：

| 维度 | npm | pnpm |
|------|-----|------|
| 5 个前端总占用 | ~900MB | ~59MB + store ~320MB |
| 实际磁盘占用 | ~900MB | ~379MB |
| 节省 | — | **~58%** |
| 幽灵依赖风险 | 有 | 无 |

store 里的 320MB 是所有项目共享的，不管你装多少个前端项目，store 只增不增（除非引入新版本）。

### 3.3 关键配置

```bash
# 把 store 放到 D 盘，避免 C 盘爆满
pnpm config set store-dir "D:\.pnpm-store\v11"

# 清理 store 中不再被任何项目引用的包
pnpm store prune

# 查看 store 当前大小
pnpm store path
```

### 3.4 微服务场景下的 install 策略

很多人以为 pnpm 必须用 workspace，其实不是。我们的 5 个前端是**独立部署**的服务，各自有独立的端口、构建流程和部署目标。我们的策略是：

```
每个前端目录独立 pnpm-lock.yaml，独立 pnpm install
不使用根目录 pnpm-workspace.yaml
```

这样每个服务可以独立安装、独立升级依赖，互不影响。store 是共享的，但 lock 文件和 node_modules 是隔离的。

---

## 四、pnpm 11+ 的踩坑：ERR_PNPM_IGNORED_BUILDS

升级到 pnpm 11 后，很多人会遇到这个报错：

```
[ERR_PNPM_IGNORED_BUILDS] Ignored build scripts: esbuild@0.18.20, vue-demi@0.14.10
Run "pnpm approve-builds" to pick which dependencies should be allowed to run scripts.
```

这不是 bug，是 pnpm 11 引入的**安全机制**：默认禁止所有包执行 postinstall 脚本（防止供应链攻击）。

**解决方法**：在每个项目目录下执行交互式审批：

```bash
cd mybilibili-web
pnpm approve-builds
# 用空格勾选需要编译原生模块的包（esbuild、@parcel/watcher、vue-demi 等）
# 按回车确认
```

审批后会在 `package.json` 里生成 `pnpm.onlyBuiltDependencies` 字段，之后 install 不再报错。

**注意**：每个前端目录都要单独执行一次，因为审批是项目级别的。

---

## 五、conda 的空间问题

### 5.1 全量复制的环境

conda 每创建一个环境，就复制一份完整的 Python 解释器 + 依赖包：

```bash
conda create -n env1 python=3.11   # ~400MB
conda create -n env2 python=3.11  400MB
```

10 个环境轻松吃掉 4GB+，而且每个环境的 torch/transformers 还各存一份（~2.5GB/份）。

### 5.2 为什么不直接用 venv

venv 只是创建一个软链接到系统 Python，但 pip 安装的包仍然是每个环境独立复制的。所以 venv 比 conda 省了解释器的 400MB，但依赖包的重复问题没解决。

---

## 六、uv：Python 世界的 pnpm

### 6.1 核心原理

uv 是 Astral 出品的 Python 包管理器，用 Rust 写的，核心思路和 pnpm 一模一样——**全局缓存 + 硬链接**。

```bash
# 安装 uv
pip install uv

# 创建虚拟环境（秒级，因为 Python 解释器也走缓存）
uv venv .venv --python 3.11

# 安装依赖（比 pip 快 10-100 倍）
uv pip install -r requirements.txt

# 管理多个 Python 版本
uv python install 3.11 3.12
```

uv 的缓存目录默认在 `%LOCALAPPDATA%\uv\cache`（Windows），所有环境共享同一份已下载的包。

### 6.2 实测对比

| 操作 | conda | uv |
|------|-------|-----|
| 创建 Python 3.11 环境 | ~400MB, 30s | ~10MB, 1s |
| 安装 torch + transformers | ~2.5GB, 3min | ~2.5GB 缓存 + ~1MB 链接, 15s |
| 10 个重复环境总占用 | ~4GB | ~2.5GB（缓存共享） |

### 6.3 uv 的缓存管理

```bash
# 查看缓存目录
uv cache dir

# 清理不再被引用的缓存
uv cache prune

# 清理所有缓存
uv cache clean
```

---

## 七、统一方案：pnpm + uv 的工程实践

我们最终的项目结构：

```
mybilibili-cloud/
├── mybilibili-web/          # Vue 3 + pnpm, 端口 5173
├── mybilibili-admin/        # Vue 3 + pnpm, 端口 5174
├── mybilibili-wap/          # Vue 3 + pnpm, 端口 5175
├── mybilibili-studio-web/   # Vue 3 + pnpm, 端口 5176
├── mybilibili-live-desktop/ # Electron + pnpm
├── mybilibili-cloud/        # Java Maven（不用 pnpm）
└── scripts/                 # Python 脚本用 uv
```

每个前端独立 pnpm install，Python 工具链用 uv。两者共享同一个设计哲学：

> **全局缓存 + 硬链接 + 严格隔离 = 省空间 + 省时间 + 无幽灵依赖**

---

## 八、总结表

| 维度 | npm + conda | pnpm + uv |
|------|------------|-----------|
| 前端依赖空间 | ~900MB | ~379MB |
| Python 环境空间 | ~4GB/10 env | ~2.5GB/10 env |
| 安装速度 | 慢 | 快 5-100x |
| 幽灵依赖风险 | 有 | 无 |
| 磁盘回收 | 手动删目录 | store prune / cache clean |
| 供应链安全 | 无 | pnpm 11 approve-builds |

迁移成本很低：`npm install` 换成 `pnpm install`，`pip install` 换成 `uv pip install`，收益立竿见影。

---

> 原创不易，欢迎点赞收藏评论三连。有问题评论区交流，看到都会回。
