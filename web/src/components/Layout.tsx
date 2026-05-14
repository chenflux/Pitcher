import { type ReactNode, useState, useRef, useEffect } from "react";
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
  const [themeMenuOpen, setThemeMenuOpen] = useState(false);
  const { theme, setTheme, themeInfo } = useTheme();
  const themeMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!themeMenuOpen) return;
    const handler = (e: MouseEvent) => {
      if (themeMenuRef.current && !themeMenuRef.current.contains(e.target as Node)) {
        setThemeMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [themeMenuOpen]);

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
          <div className="relative" ref={themeMenuRef}>
            <button
              onClick={() => setThemeMenuOpen(!themeMenuOpen)}
              className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors px-2 py-1 rounded-md hover:bg-secondary"
            >
              <span>{themeInfo.name}</span>
            </button>
            {themeMenuOpen && (
              <div className="absolute right-0 top-full mt-1 w-36 bg-card border border-border rounded-lg shadow-lg py-1 z-50">
                {THEMES.map((t) => (
                  <button
                    key={t.id}
                    onClick={() => { setTheme(t.id as any); setThemeMenuOpen(false); }}
                    className={`w-full text-left px-3 py-1.5 text-xs transition-colors ${
                      theme === t.id
                        ? "text-primary bg-primary/10"
                        : "text-foreground hover:bg-secondary"
                    }`}
                  >
                    {t.name}
                  </button>
                ))}
              </div>
            )}
          </div>
        </header>

        <header className="hidden lg:flex items-center justify-between px-6 py-2 border-b border-border bg-card/80 backdrop-blur-sm">
          <div />
          <div className="flex items-center gap-4">
            <div className="relative" ref={themeMenuRef}>
              <button
                onClick={() => setThemeMenuOpen(!themeMenuOpen)}
                className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors px-2 py-1 rounded-md hover:bg-secondary"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="13.5" cy="6.5" r=".5" fill="currentColor"/><circle cx="17.5" cy="10.5" r=".5" fill="currentColor"/><circle cx="8.5" cy="7.5" r=".5" fill="currentColor"/><circle cx="6.5" cy="12.5" r=".5" fill="currentColor"/><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.555C21.965 6.012 17.461 2 12 2z"/></svg>
                <span>{themeInfo.name}</span>
              </button>
              {themeMenuOpen && (
                <div className="absolute right-0 top-full mt-1 w-36 bg-card border border-border rounded-lg shadow-lg py-1 z-50">
                  {THEMES.map((t) => (
                    <button
                      key={t.id}
                      onClick={() => { setTheme(t.id as any); setThemeMenuOpen(false); }}
                      className={`w-full text-left px-3 py-1.5 text-xs transition-colors ${
                        theme === t.id
                          ? "text-primary bg-primary/10"
                          : "text-foreground hover:bg-secondary"
                      }`}
                    >
                      {t.name}
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="flex items-center gap-2 text-xs">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
              <span className="text-emerald-400">运行中</span>
            </div>
          </div>
        </header>

        <main className="flex-1 overflow-y-auto p-6">{children}</main>
      </div>
    </div>
  );
}
