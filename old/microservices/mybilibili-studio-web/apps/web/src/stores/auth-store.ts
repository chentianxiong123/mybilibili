import { create } from "zustand";
import {
  clearAuthSession,
  getStoredAuthSession,
  isJwtExpired,
  mybilibiliAuthApi,
  storeAuthSession,
  type AuthSession,
  type MybilibiliUser,
} from "../services/mybilibili-api";

type AuthStatus = "checking" | "authenticated" | "anonymous";

interface AuthState {
  initialized: boolean;
  status: AuthStatus;
  user: MybilibiliUser | null;
  token: string | null;
  refreshToken: string | null;
  loginLoading: boolean;
  initialize: () => Promise<void>;
  loginWithPassword: (username: string, password: string) => Promise<void>;
  loginWithEmailCode: (email: string, emailCode: string) => Promise<void>;
  logout: () => void;
  refreshSession: () => Promise<boolean>;
}

function applySession(session: AuthSession) {
  storeAuthSession(session);
  return {
    initialized: true,
    status: "authenticated" as const,
    user: session.user,
    token: session.token,
    refreshToken: session.refreshToken,
  };
}

function anonymousState() {
  return {
    initialized: true,
    status: "anonymous" as const,
    user: null,
    token: null,
    refreshToken: null,
  };
}

export const useAuthStore = create<AuthState>((set, get) => ({
  initialized: false,
  status: "checking",
  user: null,
  token: null,
  refreshToken: null,
  loginLoading: false,

  initialize: async () => {
    if (get().initialized) return;

    const stored = getStoredAuthSession();
    if (!stored) {
      set(anonymousState());
      return;
    }

    if (isJwtExpired(stored.token) && stored.refreshToken) {
      try {
        const refreshed = await mybilibiliAuthApi.refresh(stored.refreshToken);
        set(applySession({ ...refreshed, user: stored.user }));
      } catch {
        clearAuthSession();
        set(anonymousState());
      }
      return;
    }

    set(applySession(stored));

    mybilibiliAuthApi
      .fetchUser(stored.user.id)
      .then((user) => {
        const current = get();
        if (current.status !== "authenticated" || current.user?.id !== user.id) return;
        const nextSession = {
          token: current.token!,
          refreshToken: current.refreshToken!,
          user,
        };
        set(applySession(nextSession));
      })
      .catch(() => {
        // Token is still useful for later API calls; stale profile details are acceptable.
      });
  },

  loginWithPassword: async (username: string, password: string) => {
    set({ loginLoading: true });
    try {
      const session = await mybilibiliAuthApi.loginWithPassword(username, password);
      set(applySession(session));
    } finally {
      set({ loginLoading: false });
    }
  },

  loginWithEmailCode: async (email: string, emailCode: string) => {
    set({ loginLoading: true });
    try {
      const session = await mybilibiliAuthApi.loginWithEmailCode(email, emailCode);
      set(applySession(session));
    } finally {
      set({ loginLoading: false });
    }
  },

  logout: () => {
    clearAuthSession();
    set(anonymousState());
  },

  refreshSession: async () => {
    const refreshToken = get().refreshToken || getStoredAuthSession()?.refreshToken;
    if (!refreshToken) return false;
    try {
      const refreshed = await mybilibiliAuthApi.refresh(refreshToken);
      set(applySession(refreshed));
      return true;
    } catch {
      clearAuthSession();
      set(anonymousState());
      return false;
    }
  },
}));
