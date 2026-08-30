import React, { useCallback, useEffect, useMemo, useState } from "react";
import { LogIn, Mail, RefreshCcw, UserRound } from "lucide-react";
import {
  Button,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  Input,
  Label,
} from "@mybilibili-studio/ui";
import { MYBILIBILI_WEB_URL } from "../config/api-endpoints";
import { mybilibiliAuthApi } from "../services/mybilibili-api";
import { useAuthStore } from "../stores/auth-store";
import { toast } from "../stores/notification-store";

interface AuthDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

type LoginMode = "password" | "email";

const isEmail = (value: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);

export const AuthDialog: React.FC<AuthDialogProps> = ({ open, onOpenChange }) => {
  const [mode, setMode] = useState<LoginMode>("password");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [emailCode, setEmailCode] = useState("");
  const [captchaId, setCaptchaId] = useState("");
  const [captchaQuestion, setCaptchaQuestion] = useState("");
  const [captchaAnswer, setCaptchaAnswer] = useState("");
  const [captchaLoading, setCaptchaLoading] = useState(false);
  const [sendCodeLoading, setSendCodeLoading] = useState(false);
  const [countdown, setCountdown] = useState(0);

  const loginLoading = useAuthStore((state) => state.loginLoading);
  const loginWithPassword = useAuthStore((state) => state.loginWithPassword);
  const loginWithEmailCode = useAuthStore((state) => state.loginWithEmailCode);

  const registerUrl = useMemo(() => `${MYBILIBILI_WEB_URL}/register`, []);

  const loadCaptcha = useCallback(async () => {
    setCaptchaLoading(true);
    try {
      const captcha = await mybilibiliAuthApi.createCaptcha();
      setCaptchaId(captcha.captchaId);
      setCaptchaQuestion(captcha.question);
      setCaptchaAnswer("");
    } catch (error) {
      toast.error("验证码加载失败", error instanceof Error ? error.message : "请检查网关服务是否启动");
    } finally {
      setCaptchaLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      void loadCaptcha();
    }
  }, [open, loadCaptcha]);

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = window.setInterval(() => {
      setCountdown((value) => Math.max(0, value - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [countdown]);

  const verifyCaptcha = useCallback(async () => {
    if (!captchaId || !captchaAnswer.trim()) {
      toast.warning("请输入图形验证码");
      return false;
    }
    const ok = await mybilibiliAuthApi.verifyCaptcha(captchaId, captchaAnswer);
    if (!ok) {
      toast.error("图形验证码错误");
      void loadCaptcha();
      return false;
    }
    return true;
  }, [captchaAnswer, captchaId, loadCaptcha]);

  const handlePasswordLogin = useCallback(async () => {
    const normalizedUsername = username.trim();
    if (!normalizedUsername || !password) {
      toast.warning("请输入账号和密码");
      return;
    }

    try {
      if (!(await verifyCaptcha())) return;
      await loginWithPassword(normalizedUsername, password);
      toast.success("登录成功");
      onOpenChange(false);
      setPassword("");
    } catch (error) {
      toast.error("登录失败", error instanceof Error ? error.message : "请检查账号和密码");
      void loadCaptcha();
    }
  }, [loadCaptcha, loginWithPassword, onOpenChange, password, username, verifyCaptcha]);

  const handleSendEmailCode = useCallback(async () => {
    const normalizedEmail = email.trim();
    if (!isEmail(normalizedEmail)) {
      toast.warning("请输入有效邮箱");
      return;
    }

    setSendCodeLoading(true);
    try {
      if (!(await verifyCaptcha())) return;
      await mybilibiliAuthApi.sendEmailCode(normalizedEmail);
      toast.success("验证码已发送");
      setCountdown(60);
      void loadCaptcha();
    } catch (error) {
      toast.error("验证码发送失败", error instanceof Error ? error.message : "请稍后重试");
      void loadCaptcha();
    } finally {
      setSendCodeLoading(false);
    }
  }, [email, loadCaptcha, verifyCaptcha]);

  const handleEmailLogin = useCallback(async () => {
    const normalizedEmail = email.trim();
    if (!isEmail(normalizedEmail) || !emailCode.trim()) {
      toast.warning("请输入邮箱和验证码");
      return;
    }

    try {
      await loginWithEmailCode(normalizedEmail, emailCode.trim());
      toast.success("登录成功");
      onOpenChange(false);
      setEmailCode("");
    } catch (error) {
      toast.error("登录失败", error instanceof Error ? error.message : "请检查邮箱验证码");
    }
  }, [email, emailCode, loginWithEmailCode, onOpenChange]);

  const handleSubmit = useCallback(
    (event: React.FormEvent) => {
      event.preventDefault();
      if (mode === "password") {
        void handlePasswordLogin();
      } else {
        void handleEmailLogin();
      }
    },
    [handleEmailLogin, handlePasswordLogin, mode],
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[420px] rounded-xl border-border bg-background-secondary text-text-primary">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserRound size={18} />
            登录 mybilibili
          </DialogTitle>
          <DialogDescription>
            使用主站账号登录，剪辑仍可本地使用，上传投稿时会复用此登录状态。
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-2 rounded-lg bg-background-tertiary p-1">
          <button
            type="button"
            onClick={() => setMode("password")}
            className={`h-9 rounded-md text-sm transition-colors ${
              mode === "password"
                ? "bg-background text-text-primary"
                : "text-text-muted hover:text-text-primary"
            }`}
          >
            密码登录
          </button>
          <button
            type="button"
            onClick={() => setMode("email")}
            className={`h-9 rounded-md text-sm transition-colors ${
              mode === "email"
                ? "bg-background text-text-primary"
                : "text-text-muted hover:text-text-primary"
            }`}
          >
            邮箱验证码
          </button>
        </div>

        <form className="space-y-4" onSubmit={handleSubmit}>
          {mode === "password" ? (
            <>
              <div className="space-y-2">
                <Label htmlFor="studio-login-username">账号</Label>
                <Input
                  id="studio-login-username"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  placeholder="用户名或邮箱"
                  autoComplete="username"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="studio-login-password">密码</Label>
                <Input
                  id="studio-login-password"
                  type="password"
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder="请输入密码"
                  autoComplete="current-password"
                />
              </div>
            </>
          ) : (
            <>
              <div className="space-y-2">
                <Label htmlFor="studio-login-email">邮箱</Label>
                <Input
                  id="studio-login-email"
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="请输入邮箱"
                  autoComplete="email"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="studio-login-email-code">邮箱验证码</Label>
                <div className="flex gap-2">
                  <Input
                    id="studio-login-email-code"
                    value={emailCode}
                    onChange={(event) => setEmailCode(event.target.value)}
                    placeholder="6 位验证码"
                    inputMode="numeric"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    disabled={sendCodeLoading || countdown > 0}
                    onClick={handleSendEmailCode}
                    className="shrink-0"
                  >
                    <Mail size={15} />
                    {countdown > 0 ? `${countdown}s` : "发送"}
                  </Button>
                </div>
              </div>
            </>
          )}

          <div className="space-y-2">
            <Label htmlFor="studio-login-captcha">图形验证码</Label>
            <div className="flex gap-2">
              <Input
                id="studio-login-captcha"
                value={captchaAnswer}
                onChange={(event) => setCaptchaAnswer(event.target.value)}
                placeholder="答案"
                inputMode="numeric"
              />
              <button
                type="button"
                onClick={() => void loadCaptcha()}
                className="flex min-w-[120px] items-center justify-center gap-2 rounded-md border border-border bg-background-tertiary px-3 text-sm text-text-primary hover:bg-background"
              >
                {captchaLoading ? (
                  <RefreshCcw size={14} className="animate-spin" />
                ) : (
                  captchaQuestion || "刷新"
                )}
              </button>
            </div>
          </div>

          <Button type="submit" className="w-full" disabled={loginLoading || captchaLoading}>
            <LogIn size={16} />
            {loginLoading ? "登录中" : "登录"}
          </Button>
        </form>

        <div className="flex items-center justify-between text-xs text-text-muted">
          <span>没有账号？</span>
          <a
            href={registerUrl}
            target="_blank"
            rel="noreferrer"
            className="text-primary hover:underline"
          >
            去主站注册
          </a>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default AuthDialog;
