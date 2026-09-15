FROM oven/bun:1-alpine AS build
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY packages ./packages
COPY apps/web ./apps/web

RUN bun install

WORKDIR /app/apps/web
RUN NITRO_PRESET=bun bun run build

WORKDIR /app/apps/web/.output/server
RUN bun install --production --no-cache && rm -rf /root/.bun \
 && rm -rf node_modules/@types node_modules/*/README* node_modules/*/LICENSE* \
 && find . -name '*.map' -delete

FROM alpine:3.21
COPY --from=oven/bun:1-alpine /usr/local/bin/bun /usr/local/bin/bun
RUN apk add --no-cache libgcc libstdc++ && chmod +x /usr/local/bin/bun
WORKDIR /app
COPY --from=build /app/apps/web/.output .output
EXPOSE 3200
ENV NITRO_PORT=3200
CMD ["bun", "run", ".output/server/index.mjs"]
