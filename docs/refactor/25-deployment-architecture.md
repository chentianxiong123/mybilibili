# 部署架构：去 Nginx，Traefik 网关 + Go embed

## 1. 最终架构

```
浏览器/APP/桌面端
        │
        ▼
    Traefik（网关）
    │
    ├── / → Core 胖单体（go:embed 前端静态文件）
    │         ├── GET  /                → index.html（前端 SPA）
    │         ├── GET  /assets/*        → 前端 JS/CSS
    │         ├── POST /api/v1/*        → REST API（APP/桌面端也调这个）
    │         └── GET  /sse/*           → SSE 推送
    │
    ├── /ws/* → live 服务（语聊房信令，独立服务）
    │
    └── gRPC 内部 → Core/Live/Media 服务间通信
```

**Nginx 彻底下岗。** Traefik 负责网关层，Go embed 负责文件服务层，零额外进程。

---

## 2. 各端请求路径

| 端 | 入口 | 静态文件 | API | 区别 |
|----|------|---------|-----|------|
| 浏览器 | `https://mylib.com` | Traefik → Core embed（`/`） | Traefik → Core（`/api/v1/*`） | 一个域名全包 |
| 手机 APP | `https://mylib.com` | 无（APP 无浏览器） | 直接调 `POST /api/v1/*` | 不关心 embed |
| 桌面端 | `https://mylib.com` | 无（Electron 无浏览器壳） | 直接调 `POST /api/v1/*` | 不关心 embed |

**核心：embed 只影响浏览器端，APP 和桌面端完全无感，API 路径不变。**

---

## 3. 为什么不用 Nginx

| 对比 | Nginx | Traefik | Go embed |
|------|-------|---------|----------|
| 静态文件服务 | ✅ 强项 | ❌ 不做 | ✅ `embed.FS` 零进程 |
| HTTP 路由 | ✅ 要配 conf | ✅ 自动发现 | ❌ 不做 |
| gRPC 路由 | ❌ 不支持 | ✅ 原生 | ❌ 不做 |
| JWT 认证 | ❌ Lua/插件 | ✅ 中间件 | ❌ 不做 |
| 限流 | ❌ 插件 | ✅ 中间件 | ❌ 不做 |
| Let's Encrypt | ❌ certbot | ✅ 自动 | ❌ 不做 |
| 服务发现 | ❌ 手动 | ✅ etcd 对接 | ❌ 不做 |
| 进程数 | 额外 1 个 | 1 个 | 0（内嵌在 Go 进程） |

**Nginx 能做的，Traefik 做得更好（gRPC/自动证书/服务发现）；Nginx 拿手的静态文件，Go embed 零成本替代。**

---

## 4. Traefik 路由配置

```yaml
# traefik/dynamic.yml
http:
  routers:
    # 浏览器 → Core 胖单体（前端 + API）
    core:
      rule: "Host(`mylib.com`) && PathPrefix(`/`)"
      service: core
      middlewares:
        - jwt-auth   # JWT 认证（/api/v1/* 需要，/ 不需要）
        - rate-limit # 限流

    # 直播信令 WS → live 服务
    live-ws:
      rule: "Host(`mylib.com`) && PathPrefix(`/ws/`)"
      service: live
      middlewares:
        - jwt-auth
        - ws-timeout # WS 超时设为 0

  services:
    core:
      loadBalancer:
        servers:
          - url: "http://localhost:8080"  # Go 胖单体

    live:
      loadBalancer:
        servers:
          - url: "http://localhost:8081"  # live 独立服务
```

**Go 胖单体只需要在 `:8080` 上监听 HTTP，Traefik 负责 HTTPS、路由、认证、限流。**

---

## 5. Go embed 实现要点

```go
// cmd/core/main.go
package main

import (
    "embed"
    "net/http"
)

//go:embed web/dist/*
var staticFS embed.FS

func main() {
    mux := http.NewServeMux()

    // 前端静态文件（浏览器访问 /）
    mux.Handle("/", http.FileServer(http.FS(staticFS)))

    // API（浏览器/APP/桌面端都调）
    mux.Handle("/api/v1/", apiHandler())

    // SSE 推送
    mux.Handle("/sse/", sseHandler())

    http.ListenAndServe(":8080", mux)
}
```

**目录结构：**
```
web/
├── dist/             ← vite build 产物，go:embed 目标
│   ├── index.html
│   └── assets/
├── src/              ← Vue 3 源码
├── vite.config.ts
└── package.json
```

**构建流程：**
```bash
cd web && npm run build    # 产出 dist/
cd .. && go build ./cmd/core  # embed 进二进制
```

---

## 6. 总结

1. **Nginx 下岗**：静态文件 → Go embed，网关 → Traefik
2. **Traefik 统一网关**：路由/JWT/限流/证书/服务发现，一个搞定
3. **Go embed 零成本**：10KB 前端嵌入 15MB 二进制，体积增加 0.06%
4. **APP/桌面端无感**：只调 `/api/v1/*`，不关心 embed 存在
5. **部署最简**：一个二进制 + Traefik，盒子 1GB 轻松跑