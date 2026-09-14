# deploy/ — 生产部署配置（K3s）

`deploy/` 只放**生产部署**用的 manifest，**开发机编排已迁出到 `dev/docker-compose.yml`**。

| 目录 | 适用机器 | 说明 |
|------|---------|------|
| `k3s/` | **部署机**（fnos，192.168.31.225） | K3s 生产清单：kustomize 编排，镜像锁 git-sha |

## 开发机 → 见 `dev/README.md`（或直接 `dev/docker-compose.yml`）

```bash
docker compose -f dev/docker-compose.yml up -d --build
```

## 部署机 → k3s

```bash
kubectl apply -k deploy/k3s/overlays/prod
```

## 两者差异（关键点）

- 基础设施地址：compose 用容器名（`pg16`/`mybilibili-redis`/`mybilibili-minio`），k3s 用 Service 名（`postgres`/`redis`/`minio`）
- `TRANSCODER_ADDR`：compose 用 docker 网关 `172.18.0.1`，k3s 用部署机本机 IP `192.168.31.225`
- 镜像：compose 本地构建，k3s 从 GHCR 拉取（CI 自动推送）

## 统一配置源（两套部署同源）

非敏感配置按"环境差异"分散：

| 文件 | 在哪 | 被谁引用 |
|------|------|---------|
| `dev/_env/common.env` | dev/ | dev `env_file` |
| `dev/_env/dev.env` | dev/ | dev `env_file` |
| `deploy/k3s/base/config/common.env` | deploy/k3s/base/config/ | k3s `configMapGenerator` |
| `deploy/k3s/base/config/prod.env` | deploy/k3s/base/config/ | k3s `configMapGenerator` |

敏感项（JWT/MinIO 凭据）**不落配置文件**：
- compose 走 `dev/.env`（示例看 `dev/.env.example`，**待补**）
- k3s 走 `deploy/k3s/base/secret.yaml`

改配置只需编辑对应的 `.env` 文件：compose 改动后 `docker compose up -d` 自动生效；k3s 改动后 `kubectl apply -k deploy/k3s/overlays/prod` 重建 ConfigMap。