import { Inject, StatefulService, mutation } from 'services/core';
import { StreamSettingsService } from 'services/settings/streaming';
import {
  IMybilibiliLoginSession,
  IMybilibiliLoginUser,
  MybilibiliApiClient,
} from './api-client';

export interface IMybilibiliLiveRoom {
  id: number;
  userId: number;
  roomName: string;
  streamKey: string;
  status: 'live' | 'offline' | string;
  coverUrl?: string;
  category?: string;
  scheduledAt?: string | number | null;
}

export interface IMybilibiliLiveState {
  baseUrl: string;
  token: string;
  refreshToken: string;
  user: IMybilibiliLoginUser | null;
  room: IMybilibiliLiveRoom | null;
  rtmpUrl: string;
  lastError: string;
}

const DEFAULT_BASE_URL = 'http://localhost:8080/api';
const DEFAULT_ROOM_NAME = '我的直播间';

const getGatewayHost = (baseUrl: string) => {
  try {
    return new URL(baseUrl).hostname || 'localhost';
  } catch {
    return 'localhost';
  }
};

export class MybilibiliLiveService extends StatefulService<IMybilibiliLiveState> {
  @Inject() private streamSettingsService: StreamSettingsService;

  static initialState: IMybilibiliLiveState = {
    baseUrl: DEFAULT_BASE_URL,
    token: '',
    refreshToken: '',
    user: null,
    room: null,
    rtmpUrl: 'rtmp://localhost/live',
    lastError: '',
  };

  get api() {
    return new MybilibiliApiClient({
      baseUrl: this.state.baseUrl,
      token: this.state.token,
    });
  }

  setConnection(baseUrl: string, token = this.state.token) {
    const nextBaseUrl = baseUrl || DEFAULT_BASE_URL;
    const nextToken = nextBaseUrl === this.state.baseUrl ? token : '';
    this.SET_CONNECTION(nextBaseUrl, nextToken || '');
  }

  async login(baseUrl: string, username: string, password: string) {
    const nextBaseUrl = baseUrl || DEFAULT_BASE_URL;
    const account = username.trim();
    if (!account || !password) {
      throw new Error('请输入 mybilibili 账号和密码');
    }

    try {
      const session = await new MybilibiliApiClient({ baseUrl: nextBaseUrl }).login(
        account,
        password,
      );
      this.SET_SESSION(nextBaseUrl, session);
      return session;
    } catch (error: unknown) {
      this.SET_ERROR(error instanceof Error ? error.message : '登录失败');
      throw error;
    }
  }

  async refreshLogin() {
    if (!this.state.refreshToken) {
      throw new Error('缺少 refreshToken，请重新登录');
    }

    try {
      const session = await this.api.refreshToken(this.state.refreshToken);
      this.SET_SESSION(this.state.baseUrl, {
        ...session,
        user: session.user || this.state.user || undefined,
      });
      return session;
    } catch (error: unknown) {
      this.SET_ERROR(error instanceof Error ? error.message : '刷新登录状态失败');
      throw error;
    }
  }

  logout() {
    this.CLEAR_SESSION();
  }

  async loadOrCreateRoom() {
    try {
      if (!this.state.token) throw new Error('请先登录 mybilibili 站点');
      const api = this.api;
      let room = await api.get<IMybilibiliLiveRoom | null>('/live/room/my');
      if (!room) {
        room = await api.post<IMybilibiliLiveRoom>('/live/room/create', {
          roomName: DEFAULT_ROOM_NAME,
        });
      }
      this.SET_ROOM(room);
      this.applyStreamSettings();
      return room;
    } catch (error: unknown) {
      this.SET_ERROR(error instanceof Error ? error.message : '加载直播间失败');
      throw error;
    }
  }

  async refreshRoom() {
    try {
      if (!this.state.token) throw new Error('请先登录 mybilibili 站点');
      const room = await this.api.get<IMybilibiliLiveRoom | null>('/live/room/my');
      this.SET_ROOM(room);
      if (room?.streamKey) this.applyStreamSettings();
      return room;
    } catch (error: unknown) {
      this.SET_ERROR(error instanceof Error ? error.message : '刷新直播间失败');
      throw error;
    }
  }

  async updateRoom(
    patch: Partial<Pick<IMybilibiliLiveRoom, 'roomName' | 'coverUrl' | 'category'>>,
  ) {
    try {
      if (!this.state.room) throw new Error('直播间未加载');
      const room = await this.api.put<IMybilibiliLiveRoom>(
        `/live/room/${this.state.room.id}`,
        patch,
      );
      this.SET_ROOM(room);
      return room;
    } catch (error: unknown) {
      this.SET_ERROR(error instanceof Error ? error.message : '更新直播间失败');
      throw error;
    }
  }

  async updateRoomStatus(status: 'live' | 'offline') {
    try {
      if (!this.state.room) throw new Error('直播间未加载');
      await this.api.put(`/live/room/${this.state.room.id}/status`, { status });
      const room = await this.refreshRoom();
      return room;
    } catch (error: unknown) {
      this.SET_ERROR(error instanceof Error ? error.message : '更新直播状态失败');
      throw error;
    }
  }

  applyStreamSettings() {
    if (!this.state.room?.streamKey) {
      throw new Error('直播间缺少 streamKey');
    }

    this.streamSettingsService.setSettings({
      protectedModeEnabled: false,
      protectedModeMigrationRequired: false,
      streamType: 'rtmp_custom',
      server: this.state.rtmpUrl,
      key: this.state.room.streamKey,
    });
  }

  @mutation()
  private SET_CONNECTION(baseUrl: string, token: string) {
    this.state.baseUrl = baseUrl;
    this.state.token = token;
    if (!token) {
      this.state.refreshToken = '';
      this.state.user = null;
    }
    this.state.room = null;
    this.state.rtmpUrl = `rtmp://${getGatewayHost(baseUrl)}/live`;
    this.state.lastError = '';
  }

  @mutation()
  private SET_SESSION(baseUrl: string, session: IMybilibiliLoginSession) {
    this.state.baseUrl = baseUrl;
    this.state.token = session.token;
    this.state.refreshToken = session.refreshToken || this.state.refreshToken || '';
    this.state.user = session.user || this.state.user || null;
    this.state.room = null;
    this.state.rtmpUrl = `rtmp://${getGatewayHost(baseUrl)}/live`;
    this.state.lastError = '';
  }

  @mutation()
  private CLEAR_SESSION() {
    this.state.token = '';
    this.state.refreshToken = '';
    this.state.user = null;
    this.state.room = null;
    this.state.lastError = '';
  }

  @mutation()
  private SET_ROOM(room: IMybilibiliLiveRoom | null) {
    this.state.room = room;
    this.state.lastError = '';
  }

  @mutation()
  private SET_ERROR(message: string) {
    this.state.lastError = message;
  }
}
