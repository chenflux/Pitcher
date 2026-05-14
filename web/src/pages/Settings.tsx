import { useState, useEffect, useCallback } from "react";
import { authApi, settingsApi } from "../api";
import { useToast } from "../components/Toast";

interface AgentDownload {
  filename: string;
  os: string;
  arch: string;
  size: number;
  sha256: string;
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

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div className="bg-card rounded-xl border border-border p-6">
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

        <div className="bg-card rounded-xl border border-border p-6">
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
          <div className="border border-border rounded-lg overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-muted/50 border-b border-border">
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-muted-foreground w-8"></th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-muted-foreground">平台</th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-muted-foreground">文件名</th>
                  <th className="text-right px-4 py-2.5 text-xs font-semibold text-muted-foreground w-20">大小</th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-muted-foreground w-48">SHA256</th>
                  <th className="text-center px-4 py-2.5 text-xs font-semibold text-muted-foreground w-36">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {agents.map((a) => {
                  const deployCmd = buildDeployCmd(a.filename, agentToken || "YOUR_TOKEN");
                  const downloadUrl = `/api/downloads/agent/${a.filename}`;
                  return (
                    <tr key={a.filename} className="hover:bg-secondary/30 transition-colors">
                      <td className="px-4 py-3 text-lg text-center">{osIcon[a.os] || "🖥️"}</td>
                      <td className="px-4 py-3">
                        <div className="font-medium text-foreground">{osLabel[a.os] || a.os}</div>
                        <div className="text-xs text-muted-foreground">{a.arch}</div>
                      </td>
                      <td className="px-4 py-3">
                        <div className="font-mono text-xs text-foreground truncate max-w-xs">{a.filename}</div>
                        <div className="mt-1.5 bg-muted rounded px-2 py-1.5 text-xs font-mono text-muted-foreground whitespace-pre-wrap break-all leading-relaxed">{deployCmd}</div>
                      </td>
                      <td className="px-4 py-3 text-right text-muted-foreground text-xs whitespace-nowrap">{formatBytes(a.size)}</td>
                      <td className="px-4 py-3 font-mono text-xs text-muted-foreground truncate max-w-[12rem]" title={a.sha256}>{a.sha256 || "-"}</td>
                      <td className="px-4 py-3">
                        <div className="flex gap-1.5 justify-center">
                          <button
                            onClick={() => {
                              const link = document.createElement("a");
                              link.href = downloadUrl;
                              link.download = a.filename;
                              document.body.appendChild(link);
                              link.click();
                              document.body.removeChild(link);
                            }}
                            className="px-3 py-1.5 bg-primary text-primary-foreground rounded-lg text-xs hover:opacity-90 transition-opacity whitespace-nowrap"
                          >
                            下载
                          </button>
                          <button
                            onClick={() => {
                              navigator.clipboard.writeText(deployCmd);
                              toast("success", "部署命令已复制");
                            }}
                            className="px-3 py-1.5 bg-secondary text-secondary-foreground rounded-lg text-xs hover:opacity-80 transition-opacity whitespace-nowrap"
                          >
                            复制命令
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
