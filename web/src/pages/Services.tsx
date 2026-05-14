import { useEffect, useState } from "react";
import { serviceApi, nodeApi, configApi } from "../api";
import type { HoneypotService, Node, ConfigTemplate } from "../types";
import { HoneypotTypeMap } from "../types";
import { useToast } from "../components/Toast";

const typeDefaultPorts: Record<number, number> = {
  1: 9200, 2: 7001, 3: 3306, 4: 6379, 5: 27017, 6: 80, 7: 8080,
};

export default function Services() {
  const [services, setServices] = useState<HoneypotService[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [configs, setConfigs] = useState<ConfigTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [selectedConfigId, setSelectedConfigId] = useState<number | "">("");
  const [form, setForm] = useState({ node_id: "", name: "", type: 1, host: "0.0.0.0", port: 9200, enabled: true, config: "" });
  const [filterNode, setFilterNode] = useState<string>("");
  const [filterType, setFilterType] = useState<string>("");
  const [filterStatus, setFilterStatus] = useState<string>("");
  const [searchText, setSearchText] = useState<string>("");
  const { toast } = useToast();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [svcRes, nodeRes, cfgRes] = await Promise.all([serviceApi.list(), nodeApi.list(), configApi.list()]);
      if (svcRes.data.success) setServices(svcRes.data.services);
      if (nodeRes.data.success) setNodes(nodeRes.data.nodes);
      if (cfgRes.data.success) setConfigs(cfgRes.data.configs);
    } catch {
      toast("error", "加载数据失败");
    } finally {
      setLoading(false);
    }
  };

  const handleTypeChange = (t: number) => {
    setForm({ ...form, type: t, port: typeDefaultPorts[t] || 8080 });
  };

  const handleConfigChange = (cfgId: number | "") => {
    setSelectedConfigId(cfgId);
    if (cfgId === "") {
      setForm({ ...form, config: "" });
    } else {
      const cfg = configs.find((c) => c.id === cfgId);
      if (cfg) {
        setForm({ ...form, config: cfg.content });
      }
    }
  };

  const createService = async () => {
    if (!form.name) {
      toast("error", "服务名称不能为空");
      return;
    }
    try {
      await serviceApi.create(form);
      toast("success", "服务创建成功");
      setShowCreate(false);
      setSelectedConfigId("");
      setForm({ node_id: "", name: "", type: 1, host: "0.0.0.0", port: 9200, enabled: true, config: "" });
      loadData();
    } catch {
      toast("error", "创建服务失败");
    }
  };

  const deleteService = async (id: number) => {
    if (!confirm("确定删除该服务?")) return;
    try {
      await serviceApi.delete(id);
      toast("success", "服务已删除");
      loadData();
    } catch {
      toast("error", "删除服务失败");
    }
  };

  const actionService = async (action: "start" | "stop" | "restart", id: number) => {
    const labels = { start: "启动", stop: "停止", restart: "重启" };
    try {
      await serviceApi[action](id);
      toast("success", `服务已${labels[action]}`);
      loadData();
    } catch {
      toast("error", `${labels[action]}失败`);
    }
  };

  const getNodeName = (nodeId: string) => {
    const node = nodes.find((n) => n.id === nodeId);
    return node ? `${node.hostname} (${node.ip_address})` : nodeId || "未分配";
  };

  const getNodeIP = (nodeId: string) => {
    const node = nodes.find((n) => n.id === nodeId);
    return node?.ip_address || "";
  };

  const filteredServices = services.filter((s) => {
    if (filterNode && s.node_id !== filterNode) return false;
    if (filterType && s.type !== parseInt(filterType)) return false;
    if (filterStatus && s.status !== filterStatus) return false;
    if (searchText && !s.name.toLowerCase().includes(searchText.toLowerCase())) return false;
    return true;
  });

  const relevantConfigs = configs.filter((c) => !c.type || c.type === "default" || c.type === HoneypotTypeMap[form.type]?.toLowerCase());

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">服务管理</h1>
          <p className="text-muted-foreground mt-1">管理蜜罐服务的配置和运行状态</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm">
          + 新建服务
        </button>
      </div>

      <div className="bg-card rounded-xl shadow-sm p-4">
        <div className="flex flex-wrap gap-3">
          <input
            type="text" placeholder="搜索服务名称..." value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm flex-1 min-w-[200px]" />
          <select value={filterNode} onChange={(e) => setFilterNode(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm">
            <option value="">全部节点</option>
            {nodes.map((n) => <option key={n.id} value={n.id}>{n.hostname} ({n.ip_address})</option>)}
          </select>
          <select value={filterType} onChange={(e) => setFilterType(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm">
            <option value="">全部类型</option>
            {[1, 2, 3, 4, 5, 6, 7].map((t) => <option key={t} value={t}>{HoneypotTypeMap[t]}</option>)}
          </select>
          <select value={filterStatus} onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm">
            <option value="">全部状态</option>
            <option value="running">运行中</option>
            <option value="stopped">已停止</option>
          </select>
          {(filterNode || filterType || filterStatus || searchText) && (
            <button onClick={() => { setFilterNode(""); setFilterType(""); setFilterStatus(""); setSearchText(""); }}
              className="px-3 py-2 text-sm text-muted-foreground hover:text-foreground">清除筛选</button>
          )}
          <span className="flex items-center text-sm text-muted-foreground ml-auto">
            共 {filteredServices.length} 个服务
          </span>
        </div>
      </div>

      {showCreate && (
        <div className="bg-card rounded-xl shadow-sm p-6">
          <h2 className="text-lg font-semibold mb-4">新建蜜罐服务</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">部署节点</label>
              <select value={form.node_id} onChange={(e) => setForm({ ...form, node_id: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg">
                <option value="">选择节点 (可选)</option>
                {nodes.filter((n) => n.status === "online").map((n) => (
                  <option key={n.id} value={n.id}>{n.hostname} ({n.ip_address})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">服务名称</label>
              <input type="text" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg" placeholder="如: mysql-honeypot-1" />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">蜜罐类型</label>
              <select value={form.type} onChange={(e) => handleTypeChange(parseInt(e.target.value))}
                className="w-full px-3 py-2 border rounded-lg">
                {[1, 2, 3, 4, 5, 6, 7].map((t) => (
                  <option key={t} value={t}>{HoneypotTypeMap[t]}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">配置策略</label>
              <select value={selectedConfigId} onChange={(e) => handleConfigChange(e.target.value ? parseInt(e.target.value) : "")}
                className="w-full px-3 py-2 border rounded-lg">
                <option value="">使用默认配置</option>
                {relevantConfigs.map((c) => (
                  <option key={c.id} value={c.id}>{c.name} (v{c.version})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">监听地址</label>
              <input type="text" value={form.host} onChange={(e) => setForm({ ...form, host: e.target.value })}
                className="w-full px-3 py-2 border rounded-lg" />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">端口</label>
              <input type="number" value={form.port} onChange={(e) => setForm({ ...form, port: parseInt(e.target.value) })}
                className="w-full px-3 py-2 border rounded-lg" />
            </div>
          </div>
          <div className="mt-4 flex gap-2">
            <button onClick={createService} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm">创建</button>
            <button onClick={() => { setShowCreate(false); setSelectedConfigId(""); setForm({ ...form, config: "" }); }} className="px-4 py-2 bg-muted text-foreground rounded-lg hover:bg-accent text-sm">取消</button>
          </div>
        </div>
      )}

      {loading ? (
        <div className="flex flex-col items-center justify-center py-16 gap-3">
          <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />
          <span className="text-sm text-muted-foreground">加载中...</span>
        </div>
      ) : filteredServices.length === 0 ? (
        <div className="bg-card rounded-xl shadow-sm p-16 text-center">
          <p className="text-5xl mb-4">🔧</p>
          <h3 className="text-lg font-medium text-foreground">暂无蜜罐服务</h3>
          <p className="text-muted-foreground mt-2">点击右上角创建第一个蜜罐服务</p>
        </div>
      ) : (
        <div className="grid gap-3">
          {filteredServices.map((svc) => (
            <div key={svc.id} className="bg-card rounded-xl shadow-sm p-4 flex items-center justify-between hover:shadow-md transition-shadow">
              <div className="flex items-center gap-4">
                <div className={`w-11 h-11 rounded-xl flex items-center justify-center text-white font-bold text-sm ${
                  svc.status === "running" ? "bg-green-500" : svc.status === "stopped" ? "bg-gray-400" : "bg-yellow-500"
                }`}>
                  {HoneypotTypeMap[svc.type]?.slice(0, 2) || "??"}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-foreground">{svc.name}</span>
                    <span className={`px-2 py-0.5 text-xs rounded-full ${
                      svc.status === "running" ? "bg-green-100 text-green-700" :
                      svc.status === "stopped" ? "bg-gray-100 text-gray-600" :
                      "bg-yellow-100 text-yellow-700"
                    }`}>{svc.status === "running" ? "运行中" : svc.status === "stopped" ? "已停止" : svc.status}</span>
                    <span className="px-2 py-0.5 text-xs bg-blue-50 text-blue-600 rounded">{HoneypotTypeMap[svc.type]}</span>
                  </div>
                  <div className="text-sm text-muted-foreground flex items-center gap-2 mt-0.5 flex-wrap">
                    <span className="font-mono">{svc.host}:{svc.port}</span>
                    <span className="text-muted-foreground/50">|</span>
                    <span>{svc.protocol}</span>
                    {svc.node_id && (
                      <>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-muted-foreground">{getNodeName(svc.node_id)}</span>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-muted-foreground font-mono text-xs">{getNodeIP(svc.node_id)}</span>
                      </>
                    )}
                    {svc.config && (
                      <>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-purple-400 text-xs">已配置策略</span>
                      </>
                    )}
                  </div>
                </div>
              </div>
              <div className="flex gap-1.5">
                <button onClick={() => actionService("start", svc.id)}
                  className="px-3 py-1.5 text-xs bg-green-50 text-green-600 rounded-lg hover:bg-green-100 font-medium">启动</button>
                <button onClick={() => actionService("stop", svc.id)}
                  className="px-3 py-1.5 text-xs bg-yellow-50 text-yellow-600 rounded-lg hover:bg-yellow-100 font-medium">停止</button>
                <button onClick={() => actionService("restart", svc.id)}
                  className="px-3 py-1.5 text-xs bg-blue-50 text-blue-600 rounded-lg hover:bg-blue-100 font-medium">重启</button>
                <button onClick={() => deleteService(svc.id)}
                  className="px-3 py-1.5 text-xs bg-red-50 text-red-600 rounded-lg hover:bg-red-100 font-medium">删除</button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
