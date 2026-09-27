# web/user 主站（Nuxt SSR）→ bun 运行 .output
# 独立目录构建，主力 bun。构建上下文 = 仓库根。
FROM oven/bun:1-alpine AS build
WORKDIR /app

# 先拷清单 + 本地 workspace 依赖（ui）利用缓存
COPY web/user/package.json web/user/bun.lock ./
COPY web/user/packages ./packages
RUN bun install

# 拷源码构建
COPY web/user ./user-src
RUN mkdir -p /app/app && mv /app/user-src /app/app/web \
 && cd /app/app/web && NITRO_PRESET=bun bun run build

WORKDIR /app/app/web/.output/server
RUN bun install --production --no-cache && rm -rf /root/.bun \
 && rm -rf node_modules/@types node_modules/*/README* node_modules/*/LICENSE* \
 && find . -name '*.map' -delete

FROM alpine:3.21
COPY --from=oven/bun:1-alpine /usr/local/bin/bun /usr/local/bin/bun
RUN apk add --no-cache libgcc libstdc++ && chmod +x /usr/local/bin/bun
WORKDIR /app
COPY --from=build /app/app/web/.output .output
EXPOSE 3200
ENV NITRO_PORT=3200
CMD ["bun", "run", ".output/server/index.mjs"]