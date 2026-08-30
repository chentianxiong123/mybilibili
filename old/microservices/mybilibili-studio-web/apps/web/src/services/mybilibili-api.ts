import { MYBILIBILI_API_BASE_URL } from "../config/api-endpoints";

export interface ApiResult<T> {
  code: number;
  message?: string;
  data: T;
}

export interface MybilibiliUser {
  id: number;
  username: string;
  nickname?: string;
  avatar?: string;
  email?: string;
  level?: number;
}

export interface AuthSession {
  user: MybilibiliUser;
  token: string;
  refreshToken: string;
}

export interface CaptchaChallenge {
  captchaId: string;
  question: string;
}

export interface JwtPayload {
  sub?: string;
  username?: string;
  exp?: number;
  type?: string;
}

const TOKEN_KEY = "token";
const REFRESH_TOKEN_KEY = "refreshToken";
const USER_KEY = "user";

function joinUrl(baseUrl: string, path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  if (baseUrl === "/api") return `${baseUrl}${normalizedPath}`;
  return `${baseUrl.replace(/\/+$/, "")}${normalizedPath}`;
}

function readJson<T>(value: string | null): T | null {
  if (!value) return null;
  try {
    return JSON.parse(value) as T;
  } catch {
    return null;
  }
}

function decodeBase64Url(value: string): string {
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(
    normalized.length + ((4 - (normalized.length % 4)) % 4),
    "=",
  );
  return atob(padded);
}

export function decodeJwtPayload(token: string | null): JwtPayload | null {
  if (!token) return null;
  const [, payload] = token.split(".");
  if (!payload) return null;
  try {
    return JSON.parse(decodeBase64Url(payload)) as JwtPayload;
  } catch {
    return null;
  }
}

export function isJwtExpired(token: string | null, leewaySeconds = 30): boolean {
  const payload = decodeJwtPayload(token);
  if (!payload?.exp) return true;
  return payload.exp * 1000 <= Date.now() + leewaySeconds * 1000;
}

export function userFromToken(token: string | null): MybilibiliUser | null {
  const payload = decodeJwtPayload(token);
  if (!payload?.sub) return null;
  const id = Number(payload.sub);
  if (!Number.isFinite(id)) return null;
  return {
    id,
    username: payload.username || `用户${id}`,
    nickname: payload.username || `用户${id}`,
  };
}

export function getStoredAuthSession(): AuthSession | null {
  if (typeof window === "undefined") return null;
  const token = localStorage.getItem(TOKEN_KEY);
  const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
  if (!token || !refreshToken) return null;

  const storedUser = readJson<MybilibiliUser>(localStorage.getItem(USER_KEY));
  const tokenUser = userFromToken(token);
  const user = storedUser || tokenUser;
  if (!user) return null;

  return { token, refreshToken, user };
}

export function storeAuthSession(session: AuthSession): void {
  localStorage.setItem(TOKEN_KEY, session.token);
  localStorage.setItem(REFRESH_TOKEN_KEY, session.refreshToken);
  localStorage.setItem(USER_KEY, JSON.stringify(session.user));
}

export function clearAuthSession(): void {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

async function readResult<T>(response: Response): Promise<ApiResult<T>> {
  const contentType = response.headers.get("content-type") || "";
  const body = contentType.includes("application/json")
    ? await response.json()
    : { code: response.status, message: await response.text(), data: null };

  if (!response.ok) {
    throw new Error(body?.message || `请求失败：${response.status}`);
  }
  return body as ApiResult<T>;
}

async function refreshTokenRequest(refreshToken: string): Promise<AuthSession> {
  const response = await fetch(joinUrl(MYBILIBILI_API_BASE_URL, "/user/token/refresh"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refreshToken }),
  });
  const result = await readResult<AuthSession>(response);
  if (result.code !== 200 || !result.data?.token) {
    throw new Error(result.message || "登录续期失败");
  }

  const existing = getStoredAuthSession();
  const user = existing?.user || userFromToken(result.data.token);
  if (!user) throw new Error("登录续期返回的用户信息无效");

  return {
    token: result.data.token,
    refreshToken: result.data.refreshToken || refreshToken,
    user,
  };
}

