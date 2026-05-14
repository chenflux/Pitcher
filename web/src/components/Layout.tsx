import { type ReactNode, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useTheme, THEMES } from "../context/ThemeContext";

const navItems = [
  { path: "/", label: "仪表盘", icon: "📊" },
  { path: "/nodes", label: "节点管理", icon: "🖥️" },
  { path: "/services", label: "服务管理", icon: "🔧" },
  { path: "/logs", label: "日志查看", icon: "📋" },
  { path: "/rules", label: "规则中心", icon: "🛡️" },
  { path: "/configs", label: "配置管理", icon: "📝" },
  { path: "/settings", label: "系统设置", icon: "⚙️" },
];

export default function Layout({ children }: { children: ReactNode }) {
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { theme, setTheme } = useTheme();

  const logout = () => {
    localStorage.removeItem("token");
    window.location.href = "/login";
  };

  return (
    <div className="flex h-screen bg-background text-foreground">
      <aside
        className={`fixed inset-y-0 left-0 z-50 w-64 bg-card border-r border-border text-foreground transform transition-transform lg:relative lg:translate-x-0 ${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="p-4 border-b border-border">
          <h1 className="text-xl font-bold">🐝 HoneyWatch</h1>
          <p className="text-xs text-muted-foreground mt-1">蜜罐管理系统</p>
        </div>
        <nav className="mt-4">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              onClick={() => setSidebarOpen(false)}
              className={`flex items-center px-4 py-3 text-sm transition-colors ${
                location.pathname === item.path
                  ? "bg-primary text-primary-foreground"
                  : "text-foreground hover:bg-accent"
              }`}
            >
              <span className="mr-3">{item.icon}</span>
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="absolute bottom-0 left-0 right-0 p-4">
          <button
            onClick={logout}
            className="w-full px-4 py-2 text-sm text-muted-foreground hover:bg-accent rounded transition-colors text-left"
          >
            🚪 退出登录
          </button>
        </div>
      </aside>

      {sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="bg-card shadow-sm px-4 py-3 flex items-center justify-between lg:hidden border-b border-border">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="text-foreground"
          >
            ☰
          </button>
          <span className="ml-3 font-semibold">🐝 HoneyWatch</span>
          <select
            value={theme}
            onChange={(e) => setTheme(e.target.value as any)}
            className="bg-muted text-muted-foreground text-xs rounded px-2 py-1 border border-border"
          >
            {THEMES.map((t) => (
              <option key={t.id} value={t.id}>{t.name}</option>
            ))}
          </select>
        </header>

        <header className="hidden lg:flex items-center justify-end px-6 py-2 border-b border-border bg-card gap-3">
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">主题</span>
            <select
              value={theme}
              onChange={(e) => setTheme(e.target.value as any)}
              className="bg-muted text-foreground text-xs rounded px-2 py-1 border border-border focus:outline-none focus:ring-1 focus:ring-ring"
            >
              {THEMES.map((t) => (
                <option key={t.id} value={t.id}>{t.name}</option>
              ))}
            </select>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
