# Docker 基础设施

本项目开发期建议把基础设施放进 Docker，业务微服务和前端继续本地启动，方便断点、热重启和看日志。

所有持久化目录统一放在 `D:\DockerFiles\mybilibili`。

当前仓库只维护 `mybilibili-infra` 基础设施组。业务服务容器和前端 Nginx 镜像尚未落地，不在本文档承诺范围内。

## 组件

| 组件 | 镜像 | 端口 | 当前项目用途 |
| --- | --- | --- | --- |
| MySQL | `mysql:5.7` | `3306` | 主业务库 `mybilibili` |
| Redis | `redis:7.2-alpine` | `6379` | 缓存、计数、会话类数据 |
| Nacos | `nacos/nacos-server:v2.3.2` | `8848`, `9848`, `9849` | 服务发现 |
| MongoDB | `mongo:6.0` | `27017` | 弹幕、画像、推荐相关数据 |
| MinIO | `minio/minio:latest` | `9000`, `9001` | 对象存储 |
| RocketMQ | `apache/rocketmq:4.9.7` | `9876`, `10909`, `10911` | 视频处理消息 |
| Elasticsearch | `docker.elastic.co/elasticsearch/elasticsearch:7.17.18` | `9200`, `9300` | 搜索 |
| Sentinel Dashboard | `bladex/sentinel-dashboard:1.8.6` | `8858` | 网关限流观察 |
| SRS | `registry.cn-hangzhou.aliyuncs.com/ossrs/srs:5` | `1935`, `1985`, `28080`, `8000/udp` | 直播/拉流/推流 |

## 启动

先从 `.env.example` 复制一份 `.env`，填入 `MYSQL_ROOT_PASSWORD`、`MINIO_ROOT_USER`、`MINIO_ROOT_PASSWORD` 等真实值，真实密钥不进仓库。

```powershell
docker compose -f scripts/docker-compose-infra.yml up -d
```

当前代码里的连接地址基本都是 `127.0.0.1` 和固定端口，所以容器端口保持和本地配置一致，业务服务不用改配置即可连接。

`scripts/docker-compose-infra.yml` 会直接读取仓库内的 `scripts/rocketmq/broker.conf` 和根目录 `srs.conf`，不需要先把这两个配置文件复制到持久化目录。

业务微服务仍使用 `scripts/start-all.ps1` / `scripts/start-all.bat` 或 IDE 本地启动。前端仍在各自目录执行 `pnpm install`、`pnpm run dev`。

## FFmpeg / Whisper

本地运行 `mybilibili-ai` 需要宿主机安装 `ffmpeg` 并放入 `PATH`。Whisper 本地模式使用固定目录：

```text
D:\DockerFiles\mybilibili\whisper\bin\whisper-cli
D:\DockerFiles\mybilibili\whisper\models\ggml-small.bin
```

Windows 本地启动使用 Windows 版 `whisper-cli.exe`。也可以改用远程 Whisper/API Provider，此时本地 Whisper 目录不是硬依赖。

## 切换注意

本机如果已经有 MySQL 或 MongoDB，占用 `3306` / `27017`，需要先停掉本机服务，再启动 compose。

MySQL 切换前先备份：

```powershell
$env:MYSQL_PWD=$env:MYSQL_ROOT_PASSWORD
mysqldump -h 127.0.0.1 -P 3306 -u root --single-transaction --routines --triggers --events --default-character-set=utf8mb4 mybilibili > backups/mysql/mybilibili_before_docker.sql
Remove-Item Env:MYSQL_PWD
```

切到 Docker MySQL 后导入：

```powershell
Get-Content -Raw backups/mysql/mybilibili_before_docker.sql | docker exec -i mybilibili-mysql mysql -uroot -p$env:MYSQL_ROOT_PASSWORD --default-character-set=utf8mb4 mybilibili
```

MinIO 控制台：`http://127.0.0.1:9001`，账号密码使用启动时注入的 `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD`。

不要把所有东西塞成一个不可拆的大组，后续调试和迁移会很难收口。
