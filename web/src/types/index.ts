export interface User {
  id: number;
  username: string;
  role: string;
}

export interface Node {
  id: string;
  agent_id: string;
  hostname: string;
  ip_address: string;
  status: string;
  agent_version: string;
  os_name: string;
  os_version: string;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  last_heartbeat: string;
  created_at: string;
}

export interface HoneypotService {
  id: number;
  node_id: string;
  name: string;
  type: number;
  protocol: string;
  host: string;
  port: number;
  tls: boolean;
  enabled: boolean;
  status: string;
  config: string;
  rule_groups: string;
  created_at: string;
}

export interface RequestLog {
  id: number;
  node_id: string;
  service_id: number;
  client_ip: string;
  method: string;
  path: string;
  user_agent: string;
  honeypot_type: number;
  status_code: number;
  is_attack: boolean;
  attack_type: string;
  attack_detail: string;
  protocol: string;
  created_at: string;
}

export interface RuleGroup {
  id: number;
  name: string;
  protocol: string;
  description: string;
  enabled: boolean;
  rules: Rule[];
}

export interface Rule {
  id: number;
  group_id: number;
  type: string;
  pattern: string;
  action: string;
  enabled: boolean;
}

export interface StatsSummary {
  total_requests: number;
  honeypot_requests: number;
  attack_count: number;
  by_honeypot_type: Record<number, number>;
  by_method: Record<string, number>;
  by_attack_type: Record<string, number>;
  top_ips: { ip: string; count: number }[];
  recent_attacks: RequestLog[];
}

export interface TrendData {
  labels: string[];
  data: number[];
  attacks_data: number[];
}

export interface ConfigTemplate {
  id: number;
  name: string;
  description: string;
  type: string;
  content: string;
  version: number;
  service_ids: number[];
  node_id: string;
  created_at: string;
  updated_at: string;
}

export const HoneypotTypeMap: Record<number, string> = {
  0: "None",
  1: "Elasticsearch",
  2: "WebLogic",
  3: "MySQL",
  4: "Redis",
  5: "MongoDB",
  6: "Nginx",
  7: "Apache",
};
