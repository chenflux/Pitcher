import { useEffect, useState, useCallback } from "react";
import { ComposedChart, Legend, Line, ResponsiveContainer, Tooltip, XAxis, YAxis, Bar } from "recharts";
import { logApi, nodeApi, serviceApi } from "../api";
import type { StatsSummary, TrendData, Node, HoneypotService } from "../types";
import { useWebSocket, type AttackEventData } from "../context/WebSocketContext";

type TimeRange = "1h" | "1d" | "7d" | "30d";

const timeRangeLabels: Record<TimeRange, string> = {
  "1h": "1小时",
  "1d": "1天",
  "7d": "7天",
  "30d": "30天",
};

const timeRangeHours: Record<TimeRange, number> = {
  "1h": 1,
  "1d": 24,
  "7d": 168,
  "30d": 720,
};

export default function Dashboard() {
  const [stats, setStats] = useState<StatsSummary | null>(null);
  const [trend, setTrend] = useState<TrendData | null>(null);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [services, setServices] = useState<HoneypotService[]>([]);
  const [loading, setLoading] = useState(true);
  const [recentAttacks, setRecentAttacks] = useState<AttackEventData[]>([]);
  const [showAlert, setShowAlert] = useState(false);
  const [timeRange, setTimeRange] = useState<TimeRange>("7d");
  const { lastEvent, subscribe, unsubscribe } = useWebSocket();

  useEffect(() => {
    subscribe(["new_attack", "stats_update"]);
    return () => unsubscribe();
  }, [subscribe, unsubscribe]);

  useEffect(() => {
    if (lastEvent?.type === "new_attack") {
      const attack = lastEvent.data as AttackEventData;
      setRecentAttacks(prev => [attack, ...prev].slice(0, 10));
      setShowAlert(true);
      setTimeout(() => setShowAlert(false), 5000);
    }
    if (lastEvent?.type === "stats_update") {
      setStats(prev => prev ? {
        ...prev,
        total_requests: (lastEvent.data as any).total_requests ?? prev.total_requests,
        attack_count: (lastEvent.data as any).attack_count ?? prev.attack_count,
      } : prev);
    }
  }, [lastEvent]);

  const loadData = useCallback(async () => {
    try {
      const hours = timeRangeHours[timeRange];
      const [statsRes, trendRes, nodesRes, servicesRes] = await Promise.all([
        logApi.stats(hours),
        logApi.trend(undefined, timeRange),
        nodeApi.list(),
        serviceApi.list(),
      ]);
      if (statsRes.data.success) setStats(statsRes.data.stats);
      if (trendRes.data.success) setTrend(trendRes.data.trend_data);
      if (nodesRes.data.success) setNodes(nodesRes.data.nodes);
      if (servicesRes.data.success) setServices(servicesRes.data.services);
    } catch {} finally {
      setLoading(false);
    }
  }, [timeRange]);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [loadData]);

  const trendChartData = trend
    ? trend.labels.map((label, i) => ({
        time: label,
        访问次数: trend.data[i] || 0,
        攻击次数: trend.attacks_data[i] || 0,
      }))
    : [];

  const onlineNodes = nodes.filter((n) => n.status === "online").length;
  const runningServices = services.filter((s) => s.status === "running").length;

  if (loading) return <DashboardSkeleton />;

  return (
    <div className="space-y-6">
      {showAlert && recentAttacks[0] && (
        <div className="fixed top-4 right-4 z-50 bg-red-600 text-white px-6 py-4 rounded-xl shadow-2xl animate-pulse flex items-center gap-3 max-w-sm">
          <span className="text-2xl">🚨</span>
          <div>
            <div className="font-semibold">实时攻击告警</div>
            <div className="text-sm opacity-90">{recentAttacks[0].client_ip} → {recentAttacks[0].path}</div>
            <div className="text-xs opacity-75 mt-1">{recentAttacks[0].attack_type}</div>
          </div>
          <button onClick={() => setShowAlert(false)} className="ml-2 opacity-75 hover:opacity-100">✕</button>
        </div>
      )}

      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">蜜罐仪表盘</h1>
          <p className="text-muted-foreground mt-1">实时监控蜜罐状态和活动</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex rounded-lg border border-border overflow-hidden">
            {(["1h", "1d", "7d", "30d"] as TimeRange[]).map((r) => (
              <button
                key={r}
                onClick={() => setTimeRange(r)}
                className={`px-3 py-1.5 text-xs font-medium transition ${
                  timeRange === r
                    ? "bg-primary text-primary-foreground"
                    : "bg-card text-muted-foreground hover:bg-accent"
                }`}
              >
                {timeRangeLabels[r]}
              </button>
            ))}
          </div>
          <button onClick={loadData} className="px-4 py-2 bg-card border border-border rounded-lg hover:bg-accent text-sm text-muted-foreground transition">
            刷新
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard title="在线节点" value={onlineNodes} total={nodes.length} icon="🖥️" color="blue" />
        <StatCard title="运行服务" value={runningServices} total={services.length} icon="🔧" color="green" />
        <StatCard title="总请求数" value={stats?.total_requests || 0} icon="📊" color="indigo" />
        <StatCard title="攻击次数" value={stats?.attack_count || 0} icon="⚠️" color="red" />
      </div>

      <div className="bg-card rounded-xl shadow-sm p-6">
        <h2 className="text-lg font-semibold mb-4 text-foreground">访问趋势 ({timeRangeLabels[timeRange]})</h2>
        {trendChartData.length > 0 ? (
          <ResponsiveContainer width="100%" height={300}>
            <ComposedChart data={trendChartData}>
              <XAxis dataKey="time" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} yAxisId="left" />
              <YAxis tick={{ fontSize: 12 }} yAxisId="right" orientation="right" />
              <Tooltip />
              <Legend />
              <Bar yAxisId="right" dataKey="攻击次数" fill="#ef4444" radius={[4, 4, 0, 0]} />
              <Line yAxisId="left" type="monotone" dataKey="访问次数" stroke="#6366f1" strokeWidth={2} dot={{ r: 3 }} />
            </ComposedChart>
          </ResponsiveContainer>
        ) : (
          <div className="h-[300px] flex items-center justify-center text-muted-foreground">暂无趋势数据</div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {stats && stats.by_attack_type && Object.keys(stats.by_attack_type).length > 0 && (
          <div className="bg-card rounded-xl shadow-sm p-6">
            <h2 className="text-lg font-semibold mb-4 text-foreground">攻击类型分布</h2>
            <div className="space-y-2">
              {Object.entries(stats.by_attack_type)
                .sort((a, b) => (b[1] as number) - (a[1] as number))
                .map(([type, count]) => {
                  const max = Math.max(...Object.values(stats.by_attack_type).map(Number));
                  const pct = max > 0 ? ((count as number) / max) * 100 : 0;
                  return (
                    <div key={type} className="flex items-center gap-3">
                      <span className="text-sm text-muted-foreground w-32 truncate">{type}</span>
                      <div className="flex-1 bg-muted rounded-full h-6 relative overflow-hidden">
                        <div className="bg-red-400 h-full rounded-full transition-all" style={{ width: `${pct}%` }} />
                        <span className="absolute inset-0 flex items-center justify-center text-xs font-medium text-foreground">
                          {count as number}
                        </span>
                      </div>
                    </div>
                  );
                })}
            </div>
          </div>
        )}

        {stats && stats.top_ips && stats.top_ips.length > 0 && (
          <div className="bg-card rounded-xl shadow-sm p-6">
            <h2 className="text-lg font-semibold mb-4 text-foreground">Top 攻击来源 IP</h2>
            <div className="space-y-2">
              {stats.top_ips.slice(0, 10).map((item, i) => (
                <div key={item.ip} className="flex items-center justify-between py-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground w-5">{i + 1}</span>
                    <span className="text-sm font-mono text-foreground">{item.ip}</span>
                  </div>
                  <span className="text-sm font-semibold text-red-500">{item.count}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {recentAttacks.length > 0 && (
        <div className="bg-card rounded-xl shadow-sm p-6 border-l-4 border-red-500">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2 text-foreground">
            <span className="w-2 h-2 bg-red-500 rounded-full animate-pulse"></span>
            实时攻击
          </h2>
          <AttackTable attacks={recentAttacks} />
        </div>
      )}

      {stats && stats.recent_attacks && stats.recent_attacks.length > 0 && (
        <div className="bg-card rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold mb-4 text-foreground">最近攻击 (历史)</h2>
          <AttackTable attacks={stats.recent_attacks.map(a => ({
            id: a.id, client_ip: a.client_ip, path: a.path,
            attack_type: a.attack_type, attack_detail: a.attack_detail,
            created_at: a.created_at,
          }))} />
        </div>
      )}
    </div>
  );
}

function AttackTable({ attacks }: { attacks: Array<{ id: number; client_ip: string; path: string; attack_type: string; attack_detail: string; created_at: string }> }) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-border">
        <thead className="bg-muted">
          <tr>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">时间</th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">来源IP</th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">路径</th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">攻击类型</th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground">详情</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {attacks.map((a) => (
            <tr key={a.id} className="hover:bg-accent transition-colors">
              <td className="px-4 py-2 text-sm text-muted-foreground whitespace-nowrap">
                {new Date(a.created_at).toLocaleString()}
              </td>
              <td className="px-4 py-2 text-sm font-mono text-foreground">{a.client_ip}</td>
              <td className="px-4 py-2 text-sm text-muted-foreground max-w-xs truncate">{a.path}</td>
              <td className="px-4 py-2">
                <span className="px-2 py-0.5 text-xs bg-red-100 text-red-700 rounded-full dark:bg-red-900/30 dark:text-red-400">
                  {a.attack_type}
                </span>
              </td>
              <td className="px-4 py-2 text-xs text-muted-foreground max-w-xs truncate">{a.attack_detail || "-"}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

const colorMap: Record<string, { bg: string; text: string; light: string }> = {
  blue: { bg: "bg-blue-500", text: "text-blue-600", light: "bg-blue-50 dark:bg-blue-900/30" },
  green: { bg: "bg-green-500", text: "text-green-600", light: "bg-green-50 dark:bg-green-900/30" },
  indigo: { bg: "bg-indigo-500", text: "text-indigo-600", light: "bg-indigo-50 dark:bg-indigo-900/30" },
  red: { bg: "bg-red-500", text: "text-red-600", light: "bg-red-50 dark:bg-red-900/30" },
};

function StatCard({ title, value, total, icon, color }: any) {
  const c = colorMap[color] || colorMap.blue;
  return (
    <div className="bg-card rounded-xl shadow-sm p-5 flex items-center gap-4">
      <div className={`w-12 h-12 ${c.light} rounded-xl flex items-center justify-center text-2xl`}>
        {icon}
      </div>
      <div>
        <div className="text-sm text-muted-foreground">{title}</div>
        <div className={`text-2xl font-bold ${c.text}`}>{value}</div>
        {total !== undefined && (
          <div className="text-xs text-muted-foreground">共 {total} 个</div>
        )}
      </div>
    </div>
  );
}

function Skeleton({ className }: { className?: string }) {
  return <div className={`animate-pulse bg-muted rounded ${className || ""}`} />;
}

function DashboardSkeleton() {
  return (
    <div className="space-y-6">
      <div>
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-4 w-64 mt-2" />
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-card rounded-xl shadow-sm p-5 flex items-center gap-4">
            <Skeleton className="w-12 h-12 rounded-xl" />
            <div className="flex-1">
              <Skeleton className="h-3 w-16 mb-2" />
              <Skeleton className="h-7 w-12" />
            </div>
          </div>
        ))}
      </div>
      <div className="bg-card rounded-xl shadow-sm p-6">
        <Skeleton className="h-5 w-40 mb-4" />
        <Skeleton className="h-[300px] w-full" />
      </div>
    </div>
  );
}