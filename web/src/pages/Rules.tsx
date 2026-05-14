import { useEffect, useState } from "react";
import { ruleApi } from "../api";
import type { RuleGroup } from "../types";
import { useToast } from "../components/Toast";

export default function Rules() {
  const [groups, setGroups] = useState<RuleGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedGroup, setExpandedGroup] = useState<number | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    try {
      const res = await ruleApi.listGroups();
      if (res.data.success) setGroups(res.data.groups);
    } catch {
      toast("error", "加载规则失败");
    } finally {
      setLoading(false);
    }
  };

  const toggleGroup = (id: number) => {
    setExpandedGroup(expandedGroup === id ? null : id);
  };

  const deleteGroup = async (id: number) => {
    if (!confirm("确定删除该规则组?")) return;
    try {
      await ruleApi.deleteGroup(id);
      toast("success", "规则组已删除");
      loadGroups();
    } catch {
      toast("error", "删除规则组失败");
    }
  };

  const deleteRule = async (groupId: number, ruleId: number) => {
    try {
      await ruleApi.deleteRule(groupId, ruleId);
      toast("success", "规则已删除");
      loadGroups();
    } catch {
      toast("error", "删除规则失败");
    }
  };

  const addRule = async (groupId: number) => {
    const type = prompt("规则类型 (path/query/header/body/user_agent):");
    const pattern = prompt("匹配模式:");
    if (!type || !pattern) return;
    try {
      await ruleApi.createRule(groupId, { type, pattern, action: "alert" });
      toast("success", "规则已添加");
      loadGroups();
    } catch {
      toast("error", "添加规则失败");
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-foreground">规则中心</h1>
          <p className="text-muted-foreground mt-1">管理蜜罐安全检测规则</p>
        </div>
        <button onClick={loadGroups} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm">
          刷新
        </button>
      </div>

      {loading ? (
        <div className="text-center py-12 text-muted-foreground">加载中...</div>
      ) : (
        <div className="space-y-3">
          {groups.map((group) => (
            <div key={group.id} className="bg-card rounded-lg shadow">
              <div
                className="p-4 flex items-center justify-between cursor-pointer hover:bg-accent"
                onClick={() => toggleGroup(group.id)}
              >
                <div className="flex items-center gap-3">
                  <span className="text-lg">{expandedGroup === group.id ? "▼" : "▶"}</span>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-semibold">{group.name}</span>
                      <span className="px-2 py-0.5 text-xs bg-blue-100 text-blue-700 rounded">{group.protocol}</span>
                      <span className={`px-2 py-0.5 text-xs rounded ${
                        group.enabled ? "bg-green-100 text-green-700" : "bg-gray-100 text-gray-500"
                      }`}>{group.enabled ? "启用" : "禁用"}</span>
                    </div>
                    <p className="text-sm text-muted-foreground mt-1">{group.description}</p>
                  </div>
                </div>
                <div className="flex gap-2">
                  <span className="text-sm text-muted-foreground">{group.rules?.length || 0} 条规则</span>
                  <button onClick={(e) => { e.stopPropagation(); deleteGroup(group.id); }}
                    className="text-red-600 text-sm hover:underline">删除</button>
                </div>
              </div>

              {expandedGroup === group.id && (
                <div className="border-t px-4 py-3">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="text-muted-foreground">
                        <th className="text-left py-1 px-2">类型</th>
                        <th className="text-left py-1 px-2">模式</th>
                        <th className="text-left py-1 px-2">动作</th>
                        <th className="text-left py-1 px-2">状态</th>
                        <th className="text-left py-1 px-2">操作</th>
                      </tr>
                    </thead>
                    <tbody>
                      {group.rules?.map((rule) => (
                        <tr key={rule.id} className="border-t hover:bg-accent">
                          <td className="py-2 px-2"><span className="px-2 py-0.5 bg-muted rounded text-xs">{rule.type}</span></td>
                          <td className="py-2 px-2 font-mono text-xs">{rule.pattern}</td>
                          <td className="py-2 px-2">
                            <span className={`px-2 py-0.5 text-xs rounded ${
                              rule.action === "block" ? "bg-red-100 text-red-700" : "bg-yellow-100 text-yellow-700"
                            }`}>{rule.action}</span>
                          </td>
                          <td className="py-2 px-2">
                            <span className={`text-xs ${rule.enabled ? "text-green-600" : "text-muted-foreground"}`}>
                              {rule.enabled ? "启用" : "禁用"}
                            </span>
                          </td>
                          <td className="py-2 px-2">
                            <button onClick={() => deleteRule(group.id, rule.id)} className="text-red-600 text-xs hover:underline">删除</button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                  <button onClick={() => addRule(group.id)}
                    className="mt-2 px-3 py-1 text-sm bg-blue-50 text-blue-600 rounded hover:bg-blue-100">+ 添加规则</button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
