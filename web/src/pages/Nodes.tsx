import { useEffect, useState } from "react";
import { nodeApi } from "../api";
import type { Node } from "../types";
import { useToast } from "../components/Toast";

export default function Nodes() {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    loadNodes();
    const interval = setInterval(loadNodes, 10000);
    return () => clearInterval(interval);
  }, []);

  const loadNodes = async () => {
    try {
      const res = await nodeApi.list();
      if (res.data.success) {
        setNodes(res.data.nodes);
        if (selectedNode) {
          const updated = res.data.nodes.find((n: Node) => n.id === selectedNode.id);
          if (updated) setSelectedNode(updated);
        }
      }
    } catch {
      toast("error", "加载节点列表失败");
    } finally {
      setLoading(false);
    }
  };

  const deleteNode = async (id: string) => {
    if (!confirm("确定删除该节点?")) return;
    try {
      await nodeApi.delete(id);
      toast("success", "节点已删除");
      if (selectedNode?.id === id) setSelectedNode(null);
      loadNodes();
    } catch {
      toast("error", "删除节点失败");
    }
  };

  const uptime = (lastHeartbeat: string) => {
    const diff = Date.now() - new Date(lastHeartbeat).getTime();
    const mins = Math.floor(diff / 60000);
    if (mins < 60) return `${mins} 分钟前`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `${hours} 小时前`;
    return `${Math.floor(hours / 24)} 天前`;
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">节点管理</h1>
          <p className="text-muted-foreground mt-1">管理和监控所有蜜罐节点</p>
        </div>
        <div className="flex gap-2 items-center">
          <span className="text-sm text-muted-foreground">自动刷新 10s</span>
          <button onClick={loadNodes} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm">
            刷新
          </button>
        </div>
      </div>

      {loading ? (
        <div className="flex flex-col items-center justify-center py-16 gap-3">
          <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />
          <span className="text-sm text-muted-foreground">加载中...</span>
        </div>
      ) : nodes.length === 0 ? (
        <div className="bg-card rounded-xl shadow-sm p-16 text-center">
          <p className="text-5xl mb-4">📡</p>
          <h3 className="text-lg font-medium text-foreground">暂无注册节点</h3>
          <p className="text-muted-foreground mt-2">请部署 Pitcher Agent 以注册节点</p>
        </div>
      ) : (
        <div className="bg-card rounded-xl shadow-sm overflow-hidden">
          <table className="min-w-full divide-y divide-border">
            <thead className="bg-muted">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">状态</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">主机名</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">IP地址</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">系统</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">资源使用</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">Agent</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">最后心跳</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-muted-foreground">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {nodes.map((node) => (
                <tr
                  key={node.id}
                  className={`hover:bg-accent cursor-pointer transition-colors ${selectedNode?.id === node.id ? "bg-blue-50" : ""}`}
                  onClick={() => setSelectedNode(node)}
                >
                  <td className="px-4 py-3">
                    <span className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${
                      node.status === "online" ? "bg-green-100 text-green-700" : "bg-muted text-muted-foreground"
                    }`}>
                      <span className={`w-1.5 h-1.5 rounded-full ${node.status === "online" ? "bg-green-500" : "bg-muted-foreground"}`}></span>
                      {node.status === "online" ? "在线" : "离线"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm font-semibold text-foreground">{node.hostname}</td>
                  <td className="px-4 py-3 text-sm font-mono text-foreground">{node.ip_address}</td>
                  <td className="px-4 py-3 text-sm text-foreground">{node.os_name || "-"} {node.os_version || ""}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-3 text-xs">
                      <ResourceBar label="CPU" value={node.cpu_usage} />
                      <ResourceBar label="MEM" value={node.memory_usage} />
                      <ResourceBar label="DISK" value={node.disk_usage} />
                    </div>
                  </td>
                  <td className="px-4 py-3 text-sm text-muted-foreground">{node.agent_version}</td>
                  <td className="px-4 py-3 text-sm text-muted-foreground whitespace-nowrap">
                    {uptime(node.last_heartbeat)}
                  </td>
                  <td className="px-4 py-3">
                    <button onClick={(e) => { e.stopPropagation(); deleteNode(node.id); }}
                      className="text-red-500 text-xs hover:underline font-medium">删除</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {selectedNode && (
        <div className="fixed inset-0 z-50 flex justify-end" onClick={() => setSelectedNode(null)}>
          <div className="absolute inset-0 bg-black/30" />
          <div className="relative w-full max-w-md bg-card shadow-xl overflow-y-auto" onClick={(e) => e.stopPropagation()}>
            <div className="sticky top-0 bg-card border-b px-6 py-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold">节点详情</h2>
              <button onClick={() => setSelectedNode(null)} className="text-muted-foreground hover:text-foreground text-xl">&times;</button>
            </div>
            <div className="p-6 space-y-6">
              <div className="flex items-center gap-3">
                <div className={`w-12 h-12 rounded-xl flex items-center justify-center ${
                  selectedNode.status === "online" ? "bg-green-100" : "bg-muted"
                }`}>
                  <span className="text-2xl">🖥️</span>
                </div>
                <div>
                  <div className="font-bold text-lg">{selectedNode.hostname}</div>
                  <span className={`px-2 py-0.5 text-xs rounded-full ${
                    selectedNode.status === "online" ? "bg-green-100 text-green-700" : "bg-muted text-muted-foreground"
                  }`}>{selectedNode.status === "online" ? "在线" : "离线"}</span>
                </div>
              </div>

              <div className="space-y-3">
                <DetailRow label="Agent ID" value={selectedNode.agent_id} mono />
                <DetailRow label="IP 地址" value={selectedNode.ip_address} mono />
                <DetailRow label="操作系统" value={`${selectedNode.os_name || "-"} ${selectedNode.os_version || ""}`} />
                <DetailRow label="Agent 版本" value={selectedNode.agent_version} />
                <DetailRow label="注册时间" value={new Date(selectedNode.created_at).toLocaleString()} />
                <DetailRow label="最后心跳" value={new Date(selectedNode.last_heartbeat).toLocaleString()} />
              </div>

              <div>
                <h3 className="text-sm font-semibold text-foreground mb-3">资源使用</h3>
                <div className="space-y-3">
                  <ResourceBarLarge label="CPU" value={selectedNode.cpu_usage} color="blue" />
                  <ResourceBarLarge label="内存" value={selectedNode.memory_usage} color="green" />
                  <ResourceBarLarge label="磁盘" value={selectedNode.disk_usage} color="yellow" />
                </div>
              </div>

              <div className="pt-4 border-t">
                <button onClick={() => deleteNode(selectedNode.id)}
                  className="w-full px-4 py-2 bg-red-50 text-red-600 rounded-lg hover:bg-red-100 text-sm font-medium">
                  删除节点
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function ResourceBar({ label, value }: { label: string; value: number }) {
  const pct = Math.min(value, 100);
  const color = pct > 80 ? "bg-red-400" : pct > 50 ? "bg-yellow-400" : "bg-green-400";
  return (
    <div className="flex items-center gap-1.5 w-24">
      <span className="text-muted-foreground w-8">{label}</span>
      <div className="flex-1 bg-muted rounded-full h-1.5 overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-muted-foreground w-8 text-right">{pct > 0 ? `${pct.toFixed(0)}%` : "-"}</span>
    </div>
  );
}

function ResourceBarLarge({ label, value, color }: { label: string; value: number; color: string }) {
  const pct = Math.min(value, 100);
  const colors: Record<string, string> = {
    blue: "bg-blue-500",
    green: "bg-green-500",
    yellow: "bg-yellow-500",
  };
  return (
    <div>
      <div className="flex justify-between text-sm mb-1">
        <span className="text-foreground">{label}</span>
        <span className="font-medium text-foreground">{pct > 0 ? `${pct.toFixed(1)}%` : "-"}</span>
      </div>
      <div className="bg-muted rounded-full h-2.5 overflow-hidden">
        <div className={`h-full rounded-full transition-all ${colors[color] || "bg-blue-500"}`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

function DetailRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex justify-between items-start py-1">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className={`text-sm text-foreground text-right max-w-[60%] break-all ${mono ? "font-mono text-xs" : ""}`}>{value || "-"}</span>
    </div>
  );
}
