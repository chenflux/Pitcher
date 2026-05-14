import { useState, useEffect, useCallback } from "react";
import { authApi, settingsApi } from "../api";
import { useToast } from "../components/Toast";
import { useTheme, THEMES, type ThemeName } from "../context/ThemeContext";

interface AgentDownload {
  filename: string;
  os: string;
  arch: string;
  size: number;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

const osIcon: Record<string, string> = {
  windows: "🪟",
  linux: "🐧",
  darwin: "🍎",
};

const osLabel: Record<string, string> = {
  windows: "Windows",
  linux: "Linux",
  darwin: "macOS",
};

function buildDeployCmd(filename: string, token: string, serverHost: string = "YOUR_SERVER_IP") {
  const downloadUrl = `/api/downloads/agent/${filename}`;
  const envVars = `HONEYWATCH_SERVER=http://${serverHost}:8090 HONEYWATCH_AGENT_TOKEN=${token}`;
  if (filename.includes("windows")) {
    return `curl -sSL ${downloadUrl} -o honeywatch-agent.exe\nset ${envVars}\nhoneywatch-agent.exe`;
  }
  return `curl -sSL ${downloadUrl} -o honeywatch-agent && chmod +x honeywatch-agent\n${envVars} ./honeywatch-agent`;
}

export default function Settings() {
  const [oldPwd, setOldPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [confirmPwd, setConfirmPwd] = useState("");
  const [pwdMsg, setPwdMsg] = useState("");
  const [pwdError, setPwdError] = useState("");
  const [agentToken, setAgentToken] = useState("");
  const [tokenLoading, setTokenLoading] = useState(true);
  const [agents, setAgents] = useState<AgentDownload[]>([]);
  const [serverVersion, setServerVersion] = useState("");
  const [refreshing, setRefreshing] = useState(false);
  const { theme, setTheme, themeInfo } = useTheme();
  const { toast } = useToast();

  const loadData = useCallback(async () => {
    try {
      const [verRes, dlRes] = await Promise.all([
        settingsApi.getVersion(),
        settingsApi.getAgentDownloads(),
      ]);
      if (verRes.data.version) setServerVersion(verRes.data.version);
      if (dlRes.data.agents) setAgents(dlRes.data.agents);
    } catch {}
  }, []);

  const handleRefresh = async () => {
    setRefreshing(true);
    await loadData();
    setRefreshing(false);
    toast("success", "已刷新");
  };

  useEffect(() => {
    (async () => {
      try {
        const res = await settingsApi.getAgentToken();
        if (res.data.success) setAgentToken(res.data.token);
      } catch {
        toast("error", "加载Agent Token失败");
      } finally {
        setTokenLoading(false);
      }
    })();
    loadData();
  }, [loadData, toast]);

  const regenerateToken = async () => {
    if (!confirm("确定要重新生成Agent Token吗？现有所有Agent需要更新配置才能继续通信。")) return;
    try {
      const res = await settingsApi.regenerateAgentToken();
      if (res.data.success) {
        setAgentToken(res.data.token);
        toast("success", "Token已重新生成，请更新所有Agent配置");
      }
    } catch {
      toast("error", "重新生成Token失败");
    }
  };

  const changePassword = async () => {
    setPwdError("");
    setPwdMsg("");
    if (newPwd !== confirmPwd) {
      setPwdError("两次输入的密码不一致");
      return;
    }
    if (newPwd.length < 6) {
      setPwdError("密码长度不能少于6位");
      return;
    }
    try {
      await authApi.changePassword(oldPwd, newPwd);
      setPwdMsg("密码修改成功");
      setOldPwd("");
      setNewPwd("");
      setConfirmPwd("");
    } catch (err: any) {
      setPwdError(err.response?.data?.error || "密码修改失败");
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-foreground">系统设置</h1>
        <p className="text-sm text-muted-foreground mt-1">管理账户安全、通信密钥与 Agent 部署</p>
      </div>

      <div className="bg-card rounded-xl border border-border p-6 max-w-2xl">
        <h2 className="text-base font-semibold mb-4 text-foreground">修改密码</h2>
        {pwdMsg && <div className="bg-emerald-500/10 text-emerald-400 px-4 py-2 rounded-lg text-sm mb-4">{pwdMsg}</div>}
        {pwdError && <div className="bg-red-500/10 text-red-400 px-4 py-2 rounded-lg text-sm mb-4">{pwdError}</div>}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">当前密码</label>
            <input type="password" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground text-sm focus:outline-none focus:ring-1 focus:ring-ring" />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">新密码</label>
            <input type="password" value={newPwd} onChange={(e) => setNewPwd(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground text-sm focus:outline-none focus:ring-1 focus:ring-ring" />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">确认新密码</label>
            <input type="password" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-lg bg-background text-foreground text-sm focus:outline-none focus:ring-1 focus:ring-ring" />
          </div>
          <button onClick={changePassword} className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm hover:opacity-90 transition-opacity">
            修改密码
          </button>
        </div>
      </div>

      <div className="bg-card rounded-xl border border-border p-6 max-w-2xl">
        <h2 className="text-base font-semibold mb-2 text-foreground">Agent 通信密钥</h2>
        <p className="text-xs text-muted-foreground mb-4">此 Token 用于 Agent 与 Server 数据面(8090)通信认证，请妥善保管。</p>
        {tokenLoading ? (
          <div className="text-sm text-muted-foreground">加载中...</div>
        ) : (
          <>
            <div className="flex gap-2 items-center">
              <code className="flex-1 bg-muted px-3 py-2 rounded-lg text-sm font-mono break-all text-foreground">{agentToken}</code>
              <button onClick={() => { navigator.clipboard.writeText(agentToken); toast("success", "已复制"); }}
                className="px-3 py-2 bg-secondary text-secondary-foreground rounded-lg text-sm hover:opacity-80 shrink-0">复制</button>
            </div>
            <button onClick={regenerateToken} className="mt-3 px-4 py-2 bg-orange-500/10 text-orange-400 border border-orange-500/30 rounded-lg text-sm hover:bg-orange-500/20 transition-colors">
              重新生成密钥
            </button>
          </>
        )}
      </div>

      <div className="bg-card rounded-xl border border-border p-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-base font-semibold text-foreground">Agent 下载部署</h2>
            <p className="text-xs text-muted-foreground mt-1">
              Server 版本：<span className="font-mono text-foreground">{serverVersion || "-"}</span>
            </p>
          </div>
          <button onClick={handleRefresh} disabled={refreshing}
            className="px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground hover:bg-secondary rounded-lg transition-colors">
            {refreshing ? "刷新中..." : "刷新"}
          </button>
        </div>
        <p className="text-xs text-muted-foreground mb-4">
          下载后设置环境变量 <code className="bg-muted px-1.5 py-0.5 rounded text-foreground">HONEYWATCH_SERVER</code> 和 <code className="bg-muted px-1.5 py-0.5 rounded text-foreground">HONEYWATCH_AGENT_TOKEN</code> 即可运行。
        </p>

        {agents.length === 0 ? (
          <div className="text-center py-10 text-muted-foreground">
            <p className="text-2xl mb-2">📦</p>
            <p className="text-sm">暂无可用 Agent 二进制文件</p>
            <p className="text-xs mt-1">请运行 build.bat 编译 Agent 后刷新此页面</p>
          </div>
        ) : (
          <div className="space-y-2">
            <div className="grid grid-cols-[auto_1fr_auto_auto_auto] gap-x-4 gap-y-1 px-4 py-2 text-xs font-semibold text-muted-foreground border-b border-border">
              <span className="w-6"></span>
              <span>文件</span>
              <span className="w-20 text-right">大小</span>
              <span className="w-28 text-center">操作</span>
              <span className="w-20 text-center">SHA256</span>
            </div>
            {agents.map((a) => {
              const deployCmd = buildDeployCmd(a.filename, agentToken || "YOUR_TOKEN");
              const downloadUrl = `/api/downloads/agent/${a.filename}`;
              return (
                <div key={a.filename} className="grid grid-cols-[auto_1fr_auto_auto_auto] gap-x-4 gap-y-1 items-start px-4 py-3 rounded-lg hover:bg-secondary/50 transition-colors border border-transparent hover:border-border">
                  <span className="text-xl w-6 pt-0.5">{osIcon[a.os] || "🖥️"}</span>
                  <div className="min-w-0">
                    <div className="text-sm font-medium text-foreground">{osLabel[a.os] || a.os} / {a.arch}</div>
                    <div className="text-xs text-muted-foreground font-mono truncate">{a.filename}</div>
                    <div className="mt-2 bg-muted rounded-lg px-3 py-2 text-xs font-mono text-muted-foreground whitespace-pre-wrap break-all leading-relaxed">{deployCmd}</div>
                  </div>
                  <div className="text-xs text-muted-foreground pt-1 w-20 text-right shrink-0">{formatBytes(a.size)}</div>
                  <div className="flex gap-1 pt-1 w-28 justify-center shrink-0">
                    <button
                      onClick={() => {
                        const link = document.createElement("a");
                        link.href = downloadUrl;
                        link.download = a.filename;
                        document.body.appendChild(link);
                        link.click();
                        document.body.removeChild(link);
                      }}
                      className="px-2.5 py-1 bg-primary text-primary-foreground rounded-lg text-xs hover:opacity-90 transition-opacity"
                    >
                      下载
                    </button>
                    <button
                      onClick={() => {
                        navigator.clipboard.writeText(deployCmd);
                        toast("success", "部署命令已复制");
                      }}
                      className="px-2.5 py-1 bg-secondary text-secondary-foreground rounded-lg text-xs hover:opacity-80 transition-opacity"
                    >
                      复制
                    </button>
                  </div>
                  <div className="text-xs text-muted-foreground pt-1 w-28 text-center shrink-0 font-mono">-</div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="bg-card rounded-xl border border-border p-6 max-w-2xl">
        <h2 className="text-base font-semibold mb-2 text-foreground">外观主题</h2>
        <p className="text-xs text-muted-foreground mb-4">当前：<span className="font-medium text-foreground">{themeInfo.name}</span></p>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {THEMES.map((t) => (
            <button
              key={t.id}
              onClick={() => setTheme(t.id as ThemeName)}
              className={`p-3 rounded-lg border-2 transition-all text-left ${
                theme === t.id ? "border-primary" : "border-border hover:border-primary/30"
              }`}
              style={{ backgroundColor: t.preview.bg }}
            >
              <div className="flex items-center gap-2 mb-2">
                <div className="w-4 h-4 rounded-full" style={{ backgroundColor: t.preview.primary }} />
                <span className="text-sm font-medium" style={{ color: t.preview.fg }}>{t.name}</span>
              </div>
              <div className="text-xs" style={{ color: t.preview.fg, opacity: 0.7 }}>{t.desc}</div>
              <div className="mt-2 h-3 rounded" style={{ backgroundColor: t.preview.card }} />
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
