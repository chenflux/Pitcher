import { useEffect, useState } from "react";
import { logApi } from "../api";
import type { RequestLog } from "../types";
import { HoneypotTypeMap } from "../types";
import { useToast } from "../components/Toast";

export default function Logs() {
  const [logs, setLogs] = useState<RequestLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [isAttack, setIsAttack] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    loadLogs();
  }, [page, isAttack]);

  const loadLogs = async () => {
    setLoading(true);
    try {
      const params: any = { page, page_size: 20 };
      if (isAttack !== "") params.is_attack = isAttack;
      const res = await logApi.query(params);
      if (res.data.success) {
        setLogs(res.data.logs || []);
        setTotal(res.data.total || 0);
      }
    } catch {
      toast("error", "加载日志失败");
    } finally {
      setLoading(false);
    }
  };

  const totalPages = Math.ceil(total / 20);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">日志查看</h1>
          <p className="text-muted-foreground mt-1">查看蜜罐访问日志和攻击记录</p>
        </div>
        <div className="flex gap-2">
          <select value={isAttack} onChange={(e) => { setIsAttack(e.target.value); setPage(1); }}
            className="px-3 py-2 border rounded-lg text-sm bg-card">
            <option value="">全部</option>
            <option value="true">仅攻击</option>
            <option value="false">仅正常</option>
          </select>
          <button onClick={loadLogs} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm">
            刷新
          </button>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm overflow-hidden">
        {loading ? (
          <div className="flex flex-col items-center justify-center py-16 gap-3">
            <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />
            <span className="text-sm text-muted-foreground">加载中...</span>
          </div>
        ) : logs.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-4xl mb-3">📋</p>
            <p className="text-muted-foreground">暂无日志记录</p>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-border">
                <thead className="bg-muted">
                  <tr>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground w-8"></th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">时间</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">IP</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">方法</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">路径</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">蜜罐</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">状态</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">攻击</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {logs.map((log) => (
                    <LogRow key={log.id} log={log} expanded={expandedId === log.id} onToggle={() => setExpandedId(expandedId === log.id ? null : log.id)} />
                  ))}
                </tbody>
              </table>
            </div>
            <div className="px-4 py-3 flex items-center justify-between border-t bg-muted">
              <span className="text-sm text-muted-foreground">共 {total} 条记录</span>
              <div className="flex gap-2 items-center">
                <button disabled={page <= 1} onClick={() => setPage(page - 1)}
                  className="px-3 py-1.5 text-sm border rounded-lg bg-card hover:bg-accent disabled:opacity-40 disabled:cursor-not-allowed">
                  上一页
                </button>
                <span className="text-sm text-muted-foreground">{page} / {totalPages || 1}</span>
                <button disabled={page >= totalPages} onClick={() => setPage(page + 1)}
                  className="px-3 py-1.5 text-sm border rounded-lg bg-card hover:bg-accent disabled:opacity-40 disabled:cursor-not-allowed">
                  下一页
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

function LogRow({ log, expanded, onToggle }: { log: RequestLog; expanded: boolean; onToggle: () => void }) {
  return (
    <>
      <tr className={`hover:bg-accent cursor-pointer transition-colors ${expanded ? "bg-blue-50" : ""}`} onClick={onToggle}>
        <td className="px-4 py-2 text-xs text-muted-foreground">{expanded ? "▼" : "▶"}</td>
        <td className="px-4 py-2 text-sm text-muted-foreground whitespace-nowrap">
          {new Date(log.created_at).toLocaleString()}
        </td>
        <td className="px-4 py-2 text-sm font-mono text-foreground">{log.client_ip}</td>
        <td className="px-4 py-2">
          <span className="px-2 py-0.5 text-xs bg-muted text-muted-foreground rounded">{log.method}</span>
        </td>
        <td className="px-4 py-2 text-sm text-muted-foreground max-w-xs truncate">{log.path}</td>
        <td className="px-4 py-2 text-sm text-muted-foreground">{HoneypotTypeMap[log.honeypot_type] || "-"}</td>
        <td className="px-4 py-2 text-sm text-muted-foreground">{log.status_code}</td>
        <td className="px-4 py-2">
          {log.is_attack ? (
            <span className="px-2 py-0.5 text-xs bg-red-100 text-red-700 rounded-full">{log.attack_type}</span>
          ) : (
            <span className="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded-full">正常</span>
          )}
        </td>
      </tr>
      {expanded && (
        <tr>
          <td colSpan={8} className="px-6 py-4 bg-muted border-t border-b">
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
              <DetailItem label="协议" value={log.protocol} />
              <DetailItem label="User-Agent" value={log.user_agent} mono />
              <DetailItem label="攻击详情" value={log.attack_detail} />
              <DetailItem label="节点ID" value={log.node_id} mono />
              <DetailItem label="服务ID" value={String(log.service_id)} />
              <DetailItem label="蜜罐类型" value={`${HoneypotTypeMap[log.honeypot_type] || "Unknown"} (${log.honeypot_type})`} />
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

function DetailItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <div className="text-xs text-muted-foreground mb-0.5">{label}</div>
      <div className={`text-foreground break-all ${mono ? "font-mono text-xs" : ""}`}>{value || "-"}</div>
    </div>
  );
}
