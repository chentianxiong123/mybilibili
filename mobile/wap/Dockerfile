# wap 移动端 Single-Page App（vite）→ nginx 静态镜像
FROM node:26-alpine AS build
RUN npm install -g pnpm@11
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
# 无 TTY 环境下 pnpm 会拒绝清理 modules 目录，显式确认跳过
RUN pnpm install --frozen-lockfile --ignore-scripts --config.confirmModulesPurge=false

COPY . .
RUN pnpm build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf.template /etc/nginx/templates/default.conf.template
EXPOSE 5174
CMD ["nginx", "-g", "daemon off;"]