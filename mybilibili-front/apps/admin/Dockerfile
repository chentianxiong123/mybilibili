FROM node:26-alpine AS build
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app

# 复制 monorepo 清单 + workspace 配置
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages ./packages

# 复制 admin app
COPY apps/admin ./apps/admin

# 过滤安装（仅 admin 及其 workspace 依赖）
RUN pnpm install --frozen-lockfile --filter @mybilibili/admin --ignore-scripts

WORKDIR /app/apps/admin
RUN pnpm build

FROM node:26-alpine
WORKDIR /app/apps/admin
COPY --from=build /app/apps/admin/.output .output
EXPOSE 3100
ENV NITRO_PORT=3100
CMD ["node", ".output/server/index.mjs"]
