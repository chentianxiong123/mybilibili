/**
 * Centralized API endpoint configuration.
 *
 * All external service URLs should be defined here so they can be
 * swapped for different environments or self-hosted instances.
 */

const trimUrl = (value: string | undefined) => (value ?? "").trim().replace(/\/+$/, "");

/** Main mybilibili gateway API. Defaults to the local Vite /api proxy. */
export const MYBILIBILI_API_BASE_URL =
  trimUrl(import.meta.env.VITE_MYBILIBILI_API_BASE_URL) || "/api";

/** Main web app entry used for account registration and profile pages. */
export const MYBILIBILI_WEB_URL =
  trimUrl(import.meta.env.VITE_MYBILIBILI_WEB_URL) || "http://127.0.0.1:5173";

/** Studio cloud services. Must be configured explicitly. */
export const STUDIO_CLOUD_URL = trimUrl(import.meta.env.VITE_MYBILIBILI_STUDIO_CLOUD_URL);

/** Local/self-hosted TTS service. */
export const STUDIO_TTS_URL = trimUrl(import.meta.env.VITE_MYBILIBILI_STUDIO_TTS_URL);

/** Local/self-hosted transcription service. */
export const STUDIO_TRANSCRIBE_URL = trimUrl(
  import.meta.env.VITE_MYBILIBILI_STUDIO_TRANSCRIBE_URL,
);

export function requireStudioEndpoint(url: string, serviceName: string): string {
  if (!url) {
    throw new Error(
      `${serviceName}接口未配置。请设置对应的 VITE_MYBILIBILI_STUDIO_*_URL 环境变量。`,
    );
  }
  return url;
}

/**
 * Third-party API base URLs.
 * These are used by the api-proxy service in dev mode (direct calls)
 * and by the Cloudflare Pages Function proxy in production.
 * Application code should use apiFetch() from services/api-proxy.ts
 * instead of importing these directly.
 */