export async function mybilibiliFetch<T>(
  path: string,
  init: RequestInit = {},
  options: { auth?: boolean; retryOnUnauthorized?: boolean } = {},
): Promise<ApiResult<T>> {
  const auth = options.auth ?? true;
  const retryOnUnauthorized = options.retryOnUnauthorized ?? true;
  const headers = new Headers(init.headers);
  if (!headers.has("Content-Type") && init.body && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const session = getStoredAuthSession();
  if (auth && session?.token) {
    headers.set("Authorization", `Bearer ${session.token}`);
  }

  const response = await fetch(joinUrl(MYBILIBILI_API_BASE_URL, path), {
    ...init,
    headers,
  });

  if (response.status === 401 && auth && retryOnUnauthorized && session?.refreshToken) {
    const refreshed = await refreshTokenRequest(session.refreshToken);
    storeAuthSession(refreshed);
    const retryHeaders = new Headers(headers);
    retryHeaders.set("Authorization", `Bearer ${refreshed.token}`);
    const retryResponse = await fetch(joinUrl(MYBILIBILI_API_BASE_URL, path), {
      ...init,
      headers: retryHeaders,
    });
    return readResult<T>(retryResponse);
  }

  return readResult<T>(response);
}

export const mybilibiliAuthApi = {
  async loginWithPassword(username: string, password: string): Promise<AuthSession> {
    const result = await mybilibiliFetch<AuthSession>(
      "/user/login",
      {
        method: "POST",
        body: JSON.stringify({ username, password, loginType: "password" }),
      },
      { auth: false },
    );
    if (result.code !== 200 || !result.data?.token) {
      throw new Error(result.message || "登录失败");
    }
    return result.data;
  },

  async loginWithEmailCode(email: string, emailCode: string): Promise<AuthSession> {
    const result = await mybilibiliFetch<AuthSession>(
      "/user/login",
      {
        method: "POST",
        body: JSON.stringify({ email, emailCode, loginType: "email_code" }),
      },
      { auth: false },
    );
    if (result.code !== 200 || !result.data?.token) {
      throw new Error(result.message || "登录失败");
    }
    return result.data;
  },

  async refresh(refreshToken: string): Promise<AuthSession> {
    return refreshTokenRequest(refreshToken);
  },

  async createCaptcha(): Promise<CaptchaChallenge> {
    const result = await mybilibiliFetch<CaptchaChallenge>(
      "/captcha/new",
      { method: "POST" },
      { auth: false },
    );
    if (result.code !== 200 || !result.data?.captchaId) {
      throw new Error(result.message || "验证码生成失败");
    }
    return result.data;
  },

  async verifyCaptcha(captchaId: string, answer: string): Promise<boolean> {
    const result = await mybilibiliFetch<boolean>(
      "/captcha/verify",
      {
        method: "POST",
        body: JSON.stringify({ captchaId, answer }),
      },
      { auth: false },
    );
    return result.code === 200 && Boolean(result.data);
  },

  async sendEmailCode(email: string): Promise<void> {
    const result = await mybilibiliFetch<void>(
      "/user/email/code",
      {
        method: "POST",
        body: JSON.stringify({ email }),
      },
      { auth: false },
    );
    if (result.code !== 200) {
      throw new Error(result.message || "验证码发送失败");
    }
  },

  async fetchUser(userId: number): Promise<MybilibiliUser> {
    const result = await mybilibiliFetch<MybilibiliUser>(
      `/user/${userId}`,
      { method: "GET" },
      { auth: false },
    );
    if (result.code !== 200 || !result.data) {
      throw new Error(result.message || "用户信息获取失败");
    }
    return result.data;
  },
};
