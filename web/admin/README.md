# mybilibili admin 前端

`web/admin` 管理后台（Vite + Vue 3 + Element Plus）。

## 安装依赖

主力包管理器是 **bun**（与 dev/docker-compose.yml 里的 `oven/bun` 容器一致）：

```bash
bun install
```

pnpm 作为可选可保留使用（锁文件覆盖 `bun.lock` + `pnpm-lock.yaml`），但常规开发请用 bun:

```bash
# 可选：pnpm
pnpm install
```

## 开发

```bash
bun run dev
```

## 测试

```bash
bun run test
```

## 构建

```bash
bun run build
bun run preview
```