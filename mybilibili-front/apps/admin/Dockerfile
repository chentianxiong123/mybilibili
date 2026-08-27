FROM node:26-alpine AS build
RUN npm install -g pnpm@11
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

FROM nginx:alpine
COPY --from=build /app/apps/admin/dist /usr/share/nginx/html
COPY --from=build /app/apps/admin/nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 3100
CMD ["nginx", "-g", "daemon off;"]