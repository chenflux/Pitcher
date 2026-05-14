import { useState, useEffect } from "react";
import api from "../api";

interface SecurityEntry {
  id: number;
  ip_address: string;
  list_type: string;
  reason: string;
  expires_at: string | null;
  created_by: string;
  created_at: string;
}

export default function Security() {
  const [entries, setEntries] = useState<SecurityEntry[]>([]);
  const [stats, setStats] = useState({ blacklist: 0, whitelist: 0, graylist: 0 });
  const [filter, setFilter] = useState("");
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [newEntry, setNewEntry] = useState({ ip_address: "", list_type: "blacklist", reason: "", duration_minutes: 0 });

  useEffect(() => {
    fetchEntries();
    fetchStats();
  }, [filter]);

  const fetchEntries = async () => {
    setLoading(true);
    try {
      const params = filter ? { list_type: filter } : {};
      const res = await api.get("/security/entries", { params });
      setEntries(res.data.entries || []);
    } catch (e) {
      console.error("Failed to fetch entries", e);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const res = await api.get("/security/stats");
      setStats(res.data.stats || { blacklist: 0, whitelist: 0, graylist: 0 });
    } catch (e) {
      console.error("Failed to fetch stats", e);
    }
  };

  const handleAddEntry = async () => {
    if (!newEntry.ip_address) return;
    try {
      await api.post("/security/entries", newEntry);
      setShowModal(false);
      setNewEntry({ ip_address: "", list_type: "blacklist", reason: "", duration_minutes: 0 });
      fetchEntries();
      fetchStats();
    } catch (e) {
      console.error("Failed to add entry", e);
    }
  };

  const handleRemoveEntry = async (ip: string, listType: string) => {
    try {
      await api.delete("/security/entries", { data: { ip_address: ip, list_type: listType } });
      fetchEntries();
      fetchStats();
    } catch (e) {
      console.error("Failed to remove entry", e);
    }
  };

  const getListTypeBadge = (type: string) => {
    const colors = {
      blacklist: "bg-red-500/10 text-red-500 border-red-500/30",
      whitelist: "bg-green-500/10 text-green-500 border-green-500/30",
      graylist: "bg-yellow-500/10 text-yellow-500 border-yellow-500/30",
    };
    const labels = { blacklist: "黑名单", whitelist: "白名单", graylist: "灰名单" };
    return (
      <span className={`px-2 py-0.5 rounded text-xs border ${colors[type as keyof typeof colors] || ""}`}>
        {labels[type as keyof typeof labels] || type}
      </span>
    );
  };

  const formatExpiry = (exp: string | null) => {
    if (!exp) return "永久";
    const d = new Date(exp);
    const now = new Date();
    const diff = d.getTime() - now.getTime();
    if (diff < 0) return "已过期";
    const mins = Math.floor(diff / 60000);
    if (mins < 60) return `${mins}分钟后`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours}小时后`;
    return `${Math.floor(hours / 24)}天后`;
  };

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">安全管理</h1>
          <p className="text-muted-foreground text-sm mt-1">管理 IP 黑名单、白名单和灰名单</p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="bg-primary text-primary-foreground px-4 py-2 rounded-lg hover:opacity-90 transition"
        >
          + 添加规则
        </button>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-card rounded-lg p-4 border">
          <div className="text-3xl font-bold text-red-500">{stats.blacklist}</div>
          <div className="text-sm text-muted-foreground mt-1">黑名单</div>
        </div>
        <div className="bg-card rounded-lg p-4 border">
          <div className="text-3xl font-bold text-green-500">{stats.whitelist}</div>
          <div className="text-sm text-muted-foreground mt-1">白名单</div>
        </div>
        <div className="bg-card rounded-lg p-4 border">
          <div className="text-3xl font-bold text-yellow-500">{stats.graylist}</div>
          <div className="text-sm text-muted-foreground mt-1">灰名单</div>
        </div>
      </div>

      <div className="flex gap-2 mb-4">
        {["", "blacklist", "whitelist", "graylist"].map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-4 py-1.5 rounded-lg text-sm transition ${
              filter === f
                ? "bg-primary text-primary-foreground"
                : "bg-muted text-muted-foreground hover:bg-muted/80"
            }`}
          >
            {f === "" ? "全部" : f === "blacklist" ? "黑名单" : f === "whitelist" ? "白名单" : "灰名单"}
          </button>
        ))}
      </div>

      <div className="bg-card rounded-lg border">
        <table className="w-full">
          <thead>
            <tr className="border-b">
              <th className="text-left p-4 text-sm font-medium text-muted-foreground">IP地址</th>
              <th className="text-left p-4 text-sm font-medium text-muted-foreground">类型</th>
              <th className="text-left p-4 text-sm font-medium text-muted-foreground">原因</th>
              <th className="text-left p-4 text-sm font-medium text-muted-foreground">过期时间</th>
              <th className="text-left p-4 text-sm font-medium text-muted-foreground">创建者</th>
              <th className="text-right p-4 text-sm font-medium text-muted-foreground">操作</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} className="p-8 text-center text-muted-foreground">加载中...</td>
              </tr>
            ) : entries.length === 0 ? (
              <tr>
                <td colSpan={6} className="p-8 text-center text-muted-foreground">暂无数据</td>
              </tr>
            ) : (
              entries.map((entry) => (
                <tr key={entry.id} className="border-b hover:bg-muted/30 transition">
                  <td className="p-4 font-mono text-sm">{entry.ip_address}</td>
                  <td className="p-4">{getListTypeBadge(entry.list_type)}</td>
                  <td className="p-4 text-sm text-muted-foreground">{entry.reason || "-"}</td>
                  <td className="p-4 text-sm">{formatExpiry(entry.expires_at)}</td>
                  <td className="p-4 text-sm">{entry.created_by}</td>
                  <td className="p-4 text-right">
                    <button
                      onClick={() => handleRemoveEntry(entry.ip_address, entry.list_type)}
                      className="text-red-500 hover:text-red-400 text-sm transition"
                    >
                      删除
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card rounded-xl p-6 w-full max-w-md border shadow-2xl">
            <h2 className="text-xl font-bold mb-4">添加IP规则</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1.5">IP地址</label>
                <input
                  type="text"
                  value={newEntry.ip_address}
                  onChange={(e) => setNewEntry({ ...newEntry, ip_address: e.target.value })}
                  className="w-full px-4 py-2 border border-input rounded-lg bg-background"
                  placeholder="如: 192.168.1.100"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5">名单类型</label>
                <select
                  value={newEntry.list_type}
                  onChange={(e) => setNewEntry({ ...newEntry, list_type: e.target.value })}
                  className="w-full px-4 py-2 border border-input rounded-lg bg-background"
                >
                  <option value="blacklist">黑名单 - 无法访问</option>
                  <option value="whitelist">白名单 - 无需验证码</option>
                  <option value="graylist">灰名单 - 限制访问</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5">原因</label>
                <input
                  type="text"
                  value={newEntry.reason}
                  onChange={(e) => setNewEntry({ ...newEntry, reason: e.target.value })}
                  className="w-full px-4 py-2 border border-input rounded-lg bg-background"
                  placeholder="可选"
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1.5">有效期（分钟，0=永久）</label>
                <input
                  type="number"
                  value={newEntry.duration_minutes}
                  onChange={(e) => setNewEntry({ ...newEntry, duration_minutes: parseInt(e.target.value) || 0 })}
                  className="w-full px-4 py-2 border border-input rounded-lg bg-background"
                  placeholder="0"
                />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button
                onClick={() => setShowModal(false)}
                className="flex-1 px-4 py-2 border border-input rounded-lg hover:bg-muted transition"
              >
                取消
              </button>
              <button
                onClick={handleAddEntry}
                className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition"
              >
                添加
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}