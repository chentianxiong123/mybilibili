# deploy/k3s/ — 部署机（fnos）K3s 部署清单

> **本目录只用于部署机（fnos，192.168.31.225）**。开发机用 `deploy/docker-compose.yml`。
> 两套配置互相独立：k3s 管生产，compose 管开发。

## 结构（以实际仓库为准）

```
k3s/
├── traefik.yaml                  # Traefik CRD + IngressRoute 路由定义
└── base/
    ├── kustomization.yaml        # 资源编排入口（含 configMapGenerator 生成 mybilibili-config）
    ├── namespace.yaml            # mybilibili 命名空间
    ├── infra.yaml                # 基础设施：postgres/redis/minio/nats (Deployment+Service+PVC)
    ├── config/                   # ★ 统一配置源（与 docker compose 同源）
    │   ├── common.env            #   两机一致：gRPC 地址 / 目录 / 日志 / MQ 类型
    │   ├── dev.env               #   开发机差异（compose 用：容器名 + 网关 IP）
    │   └── prod.env              #   部署机差异（k3s 用：Service 名 + 部署机 IP）
    ├── secret.yaml               # 敏感项（JWT/MinIO 凭据，不落 ConfigMap）
    ├── backend.yaml              # 8 个后端 Deployment (core/search/msg-danmaku/live/ai/studio/work/bili)
    ├── frontend.yaml             # 前端 Deployment (web + admin)
    ├── pvc.yaml                  # studio-data / work-tmp 持久卷
    ├── service.yaml              # ClusterIP Service
    └── ingress.yaml              # 入口规则
└── overlays/
    └── prod/
        └── kustomization.yaml    # product overlay：只覆盖镜像 tag（锁 git-sha）
```

**注意**：
- `base/` 已包含全部基础设施（infra.yaml），prod overlay 不再重复定义 infra。
- 非敏感配置由 `kustomization.yaml` 的 `configMapGenerator` 从 `config/common.env + config/prod.env` 生成（`kubectl apply -k` 时自动变为 `mybilibili-config`）。
- 敏感项（JWT/MinIO 凭据）只在 `secret.yaml`，compose 侧则走 `deploy/.env`，两边均不写死到配置清单。

## 部署（在部署机 fnos 上执行）

```bash
# 1. 准备 ghcr 拉取凭证
kubectl apply -f deploy/k3s/base/namespace.yaml
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<gh-user> \
  --docker-password=<gh-token> \
  -n mybilibili

# 2. 应用清单（prod overlay 锁镜像 tag，见下方升级流程）
kubectl apply -k deploy/k3s/overlays/prod
```

## 镜像与升级

- CI（`.github/workflows/ci.yml`）push main 自动构建推 GHCR：
  `ghcr.io/chentianxiong123/mybilibili-{core,search,msg-danmaku,live,ai,studio,work,bili,web,admin}:latest`（另推 git-sha tag）
- 部署清单用 `overlays/prod/kustomization.yaml` 的 `images.newTag` 锁版本。
- **升级**：改 `newTag` 为新 git-sha → `kubectl apply -k deploy/k3s/overlays/prod`
- **回滚**：改 `newTag` 回旧 git-sha → 重新 apply
- **改配置**：编辑 `base/config/common.env` 或 `base/config/prod.env` → `kubectl apply -k deploy/k3s/overlays/prod`（configMapGenerator 会自动重建 ConfigMap）
- TRANSCODER_ADDR 差异：部署机（k3s）用 `base/config/prod.env` 的 `192.168.31.225`；开发机（compose）用 `base/config/dev.env` 的 `172.18.0.1`（docker 网关），两处都已抽为变量。

## transcoder 裸跑（不在 k3s 内）

transcoder 用系统 ffmpeg + 显卡驱动，按 GPU 类型编译：
```bash
# AMD/Intel (VAAPI)
make build-transcoder-vaapi    # → /tmp/mybilibili-transcoder-vaapi

# NVIDIA (NVENC)
make build-transcoder-nvenc    # → /tmp/mybilibili-transcoder-nvenc

# 软编（无 GPU）
make build-transcoder-soft     # → /tmp/mybilibili-transcoder
```
哪台机器把对应版本放到 `/tmp/mybilibili-transcoder` 启动即可。