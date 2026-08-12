# 前端统一与跨平台方案

## 1. 现状：实际盘点的 4 个项目

| 项目 | 技术栈 | 状态 | 决策 |
|------|--------|------|------|
| mybilibili-web | Vue 3 + Element Plus | 主站 SPA，**已含 admin 页面** | ✅ 核心，保留 |
| mybilibili-admin-web | Vue 3 + Element Plus | 与主站 admin 重复 | ❌ 删除 |
| mybilibili-wap | Vue 3（无 Element Plus） | 移动端精简版 | ✅ 作为 Capacitor 基底 |
| mybilibili-studio-web | Vue+React 混用 pnpm monorepo | 复杂，技术债深 | ⏸ 后续独立处理 |
| mybilibili-live-desktop | Electron | 已废弃，不看 | ❌ 删除 |

**关键发现**：主站 web 已经包含了 admin 目录（`src/views/admin/`），admin-web 是同一套东西做了两遍。

---

## 2. 统一方案：Monorepo + 共享包

```
mybilibili/
├── packages/
│   ├── shared/          # 共享类型、API 客户端、工具函数
│   │   ├── api/         # 从 proto 自动生成的 API 客户端
│   │   ├── types/       # 共享 TS 类型
│   │   └── utils/       # 工具函数
│   ├── ui/              # 共享 UI 组件库（TailwindCSS v4）
│   │   ├── Button/
│   │   ├── Modal/
│   │   ├── VideoPlayer/
│   │   ├── DanmakuOverlay/
│   │   ├── UserCard/
│   │   └── ...
│   └── composables/     # 共享 Vue composables（useAuth/useApi/useSSE）
├── apps/
│   ├── web/             # 主站 + 管理后台 + WAP（三合一）★核心★
│   │   ├── src/
│   │   │   ├── routes/
│   │   │   │   ├── main/      # 主站页面（原 mybilibili-web）
│   │   │   │   ├── admin/     # 管理后台（原 mybilibili-admin-web）
│   │   │   │   └── studio/    # 剪辑工作室（原 mybilibili-studio-web）
│   │   │   ├── layouts/
│   │   │   │   ├── MainLayout.vue
│   │   │   │   ├── AdminLayout.vue
│   │   │   │   └── StudioLayout.vue
│   │   │   └── App.vue
│   │   └── vite.config.ts
│   ├── mobile/           # 移动端（Capacitor 包装 Web）
│   │   ├── src/          # 共享 web 代码，差异化布局
│   │   ├── android/      # Capacitor Android 配置
│   │   └── capacitor.config.ts
│   └── desktop/          # 桌面端（Electron/Tauri）
│       └── src/
└── package.json
```

---

## 3. 合并策略

### 3.1 主站 + 管理后台 + WAP → 三合一（apps/web）

**为什么能合并：**
- 都是 Vue 3 + Vite，技术栈完全一致
- 管理后台只是路由前缀不同（`/admin/*`），不需要独立项目
- WAP 只是响应式断点不同，TailwindCSS 一套代码自适应

**怎么做：**
- 路由分模块：`/` 主站、`/admin/*` 后台、`/studio/*` 剪辑
- Layout 按路由切换：MainLayout（主站导航栏）、AdminLayout（侧边栏后台）、StudioLayout（全屏编辑）
- 组件全部走 `@mybilibili/ui` 共享包

### 3.2 剪辑工作室 → 并入 web（apps/web/routes/studio）

原 `mybilibili-studio-web` 是 pnpm monorepo，结构最复杂。但它的核心就是"视频剪辑编辑器"，本质是一个页面（`/studio/editor/:id`），不需要独立项目。

### 3.3 OBS 插件 → 独立不动

C++ 项目，跟前端无关，保留。

---

## 4. 跨平台方案：Vue 3 + Capacitor

### 4.1 为什么选 Capacitor

| 方案 | 代码复用 | 性能 | 维护成本 | 适合你的规模 |
|------|---------|------|---------|------------|
| **Capacitor**（推荐） | 95% | WebView 够用 | 低 | ✅ |
| Flutter | 0% | 原生 | 高（重写全部） | ❌ |
| React Native | 0% | 接近原生 | 高（重写全部） | ❌ |
| Tauri | 90% | 优于 WebView | 中（Rust 编译链） | ⚠️ |
| uni-app | 80% | 一般 | 中（多一层抽象） | ⚠️ |

**Capacitor = WebView 壳 + 原生 API 桥接。** 你的 Vue 3 代码直接跑在 WebView 里，安卓/iOS 不用重写。Camera、Push、File 等原生能力通过 Capacitor 插件暴露。

### 4.2 架构

```
apps/web (Vue 3 SPA) ← 核心，所有平台共享
    │
    ├── apps/mobile (Capacitor)
    │   └── android/  ← 自动生成，gradle 构建
    │   └── ios/      ← 自动生成（macOS 需要）
    │
    └── apps/desktop (Electron/Tauri)
        └── 共享 apps/web 的 dist 产物
```

### 4.3 移动端差异化

移动端和桌面端用**同一套 Web 代码**，通过响应式 + 平台检测做差异化：

| 差异点 | 做法 |
|--------|------|
| 布局 | TailwindCSS 响应式断点（`sm:` `md:` `lg:`) |
| 导航 | 移动端底部 Tab，桌面端侧边栏 |
| 视频播放 | 移动端全屏横屏，桌面端弹窗 |
| 推送通知 | Capacitor Push 插件 |
| 支付 | Capacitor In-App Purchase 插件 |

### 4.4 构建流程

```bash
# 开发
cd apps/web && npm run dev          # 浏览器热更新
cd apps/mobile && npx cap run android  # 真机调试

# 构建
cd apps/web && npm run build        # 产出 dist/
cd apps/mobile && npx cap sync       # 同步到 Android
cd apps/mobile && npx cap open android  # 打开 Android Studio 打包
```

---

## 5. 与后端的关系

```
                    ┌──────────────────┐
                    │  Core 胖单体      │
                    │  (含 go:embed)    │
                    └────────┬─────────┘
                             │
              ┌──────────────┴──────────────┐
              │              │              │
              ▼              ▼              ▼
         apps/web      apps/mobile     apps/desktop
         (浏览器)      (Capacitor)     (Electron)
              │              │              │
              └──────────────┴──────────────┘
                          │
                     proto 自动生成
                     @mybilibili/shared/api
```

- **Web 端**：Core 胖单体 go:embed 输出前端，Traefik 路由
- **移动端**：Capacitor 打包，API 调用走 HTTP（或 gRPC-Web）
- **桌面端**：Electron 壳，API 调用同上

---

## 6. 重构优先级

| 阶段 | 内容 | 产出 |
|------|------|------|
| P0 | 建立 monorepo 骨架 + shared/ 包 | 可编译的 workspace |
| P1 | web 三合一（主站 + admin + 路由合并） | 单项目，admin 不走独立构建 |
| P2 | 抽 UI 组件库（@mybilibili/ui，TailwindCSS v4） | 共享组件 |
| P3 | wap 接 Capacitor，打包 APK | 移动端可用 |
| P4 | studio-web 独立处理（后续） | 暂不涉及 |

---

## 7. 总结

1. **5 个前端项目 → 1 个 monorepo**：共享组件、共享 API、共享构建
2. **跨平台 = Capacitor**：Web 代码直接打包成 App，零重写
3. **合并路径**：admin/wap/studio 都是路由/布局差异，不需要独立项目
4. **你的 APK 不是 Android 原生**：是 WAP 的 WebView 壳，Capacitor 是更好的替代