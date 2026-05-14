import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { authApi } from "../api";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [captcha, setCaptcha] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [captchaRequired, setCaptchaRequired] = useState(false);
  const [captchaCode, setCaptchaCode] = useState("");
  const [captchaId, setCaptchaId] = useState("");
  const [remainingAttempts, setRemainingAttempts] = useState(5);
  const [lockedUntil, setLockedUntil] = useState<number | null>(null);
  const [graylistTTL, setGraylistTTL] = useState<number | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    refreshCaptcha();
  }, []);

  const refreshCaptcha = async () => {
    try {
      const res = await authApi.getCaptcha();
      setCaptchaCode(res.data.code);
      setCaptchaId(res.data.id);
    } catch (e) {
      console.error("Failed to load captcha", e);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await authApi.login(username, password, captchaRequired ? captcha : undefined, captchaId);
      if (res.data.success) {
        localStorage.setItem("token", res.data.token);
        navigate("/");
      } else {
        if (res.data.captcha_required) {
          setCaptchaRequired(true);
          refreshCaptcha();
          setCaptcha("");
        }
        if (res.data.remaining_attempts !== undefined) {
          setRemainingAttempts(res.data.remaining_attempts);
        }
        if (res.data.locked_until) {
          setLockedUntil(res.data.locked_until);
        }
        if (res.data.graylist_ttl) {
          setGraylistTTL(res.data.graylist_ttl);
        }
        setError(res.data.error || "登录失败");
      }
    } catch (err: any) {
      const data = err.response?.data;
      if (data?.captcha_required) {
        setCaptchaRequired(true);
        refreshCaptcha();
        setCaptcha("");
      }
      if (data?.remaining_attempts !== undefined) {
        setRemainingAttempts(data.remaining_attempts);
      }
      if (data?.locked_until) {
        setLockedUntil(data.locked_until);
      }
      if (data?.graylist_ttl) {
        setGraylistTTL(data.graylist_ttl);
      }
      setError(data?.error || "登录失败");
    } finally {
      setLoading(false);
    }
  };

  const isLocked = !!(lockedUntil && Date.now() < lockedUntil);

  return (
    <div className="min-h-screen flex bg-background">
      <div className="hidden lg:flex lg:w-1/2 items-center justify-center bg-card">
        <div className="text-center px-12">
          <div className="text-8xl mb-6">🚀</div>
          <h1 className="text-5xl font-bold text-foreground mb-4">Pitcher</h1>
          <p className="text-xl text-muted-foreground mb-8">Web平台</p>
          <div className="flex justify-center gap-8 text-muted-foreground">
            <div className="text-center">
              <div className="text-3xl font-bold text-primary">高可用</div>
              <div className="text-sm mt-1">分布式架构</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-green-500">实时</div>
              <div className="text-sm mt-1">状态监控</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-yellow-500">弹性</div>
              <div className="text-sm mt-1">安全防护</div>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center px-4">
        <div className="bg-card rounded-2xl shadow-2xl p-10 w-full max-w-md">
          <div className="lg:hidden text-center mb-6">
            <div className="text-5xl mb-2">🚀</div>
            <h1 className="text-2xl font-bold text-foreground">Pitcher</h1>
          </div>
          <h2 className="text-2xl font-bold text-foreground mb-2 hidden lg:block">欢迎回来</h2>
          <p className="text-muted-foreground mb-8 hidden lg:block">登录到管理系统</p>

          {graylistTTL && (
            <div className="mb-4 bg-yellow-500/10 border border-yellow-500/30 text-yellow-500 px-4 py-3 rounded-lg text-sm">
              <p className="font-medium">当前 IP 处于访问限制期</p>
              <p className="text-xs mt-1">请 {Math.ceil(graylistTTL / 60)} 分钟后再试</p>
            </div>
          )}

          {isLocked && (
            <div className="mb-4 bg-red-500/10 border border-red-500/30 text-red-500 px-4 py-3 rounded-lg text-sm">
              <p className="font-medium">账户已锁定</p>
              <p className="text-xs mt-1">请 {Math.ceil((lockedUntil - Date.now()) / 1000)} 秒后再试</p>
            </div>
          )}

          {remainingAttempts <= 3 && remainingAttempts > 0 && !isLocked && (
            <div className="mb-4 bg-yellow-500/10 border border-yellow-500/30 text-yellow-500 px-4 py-3 rounded-lg text-sm">
              <p className="font-medium">登录失败过多</p>
              <p className="text-xs mt-1">剩余 {remainingAttempts} 次尝试机会</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-5">
            {error && (
              <div className="bg-red-500/10 text-red-500 px-4 py-3 rounded-lg text-sm flex items-center gap-2">
                <span>✕</span>
                <span>{error}</span>
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">用户名</label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className="w-full px-4 py-2.5 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring bg-background text-foreground transition"
                placeholder="请输入用户名"
                required
                disabled={isLocked}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">密码</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-2.5 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring bg-background text-foreground transition"
                placeholder="请输入密码"
                required
                disabled={isLocked}
              />
            </div>

            {captchaRequired && (
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">验证码</label>
                <div className="flex gap-3">
                  <input
                    type="text"
                    value={captcha}
                    onChange={(e) => setCaptcha(e.target.value)}
                    className="flex-1 px-4 py-2.5 border border-input rounded-lg focus:outline-none focus:ring-2 focus:ring-ring bg-background text-foreground transition"
                    placeholder="请输入右侧计算结果"
                    required
                    maxLength={6}
                    disabled={isLocked}
                  />
                  <button
                    type="button"
                    onClick={refreshCaptcha}
                    className="px-4 py-2.5 bg-muted text-foreground rounded-lg hover:bg-muted/80 transition text-sm font-medium"
                  >
                    刷新
                  </button>
                </div>
                <p className="text-xs text-muted-foreground mt-1.5">
                  请计算: <span className="font-bold text-foreground">{captchaCode}</span>
                </p>
              </div>
            )}

            <button
              type="submit"
              disabled={loading || isLocked}
              className="w-full bg-primary text-primary-foreground py-2.5 px-4 rounded-lg hover:opacity-90 disabled:opacity-50 transition-colors font-medium"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                  登录中...
                </span>
              ) : isLocked ? "账户已锁定" : "登 录"}
            </button>
          </form>
          <div className="mt-6 pt-6 border-t border-border text-center">
            <p className="text-xs text-muted-foreground">默认账号: admin / admin</p>
          </div>
        </div>
      </div>
    </div>
  );
}