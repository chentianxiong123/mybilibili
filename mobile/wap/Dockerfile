# mobile/wap 移动端 SPA（Vite + Vue3）→ nginx 静态镜像
# 独立目录构建，主力 bun。构建上下文 = 仓库根。
# 注意：构建层必须干净（无宿主机 node_modules），bun 才能按当前平台
# 正确解析可选依赖；跨 libc（glibc↔musl）的 node_modules 混用会导致
# rollup 等原生二进制找不到对应平台包。
FROM oven/bun:1-alpine AS build
WORKDIR /app

# 先拷清单利用缓存
COPY mobile/wap/package.json mobile/wap/bun.lock ./
RUN bun install

# 拷源码构建
COPY mobile/wap ./wap-src
RUN mkdir -p /app/app && mv /app/wap-src /app/app/wap \
 && cd /app/app/wap && bun run build

FROM nginx:alpine
COPY --from=build /app/app/wap/dist /usr/share/nginx/html
COPY --from=build /app/app/wap/nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 5174
CMD ["nginx", "-g", "daemon off;"]