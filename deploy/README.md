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

## 开发机 compose 与部署机 k3s 的差异（关键点）

- 基础设施地址：compose 用容器名（`pg16-zhparser`/`mybilibili-redis`/`mybilibili-minio`/`nats`），k3s 用 Service 名（`postgres`/`redis`/`minio`/`nats`）
- `TRANSCODER_ADDR`：compose 用 docker 网关 `172.18.0.1`（`dev/_env/dev.env`），k3s 用部署机本机 IP `192.168.31.225`（`deploy/k3s/base/config/prod.env`）
- 镜像：compose 本地构建，k3s 从 GHCR 拉取（CI 自动推送）
- 数据卷：两者共用同一批 external 卷（`pg16-data`/`redis-data`/`minio-data`/`nats-data`/`studio-data`），**同一卷不可被两套编排同时挂载**

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

> ⚠️ 每个环境**只认上表中属于自己的那几个文件**，多余的 `.env` 不会被任何编排读取。
> 改到不生效的文件**不会报错**，只会静默无效——改前先确认文件在「被谁引用」列里。
>
> 字幕后端（Whisper）为可选配置，未设置时走占位 cue（见 `services/ai/internal/subtitle/generator.go`）：
> 可在对应环境的 env 里设置 `WHISPER_API_URL`（OpenAI 兼容端点）或 `CLOUDFLARE_AI_ACCOUNT_ID` + `CLOUDFLARE_AI_API_TOKEN`（Workers AI 直连）后启用。