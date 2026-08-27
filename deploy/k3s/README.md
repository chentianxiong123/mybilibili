# deploy/k3s/ — K3s 部署预留目录

下一阶段（CI/CD + k3s）在此目录填充 Kustomize 清单。

## 计划结构

```
k3s/
├── base/
│   ├── kustomization.yaml
│   ├── namespace.yaml          # mybilibili 命名空间
│   ├── configmap.yaml          # 从 deploy/.env 非敏感项生成
│   ├── secret.yaml             # 敏感项 (JWT_SECRET/MINIO密钥) → SealedSecret
│   ├── deployment-core.yaml    # 8 个后端 Deployment
│   ├── deployment-work.yaml    # work 带 GPU (intel device plugin)
│   ├── service.yaml            # ClusterIP (内部) / NodePort (前端访问)
│   └── pvc.yaml                # studio-data / work-tmp 持久卷
└── overlays/
    ├── dev/                    # 开发: 单副本, NodePort
    └── prod/                   # 生产: 多副本, 资源限制, HPA
```

## 从 compose 到 k3s 的映射

| docker-compose | k3s/k8s |
|---------------|---------|
| 服务 + 端口映射 | Service + Deployment |
| env_file (.env) | ConfigMap + Secret |
| volumes (studio-data) | PersistentVolumeClaim |
| devices (/dev/dri) | intel GPU device plugin / `resources.limits` |
| container_name DNS | Service 名 (core.mybilibili.svc) |
| restart: unless-stopped | Deployment restartPolicy |
| healthcheck | livenessProbe / readinessProbe |
| ghcr.io image:latest | image: ghcr.io/.../core@sha256:xxx (用 sha tag) |

## 镜像来源

CI (`.github/workflows/ci.yml`) 构建推送到 GHCR:
```
ghcr.io/chentianxiong123/mybilibili-{core,search,msg-danmaku,live,ai,studio,work,bili}:<git-sha>
ghcr.io/chentianxiong123/mybilibili-<service>:latest
```
k3s 拉取: `kubectl create secret docker-registry ghcr-secret --docker-server=ghcr.io ...`
