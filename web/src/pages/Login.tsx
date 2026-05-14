import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { authApi } from "../api";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const res = await authApi.login(username, password);
      if (res.data.success) {
        localStorage.setItem("token", res.data.token);
        navigate("/");
      } else {
        setError(res.data.error || "登录失败");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "登录失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex bg-background">
      <div className="hidden lg:flex lg:w-1/2 items-center justify-center bg-card">
        <div className="text-center px-12">
          <div className="text-8xl mb-6">🐝</div>
          <h1 className="text-5xl font-bold text-foreground mb-4">HoneyWatch</h1>
          <p className="text-xl text-muted-foreground mb-8">分布式蜜罐监控系统</p>
          <div className="flex justify-center gap-8 text-muted-foreground">
            <div className="text-center">
              <div className="text-3xl font-bold text-primary">7+</div>
              <div className="text-sm mt-1">蜜罐类型</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-green-500">实时</div>
              <div className="text-sm mt-1">攻击检测</div>
            </div>
            <div className="text-center">
              <div className="text-3xl font-bold text-yellow-500">分布式</div>
              <div className="text-sm mt-1">节点部署</div>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center px-4">
        <div className="bg-card rounded-2xl shadow-2xl p-10 w-full max-w-md">
          <div className="lg:hidden text-center mb-6">
            <div className="text-5xl mb-2">🐝</div>
            <h1 className="text-2xl font-bold text-foreground">HoneyWatch</h1>
          </div>
          <h2 className="text-2xl font-bold text-foreground mb-2 hidden lg:block">欢迎回来</h2>
          <p className="text-muted-foreground mb-8 hidden lg:block">登录到蜜罐管理系统 v2.0</p>
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
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="w-full bg-primary text-primary-foreground py-2.5 px-4 rounded-lg hover:opacity-90 disabled:opacity-50 transition-colors font-medium"
            >
              {loading ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                  登录中...
                </span>
              ) : "登 录"}
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
