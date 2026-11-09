import { Subject } from 'rxjs';
import { mutation, PersistentStatefulService, Service, ViewHandler } from 'services/core';

export type TPlatform = 'mybilibili';

export enum EAuthProcessState {
  Idle = 'idle',
  Loading = 'loading',
  InProgress = 'progress',
}

export interface IPlatformAuth {
  type: TPlatform;
  username?: string;
  channelId?: string;
  token?: string;
  access_token?: string;
}

export interface IUserAuth {
  widgetToken: string;
  apiToken: string;
  platform?: IPlatformAuth;
  primaryPlatform: TPlatform;
  platforms: Record<TPlatform, IPlatformAuth>;
  partition?: string;
  hasRelogged: boolean;
  slid?: { id: string; username: string };
}

interface IUserServiceState {
  loginValidated: boolean;
  auth?: IUserAuth;
  authProcessState: EAuthProcessState;
  expires?: string;
  userId?: number;
  createdAt?: number;
  isRelog?: boolean;
}

export type LoginLifecycleOptions = {
  init: () => Promise<void>;
  destroy: () => Promise<void>;
  context: Service;
};

export type LoginLifecycle = {
  destroy: () => Promise<void>;
};

export function setSentryContext(_ctx: { username: string; platform: string }) {}

class UserViews extends ViewHandler<IUserServiceState> {
  get isLoggedIn() {
    return !!(this.state.auth?.apiToken || this.state.auth?.widgetToken);
  }

  get username() {
    return this.state.auth?.platform?.username || 'mybilibili';
  }

  get platform() {
    return this.state.auth?.platform || null;
  }

  get platforms() {
    return this.state.auth?.platforms || {};
  }

  get linkedPlatforms(): TPlatform[] {
    return this.isLoggedIn ? ['mybilibili'] : [];
  }

  get hasSLID() {
    return false;
  }

  get auth() {
    return this.state.auth;
  }

  async appStoreUrl() {
    return '';
  }
}

export class UserService extends PersistentStatefulService<IUserServiceState> {
  static defaultState: IUserServiceState = {
    loginValidated: false,
    authProcessState: EAuthProcessState.Idle,
    createdAt: Date.now(),
  };

  userLogin = new Subject<IUserAuth>();
  userLoginFinished = new Subject<IUserAuth | void>();
  userLogout = new Subject<void>();
  primaryPlatformChanged = new Subject<TPlatform>();

  get views() {
    return new UserViews(this.state);
  }

  get isLoggedIn() {
    return this.views.isLoggedIn;
  }

  get isAlphaGroup() {
    return false;
  }

  get apiToken() {
    return this.state.auth?.apiToken || '';
  }

  get platform() {
    return this.views.platform;
  }

  getLocalUserId() {
    return this.state.userId?.toString() || 'local-mybilibili-user';
  }

  setPrimaryPlatform(platform: TPlatform) {
    if (this.state.auth) this.SET_PRIMARY_PLATFORM(platform);
    this.primaryPlatformChanged.next(platform);
  }

  logout() {
    this.LOGOUT();
    this.userLogout.next();
  }

  async withLifecycle({ init, destroy }: LoginLifecycleOptions): Promise<LoginLifecycle> {
    await init();
    return { destroy };
  }

  @mutation()
  private SET_PRIMARY_PLATFORM(platform: TPlatform) {
    if (!this.state.auth) return;
    this.state.auth.primaryPlatform = platform;
    this.state.auth.platform = this.state.auth.platforms[platform];
  }

  @mutation()
  private LOGOUT() {
    this.state.auth = undefined;
    this.state.loginValidated = false;
  }
}
