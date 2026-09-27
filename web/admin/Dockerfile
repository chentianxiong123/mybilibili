# web/admin 管理后台（Vite + Vue3 + Element Plus）→ nginx 静态镜像
# 独立目录构建，主力 bun。构建上下文 = 仓库根。
# 注意：构建层必须干净（无宿主机 node_modules），bun 才能按当前平台
# 正确解析可选依赖；跨 libc（glibc↔musl）的 node_modules 混用会导致
# rollup 等原生二进制找不到对应平台包。
FROM oven/bun:1-alpine AS build
WORKDIR /app

# 先拷清单 + 本地 workspace 依赖（@mybilibili/ui 实际位于 web/user/packages/ui，
# 与 dev/docker-compose 的挂载保持一致）利用缓存
COPY web/admin/package.json web/admin/bun.lock ./
COPY web/user/packages/ui ./packages/ui
RUN bun install

# 拷源码构建
COPY web/admin ./admin-src
RUN mkdir -p /app/app && mv /app/admin-src /app/app/admin \
 && cd /app/app/admin && bun run build

FROM nginx:alpine
COPY --from=build /app/app/admin/dist /usr/share/nginx/html
COPY --from=build /app/app/admin/nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 3100
CMD ["nginx", "-g", "daemon off;"]