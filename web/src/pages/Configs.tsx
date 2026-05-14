import { useEffect, useState } from "react";
import { configApi, serviceApi, nodeApi } from "../api";
import type { ConfigTemplate, HoneypotService, Node } from "../types";
import { useToast } from "../components/Toast";
import { HoneypotTypeMap } from "../types";

export default function Configs() {
  const [configs, setConfigs] = useState<ConfigTemplate[]>([]);
  const [services, setServices] = useState<HoneypotService[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [loading, setLoading] = useState(true);
  const [editId, setEditId] = useState<number | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [filterType, setFilterType] = useState<string>("");
  const [form, setForm] = useState({ name: "", description: "", type: "default", content: "", node_id: "", service_ids: [] as number[] });
  const { toast } = useToast();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [cfgRes, svcRes, nodeRes] = await Promise.all([configApi.list(), serviceApi.list(), nodeApi.list()]);
      if (cfgRes.data.success) setConfigs(cfgRes.data.configs);
      if (svcRes.data.success) setServices(svcRes.data.services);
      if (nodeRes.data.success) setNodes(nodeRes.data.nodes);
    } catch {
      toast("error", "加载配置列表失败");
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    if (!form.name || !form.content) {
      toast("error", "名称和内容不能为空");
      return;
    }
    try {
      await configApi.create(form);
      toast("success", "配置创建成功");
      setShowCreate(false);
      setForm({ name: "", description: "", type: "default", content: "", node_id: "", service_ids: [] });
      loadData();
    } catch {
      toast("error", "创建配置失败");
    }
  };

  const handleUpdate = async (id: number) => {
    if (!form.name || !form.content) {
      toast("error", "名称和内容不能为空");
      return;
    }
    try {
      await configApi.update(id, form);
      toast("success", "配置更新成功");
      setEditId(null);
      setForm({ name: "", description: "", type: "default", content: "", node_id: "", service_ids: [] });
      loadData();
    } catch {
      toast("error", "更新配置失败");
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("确定删除该配置模板?")) return;
    try {
      await configApi.delete(id);
      toast("success", "配置已删除");
      loadData();
    } catch {
      toast("error", "删除配置失败");
    }
  };

  const handlePush = async (id: number) => {
    try {
      await configApi.push(id);
      toast("success", "配置推送已排队");
    } catch {
      toast("error", "推送失败");
    }
  };

  const cancelForm = () => {
    setEditId(null);
    setShowCreate(false);
    setForm({ name: "", description: "", type: "default", content: "", node_id: "", service_ids: [] });
  };

  const startEdit = (cfg: ConfigTemplate) => {
    setEditId(cfg.id);
    setShowCreate(false);
    let parsedIds: number[] = [];
    if (cfg.service_ids) {
      if (typeof cfg.service_ids === "string") {
        try { parsedIds = JSON.parse(cfg.service_ids); } catch { /* ignore */ }
      } else {
        parsedIds = cfg.service_ids;
      }
    }
    setForm({
      name: cfg.name,
      description: cfg.description,
      type: cfg.type,
      content: cfg.content,
      node_id: cfg.node_id || "",
      service_ids: parsedIds,
    });
  };

  const startCreate = () => {
    setShowCreate(true);
    setEditId(null);
    setForm({ name: "", description: "", type: "default", content: "", node_id: "", service_ids: [] });
  };

  const filteredServices = form.node_id ? services.filter((s) => s.node_id === form.node_id) : [];

  const getNodeName = (nid: string) => nodes.find((n) => n.id === nid)?.hostname || nid || "全局";
  const getServiceName = (sid: number) => {
    const s = services.find((svc) => svc.id === sid);
    return s ? `${s.name}(${HoneypotTypeMap[s.type] || s.type}:#${s.id})` : `#${sid}`;
  };

  const parseServiceIds = (cfg: ConfigTemplate): number[] => {
    if (!cfg.service_ids) return [];
    if (typeof cfg.service_ids === "string") {
      try { return JSON.parse(cfg.service_ids); } catch { return []; }
    }
    return cfg.service_ids;
  };

  const getServicesUsingConfig = (cfg: ConfigTemplate): HoneypotService[] => {
    if (!cfg.content) return [];
    return services.filter((s) => s.config === cfg.content);
  };

  const filteredConfigs = configs.filter((c) => {
    if (filterType && c.type !== filterType) return false;
    return true;
  });

  const typeOptions = ["default", "mysql", "redis", "mongodb", "elasticsearch", "weblogic", "nginx", "apache"];

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">配置管理</h1>
          <p className="text-muted-foreground mt-1">管理蜜罐配置模板，绑定到节点的服务实例</p>
        </div>
        <div className="flex gap-2">
          <button onClick={loadData} className="px-4 py-2 bg-muted text-foreground rounded hover:bg-accent text-sm">
            刷新
          </button>
          <button onClick={startCreate} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm">
            + 新建配置
          </button>
        </div>
      </div>

      <div className="bg-card rounded-xl shadow-sm p-4">
        <div className="flex flex-wrap gap-3 items-center">
          <span className="text-sm font-medium text-foreground">筛选类型:</span>
          <select value={filterType} onChange={(e) => setFilterType(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm">
            <option value="">全部类型</option>
            {typeOptions.map((t) => <option key={t} value={t}>{t === "default" ? "默认/通用" : HoneypotTypeMap[Object.keys(HoneypotTypeMap).find(k => HoneypotTypeMap[parseInt(k)] === t) as any] || t}</option>)}
          </select>
          <span className="flex items-center text-sm text-muted-foreground ml-auto">
            共 {filteredConfigs.length} 个配置
          </span>
        </div>
      </div>

      {showCreate && (
        <div className="bg-card rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">新建配置模板</h2>
          <ConfigForm form={form} setForm={setForm} nodes={nodes} filteredServices={filteredServices} onSave={handleCreate} onCancel={cancelForm} saveLabel="创建" />
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-muted-foreground">加载中...</div>
      ) : filteredConfigs.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">暂无配置模板，点击右上角新建</div>
      ) : (
        <div className="space-y-3">
          {filteredConfigs.map((cfg) => {
            const usingSvcs = getServicesUsingConfig(cfg);
            return (
            <div key={cfg.id} className="bg-card rounded-lg shadow">
              {editId === cfg.id ? (
                <div className="p-6">
                  <h2 className="text-lg font-semibold mb-4">编辑配置 - {cfg.name}</h2>
                  <ConfigForm form={form} setForm={setForm} nodes={nodes} filteredServices={filteredServices} onSave={() => handleUpdate(cfg.id)} onCancel={cancelForm} saveLabel="保存" />
                </div>
              ) : (
                <>
                  <div className="p-4 flex items-center justify-between">
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-foreground">{cfg.name}</span>
                        <span className="px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded">{cfg.type}</span>
                        <span className="px-2 py-0.5 text-xs bg-muted text-muted-foreground rounded">v{cfg.version}</span>
                        <span className="px-2 py-0.5 text-xs bg-green-100 text-green-700 rounded">{getNodeName(cfg.node_id)}</span>
                      </div>
                      {cfg.description && <p className="text-sm text-muted-foreground mt-1">{cfg.description}</p>}
                      <p className="text-xs text-muted-foreground mt-1">更新于 {new Date(cfg.updated_at).toLocaleString()}</p>
                      {parseServiceIds(cfg).length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          <span className="text-xs text-muted-foreground">绑定服务:</span>
                          {parseServiceIds(cfg).map((sid) => (
                            <span key={sid} className="px-2 py-0.5 text-xs bg-purple-100 text-purple-700 rounded">
                              {getServiceName(sid)}
                            </span>
                          ))}
                        </div>
                      )}
                      {usingSvcs.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-2">
                          <span className="text-xs text-green-500">使用中:</span>
                          {usingSvcs.map((s) => (
                            <span key={s.id} className="px-2 py-0.5 text-xs bg-green-50 text-green-600 rounded">
                              {s.name}({HoneypotTypeMap[s.type]})
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                    <div className="flex gap-2">
                      <button onClick={() => handlePush(cfg.id)}
                        className="px-3 py-1 text-sm bg-green-50 text-green-600 rounded hover:bg-green-100">推送</button>
                      <button onClick={() => startEdit(cfg)}
                        className="px-3 py-1 text-sm bg-blue-50 text-blue-600 rounded hover:bg-blue-100">编辑</button>
                      <button onClick={() => handleDelete(cfg.id)}
                        className="px-3 py-1 text-sm bg-red-50 text-red-600 rounded hover:bg-red-100">删除</button>
                    </div>
                  </div>
                  <div className="border-t px-4 py-3 bg-muted">
                    <pre className="text-xs text-muted-foreground overflow-x-auto whitespace-pre-wrap font-mono max-h-40 overflow-y-auto">
                      {cfg.content}
                    </pre>
                  </div>
                </>
              )}
            </div>
          );})}
        </div>
      )}
    </div>
  );
}

function ConfigForm({ form, setForm, nodes, filteredServices, onSave, onCancel, saveLabel }: {
  form: { name: string; description: string; type: string; content: string; node_id: string; service_ids: number[] };
  setForm: (f: { name: string; description: string; type: string; content: string; node_id: string; service_ids: number[] }) => void;
  nodes: Node[];
  filteredServices: HoneypotService[];
  onSave: () => void;
  onCancel: () => void;
  saveLabel: string;
}) {
  const toggleService = (id: number) => {
    const ids = form.service_ids.includes(id)
      ? form.service_ids.filter((x) => x !== id)
      : [...form.service_ids, id];
    setForm({ ...form, service_ids: ids });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">配置名称</label>
          <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })}
            className="w-full px-3 py-2 border rounded" placeholder="如: mysql-honeypot-v1" />
        </div>
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">类型</label>
          <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}
            className="w-full px-3 py-2 border rounded">
            <option value="default">default</option>
            <option value="mysql">mysql</option>
            <option value="redis">redis</option>
            <option value="mongodb">mongodb</option>
            <option value="elasticsearch">elasticsearch</option>
            <option value="weblogic">weblogic</option>
            <option value="nginx">nginx</option>
            <option value="apache">apache</option>
          </select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">所属节点</label>
          <select value={form.node_id} onChange={(e) => setForm({ ...form, node_id: e.target.value, service_ids: [] })}
            className="w-full px-3 py-2 border rounded">
            <option value="">全局(所有节点)</option>
            {nodes.map((n) => (
              <option key={n.id} value={n.id}>{n.hostname} ({n.ip_address})</option>
            ))}
          </select>
        </div>
        <div>
          <label className="block text-sm font-medium text-foreground mb-1">绑定服务</label>
          <div className="border rounded px-3 py-2 min-h-[42px] flex flex-wrap gap-1">
            {filteredServices.length === 0 && <span className="text-muted-foreground text-sm">请先选择节点</span>}
            {filteredServices.map((s) => (
              <span
                key={s.id}
                onClick={() => toggleService(s.id)}
                className={`px-2 py-0.5 text-xs rounded cursor-pointer ${form.service_ids.includes(s.id) ? "bg-purple-600 text-white" : "bg-muted text-muted-foreground hover:bg-accent"}`}
              >
                {s.name}({HoneypotTypeMap[s.type] || s.type})
              </span>
            ))}
          </div>
        </div>
      </div>
      <div>
        <label className="block text-sm font-medium text-foreground mb-1">描述</label>
        <input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })}
          className="w-full px-3 py-2 border rounded" placeholder="可选描述" />
      </div>
      <div>
        <label className="block text-sm font-medium text-foreground mb-1">配置内容 (JSON)</label>
        <textarea value={form.content} onChange={(e) => setForm({ ...form, content: e.target.value })}
          className="w-full px-3 py-2 border rounded font-mono text-sm" rows={10}
          placeholder='{"port": 3306, "banner": "5.7.0", ...}' />
      </div>
      <div className="flex gap-2">
        <button onClick={onSave} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm">{saveLabel}</button>
        <button onClick={onCancel} className="px-4 py-2 bg-muted text-foreground rounded hover:bg-accent text-sm">取消</button>
      </div>
    </div>
  );
}
