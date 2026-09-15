# cf-whisper-worker

Cloudflare Workers AI Whisper 转写代理。部署到 Cloudflare Workers，通过 Workers AI 调用 whisper 模型。

## 架构

```
业务服务 (AI) → POST /api/v1/transcribe → cf-whisper-worker → Cloudflare Workers AI → whisper
```

**被调用方**：不关心业务逻辑，只接收音频数据，返回转写结果。

## 部署步骤

### 1. 安装 Wrangler CLI

```bash
npm install -g wrangler

# 登录 Cloudflare
wrangler login

# 验证
wrangler whoami
```

### 2. 配置 wrangler.toml

已有默认配置，通常不需要修改：

```toml
name = "cf-whisper-worker"
main = "src/index.ts"
compatibility_date = "2026-07-01"

[ai]
binding = "AI"

[vars]
WHISPER_MODEL = "@cf/openai/whisper"
```

### 3. 部署到 Cloudflare

```bash
cd external/cf-whisper-worker
wrangler deploy

# 输出类似：
# ✨ Successfully published your script to https://cf-whisper-worker.<你的子域>.workers.dev
```

### 4. 记录部署地址

部署后会得到一个 URL，格式：
```
https://cf-whisper-worker.<你的子域>.workers.dev
```

将此 URL 配置到 AI 服务的环境变量中：

```env
WHISPER_API_URL=https://cf-whisper-worker.<你的子域>.workers.dev/api/v1/transcribe
```

### 5. 测试转写

```bash
# 准备一个测试音频文件
curl -X POST https://cf-whisper-worker.<你的子域>.workers.dev/api/v1/transcribe \
  -F "file=@test-audio.mp3"
```

## API

### POST /api/v1/transcribe

接收音频文件（multipart/form-data），返回 Cloudflare Workers AI whisper 转写结果。

```bash
# 请求（multipart/form-data）
-F "file=@audio.mp3"

# 响应（Cloudflare Workers AI 格式）
{
  "text": "完整转写文本",
  "segments": [
    {"id": 0, "start": 0.0, "end": 2.5, "text": "第一句"},
    {"id": 1, "start": 2.5, "end": 5.1, "text": "第二句"}
  ]
}
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| WHISPER_MODEL | @cf/openai/whisper | Cloudflare Workers AI whisper 模型 |

## 限制

- Cloudflare Workers AI 免费额度有限
- 单次请求最大 30 秒音频
- 需要 Cloudflare 账号（免费即可）

## 替代方案

如果不想用 Cloudflare Workers AI，可以用：
- `whisper-local`：本地 whisper.cpp（推荐）
- AI 服务内置的 OpenAI 兼容端点（WHISPER_API_URL）
