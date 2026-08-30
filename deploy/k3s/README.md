# deploy/k3s/ — K3s 部署清单

## 结构

```
k3s/
├── base/                        # 公共资源（不含 infra）
│   ├── kustomization.yaml
│   ├── namespace.yaml           # mybilibili 命名空间
│   ├── configmap.yaml           # 非敏感配置
│   ├── secret.yaml             # 敏感项（JWT/MinIO密钥）
│   ├── backend.yaml             # 8 后端 Deployment
│   ├── frontend.yaml            # web(Nuxt SSR) + admin(nginx SPA)
│   ├── pvc.yaml                 # studio-data / work-tmp
│   └── service.yaml             # ClusterIP Service
└── overlays/
    ├── dev/                      # 开发：自带 infra(pg/redis/minio/nats)，单副本
    │   ├── kustomization.yaml
    │   └── infra.yaml            # pg16/redis/minio/nats Deployment+Service+PVC
    └── prod/                     # 生产：连开发机 infra，不部署自己的
        └── kustomization.yaml    # ConfigMap 指向 192.168.31.204
```

## 两种部署模式

### dev：全套自包含（k3s 自己跑 infra）
```bash
kubectl apply -k deploy/k3s/overlays/dev
```
适用于：k3s 独立环境，自带 PG/Redis/MinIO/NATS（空库，需迁移数据）。

### prod：连开发机 infra（推荐两机分离架构）
```bash
# 1. 先建 namespace + ghcr 拉取凭证
kubectl apply -f deploy/k3s/base/namespace.yaml
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=<gh-user> \
  --docker-password=<gh-token> \
  -n mybilibili

# 2. 部署（ConfigMap 自动指向开发机 IP 192.168.31.204）
kubectl apply -k deploy/k3s/overlays/prod
```
适用于：部署机连开发机的 PG/MinIO（有数据），不重复部署 infra。
改开发机 IP：编辑 `overlays/prod/kustomization.yaml` 里的 ConfigMap patch。

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

## NodePort 端口

| 服务 | NodePort | 宿主机访问 |
|---|---|---|
| core | 30080 | http://<node-ip>:30080 |
| web | 30320 | http://<node-ip>:30320 |
| admin | 30310 | http://<node-ip>:30310/admin/ |

## 镜像来源

CI（`.github/workflows/ci.yml`）push main 自动构建推 GHCR：
```
ghcr.io/chentianxiong123/mybilibili-{core,search,msg-danmaku,live,ai,studio,work,bili,web,admin}:latest
```