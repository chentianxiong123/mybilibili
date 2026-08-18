FROM node:26-alpine AS build
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app

# 复制 monorepo 清单 + workspace 配置
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages ./packages

# 复制 web app
COPY apps/web ./apps/web

# 过滤安装（仅 web 及其 workspace 依赖）
RUN pnpm install --frozen-lockfile --filter @mybilibili/web --ignore-scripts

WORKDIR /app/apps/web
RUN pnpm build

FROM node:26-alpine
WORKDIR /app/apps/web
COPY --from=build /app/apps/web/.output .output
EXPOSE 3200
ENV NITRO_PORT=3200
CMD ["node", ".output/server/index.mjs"]
