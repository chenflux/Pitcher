import { useState, useEffect } from "react";
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

const installCommands: Record<string, Record<string, string>> = {
  windows: {
    amd64: `set HONEYWATCH_SERVER=http://YOUR_SERVER_IP:8090\nset HONEYWATCH_AGENT_TOKEN=YOUR_TOKEN\nbin\\honeywatch-agent-2.0.0-windows-amd64.exe`,
  },
  linux: {
    amd64: `curl -sSL http://YOUR_SERVER:8080/api/downloads/agent/honeywatch-agent-2.0.0-linux-amd64 -o honeywatch-agent && chmod +x honeywatch-agent && HONEYWATCH_SERVER=http://YOUR_SERVER_IP:8090 HONEYWATCH_AGENT_TOKEN=YOUR_TOKEN ./honeywatch-agent`,
    arm64: `curl -sSL http://YOUR_SERVER:8080/api/downloads/agent/honeywatch-agent-2.0.0-linux-arm64 -o honeywatch-agent && chmod +x honeywatch-agent && HONEYWATCH_SERVER=http://YOUR_SERVER_IP:8090 HONEYWATCH_AGENT_TOKEN=YOUR_TOKEN ./honeywatch-agent`,
  },
  darwin: {
    arm64: `curl -sSL http://YOUR_SERVER:8080/api/downloads/agent/honeywatch-agent-2.0.0-darwin-arm64 -o honeywatch-agent && chmod +x honeywatch-agent && HONEYWATCH_SERVER=http://YOUR_SERVER_IP:8090 HONEYWATCH_AGENT_TOKEN=YOUR_TOKEN ./honeywatch-agent`,
  },
};

