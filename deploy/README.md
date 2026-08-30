# deploy/ — 部署配置总览

本项目有两套部署方式，**按机器严格区分**：

| 目录 | 适用机器 | 说明 |
|------|---------|------|
| `docker-compose.yml` | **开发机**（工作站） | 一键全起：infra + 后端 8 服务 + 前端，traefik 统一入口 |
| `k3s/` | **部署机**（fnos，192.168.31.225） | K3s 生产清单：kustomize 编排，镜像锁 git-sha |

## 开发机 → docker compose

```bash
docker compose -f deploy/docker-compose.yml up -d --build
# 或一键脚本
scripts/backend.sh docker up
```

## 部署机 → k3s

```bash
kubectl apply -k deploy/k3s/overlays/prod
```

## 两者差异（关键点）

- 基础设施地址：compose 用容器名（`pg16`/`mybilibili-redis`/`mybilibili-minio`），k3s 用 Service 名（`postgres`/`redis`/`minio`）
- `TRANSCODER_ADDR`：compose 用 docker 网关 `172.18.0.1`，k3s 用部署机本机 IP `192.168.31.225`
- 镜像：compose 本地构建，k3s 从 GHCR 拉取（CI 自动推送）
