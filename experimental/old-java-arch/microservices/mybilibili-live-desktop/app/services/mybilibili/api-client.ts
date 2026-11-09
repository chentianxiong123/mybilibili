export interface IMybilibiliApiResult<T> {
  code: number;
  message?: string;
  data: T;
}

export interface IMybilibiliApiClientOptions {
  baseUrl?: string;
  token?: string;
}

export interface IMybilibiliLoginUser {
  id?: number;
  username?: string;
  nickname?: string;
  avatar?: string;
}

export interface IMybilibiliLoginSession {
  token: string;
  refreshToken?: string;
  user?: IMybilibiliLoginUser;
}

const trimRightSlash = (value: string) => value.replace(/\/+$/, '');

export class MybilibiliApiClient {
  private baseUrl: string;
  private token: string;

  constructor(options: IMybilibiliApiClientOptions = {}) {
    this.baseUrl = trimRightSlash(options.baseUrl || 'http://localhost:8080/api');
    this.token = options.token || '';
  }

  setBaseUrl(baseUrl: string) {
    this.baseUrl = trimRightSlash(baseUrl);
  }

  setToken(token: string) {
    this.token = token;
  }

  async get<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: 'GET' });
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'POST',
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  }

  async put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'PUT',
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  }

  async login(username: string, password: string): Promise<IMybilibiliLoginSession> {
    return this.post<IMybilibiliLoginSession>('/user/login', {
      username,
      password,
      loginType: 'password',
      loginIp: '',
    });
  }

  async refreshToken(refreshToken: string): Promise<IMybilibiliLoginSession> {
    return this.post<IMybilibiliLoginSession>('/user/token/refresh', { refreshToken });
  }

  private async request<T>(path: string, init: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(this.token ? { Authorization: `Bearer ${this.token}` } : {}),
        ...(init.headers || {}),
      },
    });

    if (!response.ok) {
      throw new Error(`mybilibili 接口请求失败：${response.status}`);
    }

    const result = (await response.json()) as IMybilibiliApiResult<T>;
    if (result.code !== 200) {
      throw new Error(result.message || 'mybilibili 接口请求失败');
    }
    return result.data;
  }
}