export default function Settings() {
  const [oldPwd, setOldPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [confirmPwd, setConfirmPwd] = useState("");
  const [pwdMsg, setPwdMsg] = useState("");
  const [pwdError, setPwdError] = useState("");
  const [agentToken, setAgentToken] = useState("");
  const [tokenLoading, setTokenLoading] = useState(true);
  const [agents, setAgents] = useState<AgentDownload[]>([]);
  const [agentVersion, setAgentVersion] = useState("");
  const { theme, setTheme, themeInfo } = useTheme();
  const { toast } = useToast();

  useEffect(() => {
    loadToken();
    loadVersion();
  }, []);

  const loadToken = async () => {
    try {
      const res = await settingsApi.getAgentToken();
      if (res.data.success) setAgentToken(res.data.token);
    } catch {
      toast("error", "加载Agent Token失败");
    } finally {
      setTokenLoading(false);
    }
  };

  const loadVersion = async () => {
    try {
      const [verRes, dlRes] = await Promise.all([
        settingsApi.getVersion(),
        settingsApi.getAgentDownloads(),
      ]);
      if (verRes.data.version) setAgentVersion(verRes.data.version);
      if (dlRes.data.agents) setAgents(dlRes.data.agents);
    } catch {}
  };

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
        <h1 className="text-3xl font-bold text-foreground">系统设置</h1>
        <p className="text-muted-foreground mt-1">管理用户和系统配置</p>
      </div>

      <div className="bg-card rounded-lg shadow p-6 max-w-2xl">
        <h2 className="text-lg font-semibold mb-4 text-foreground">修改密码</h2>
        {pwdMsg && <div className="bg-green-50 text-green-600 px-4 py-2 rounded text-sm mb-4">{pwdMsg}</div>}
        {pwdError && <div className="bg-red-50 text-red-600 px-4 py-2 rounded text-sm mb-4">{pwdError}</div>}
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">当前密码</label>
            <input type="password" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)}
              className="w-full px-3 py-2 border border-input rounded bg-background text-foreground" />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">新密码</label>
            <input type="password" value={newPwd} onChange={(e) => setNewPwd(e.target.value)}
              className="w-full px-3 py-2 border border-input rounded bg-background text-foreground" />
          </div>
          <div>
            <label className="block text-sm font-medium text-foreground mb-1">确认新密码</label>
            <input type="password" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)}
              className="w-full px-3 py-2 border border-input rounded bg-background text-foreground" />
          </div>
          <button onClick={changePassword} className="px-4 py-2 bg-primary text-primary-foreground rounded hover:opacity-90">
            修改密码
          </button>
        </div>
      </div>

      <div className="bg-card rounded-lg shadow p-6 max-w-2xl">
        <h2 className="text-lg font-semibold mb-4 text-foreground">Agent 通信密钥</h2>
        <p className="text-sm text-muted-foreground mb-4">此 Token 用于 Agent 与 Server 数据面(8090)通信的认证，请妥善保管。</p>
        {tokenLoading ? (
          <div className="text-muted-foreground">加载中...</div>
        ) : (
          <>
            <div className="flex gap-2 items-center mb-3">
              <code className="flex-1 bg-muted px-3 py-2 rounded text-sm font-mono break-all text-foreground">{agentToken}</code>
              <button onClick={() => {navigator.clipboard.writeText(agentToken); toast("success", "已复制");}}
                className="px-3 py-1 bg-secondary text-secondary-foreground rounded text-sm hover:opacity-80">复制</button>
            </div>
            <button onClick={regenerateToken} className="px-4 py-2 bg-orange-500 text-white rounded hover:bg-orange-600 text-sm">
              重新生成密钥
            </button>
            <p className="text-xs text-orange-500 mt-2">警告：重新生成后所有现有 Agent 需要更新 HONEYWATCH_AGENT_TOKEN 环境变量才能通信。</p>
          </>
        )}
      </div>

      <div className="bg-card rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">Agent 下载部署</h2>
          <button onClick={loadVersion} className="text-xs text-muted-foreground hover:text-foreground">
            刷新
          </button>
        </div>
        <p className="text-sm text-muted-foreground mb-1">
          Server 版本：<span className="font-mono text-foreground">{agentVersion || "?"}</span>
        </p>
        <p className="text-xs text-muted-foreground mb-4">下载后设置环境变量 <code className="bg-muted px-1 rounded">HONEYWATCH_SERVER</code> 和 <code className="bg-muted px-1 rounded">HONEYWATCH_AGENT_TOKEN</code> 即可运行。</p>

        <div className="space-y-4">
          {agents.map((a) => (
            <div key={a.filename} className="border border-border rounded-lg p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <span className="text-xl">{osIcon[a.os] || "🖥️"}</span>
                  <div>
                    <div className="font-medium text-foreground">{a.filename}</div>
                    <div className="text-xs text-muted-foreground">
                      {a.os} / {a.arch} · {formatBytes(a.size)}
                    </div>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      const url = `/api/downloads/agent/${a.filename}`;
                      const link = document.createElement("a");
                      link.href = url;
                      link.download = a.filename;
                      document.body.appendChild(link);
                      link.click();
                      document.body.removeChild(link);
                    }}
                    className="px-3 py-1 bg-primary text-primary-foreground rounded text-xs hover:opacity-80"
                  >
                    下载
                  </button>
                  <button
                    onClick={() => {
                      const cmd = installCommands[a.os]?.[a.arch] || `curl -sSL http://YOUR_SERVER:8080/api/downloads/agent/${a.filename} -o agent && chmod +x agent && ./agent`;
                      navigator.clipboard.writeText(cmd);
                      toast("success", "部署命令已复制");
                    }}
                    className="px-3 py-1 bg-secondary text-secondary-foreground rounded text-xs hover:opacity-80"
                  >
                    复制部署命令
                  </button>
                </div>
              </div>
              <div className="mt-2">
                <code className="text-xs bg-muted text-muted-foreground px-2 py-1 rounded block break-all">
                  {installCommands[a.os]?.[a.arch] || `curl -sSL http://YOUR_SERVER:8080/api/downloads/agent/${a.filename} -o honeywatch-agent && chmod +x honeywatch-agent && ./honeywatch-agent`}
                </code>
              </div>
            </div>
          ))}
        </div>

        {agents.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-lg mb-2">📦</p>
            <p className="text-sm">暂无可用 Agent 二进制文件</p>
            <p className="text-xs mt-1">请运行 build.bat 编译 Agent 后自动出现在此处</p>
          </div>
        )}
      </div>

      <div className="bg-card rounded-lg shadow p-6 max-w-2xl">
        <h2 className="text-lg font-semibold mb-4 text-foreground">外观主题</h2>
        <p className="text-sm text-muted-foreground mb-4">选择系统界面主题风格，当前：<span className="font-medium text-foreground">{themeInfo.name}</span></p>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {THEMES.map((t) => (
            <button
              key={t.id}
              onClick={() => setTheme(t.id as ThemeName)}
              className={`p-3 rounded-lg border-2 transition-all text-left ${
                theme === t.id
                  ? "border-primary"
                  : "border-transparent hover:border-border"
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