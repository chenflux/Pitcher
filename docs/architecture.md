# 系统架构

## 整体架构

Pitcher 采用双平面架构：

```
┌─────────────────────────────────────────────────────────┐
│                     管理平面 (8080)                      │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌──────────┐  │
│  │ Web UI  │  │ REST API│  │ WebSocket│ │ 静态资源  │  │
│  └─────────┘  └─────────┘  └─────────┘  └──────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                     数据平面 (8090)                      │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────┐  │
│  │  Node 注册   │  │ 心跳保活     │  │  攻击日志    │  │
│  └──────────────┘  └─────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                           ↕
┌────────────┐  ┌────────────┐  ┌────────────┐
│  Agent     │  │  Agent     │  │  Agent     │
│  (Linux)   │  │ (Windows)  │  │  (macOS)   │
└────────────┘  └────────────┘  └────────────┘
```

## 核心组件

### Server（Go）

- **管理平面**：提供 Web UI、REST API、WebSocket 实时事件
- **数据平面**：Agent 注册、心跳、攻击日志接收
- **数据库**：GORM + SQLite/PostgreSQL
- **实时通知**：WebSocket 推送攻击事件、节点状态变更

### Agent（Go）

- **蜜罐服务**：HTTP、SSH、FTP、MySQL、Redis、MongoDB、Nginx、Apache 等
- **资源监控**：CPU、内存、磁盘使用率（gopsutil）
- **心跳上报**：每 30 秒上报节点状态
- **日志上报**：攻击事件实时推送到服务端

### Frontend（React + TypeScript）

- **状态管理**：React Context（Theme、WebSocket）
- **UI 框架**：Tailwind CSS，支持 8 款主题
- **实时交互**：WebSocket 接收攻击通知
- **页面**：Dashboard、节点、日志、服务、规则、配置、设置

## 数据流

### 攻击检测流程

```
攻击者 → Agent（蜜罐服务）→ 记录攻击日志
                            ↓
                      心跳上报 + 日志上报
                            ↓
                      Server（数据平面）
                            ↓
                      WebSocket 推送
                            ↓
                      Web UI 实时显示
```

### 节点注册流程

```
Agent 启动 → 读取 agent.id（本地存储）
           ↓（无 id 或 404）
        注册接口 → 分配 node_id
                   ↓
              保存 agent.id
                   ↓
            进入心跳循环
```

## API 设计

### 管理平面（8080）

| 路径 | 方法 | 说明 |
|------|------|------|
| `/api/auth/login` | POST | 管理员登录 |
| `/api/auth/profile` | GET | 个人信息 |
| `/api/auth/password` | PUT | 修改密码 |
| `/api/nodes` | GET | 节点列表 |
| `/api/nodes/{id}` | GET/DELETE | 节点详情/删除 |
| `/api/logs` | GET | 攻击日志查询 |
| `/api/logs/stats` | GET | 攻击统计 |
| `/api/logs/trend` | GET | 攻击趋势 |
| `/api/services` | GET/POST | 服务列表/创建 |
| `/api/services/{id}` | GET/PUT/DELETE | 服务管理 |
| `/api/configs` | GET/POST | 配置列表/创建 |
| `/api/rules/groups` | GET/POST | 规则组列表/创建 |
| `/api/settings/agent-token` | GET/PUT | Agent Token 管理 |
| `/api/downloads/agent/{filename}` | GET | 下载 Agent 二进制 |
| `/ws/events` | WS | 实时事件流 |

### 数据平面（8090）

| 路径 | 方法 | 说明 |
|------|------|------|
| `/data/nodes/register` | POST | Agent 注册 |
| `/data/nodes/{id}/heartbeat` | POST | 心跳上报 |
| `/data/logs` | POST | 攻击日志 |
| `/data/logs/batch` | POST | 批量日志 |
| `/data/services` | GET | 获取服务配置 |
| `/data/rules` | GET | 获取规则 |
| `/data/configs` | GET | 获取配置 |
| `/data/agents/download/{filename}` | GET | 下载 Agent 更新 |

## 数据模型

### 节点（Node）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| agent_id | string | Agent 唯一标识 |
| hostname | string | 主机名 |
| ip_address | string | IP 地址 |
| status | string | online/offline |
| last_seen | datetime | 最后活跃时间 |
| cpu_usage | float | CPU 使用率 |
| memory_usage | float | 内存使用率 |
| disk_usage | float | 磁盘使用率 |

### 攻击日志（RequestLog）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| node_id | uint | 节点 ID |
| client_ip | string | 攻击者 IP |
| method | string | HTTP 方法 |
| path | string | 请求路径 |
| protocol | string | 协议 |
| user_agent | string | User-Agent |
| geo_country | string | 国家 |
| geo_city | string | 城市 |
| geo_isp | string | ISP |
| created_at | datetime | 时间 |

### 告警渠道（AlertChannel）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 名称 |
| type | string | email/webhook/dingtalk |
| config | json | 配置 |
| enabled | bool | 启用状态 |
| throttle_min | int | 限速（分钟） |

### 告警规则（AlertRule）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 名称 |
| condition | string | 条件类型 |
| threshold | int | 阈值 |
| channel_id | uint | 渠道 ID |
| enabled | bool | 启用状态 |

## 配置说明

### 服务配置（Service）

```json
{
  "name": "http-80",
  "type": "http",
  "port": 80,
  "protocol": "tcp",
  "config": {
    "banner": "SSH-2.0-OpenSSH_8.0",
    "fake_content": "/etc/issue.net"
  }
}
```

### 蜜罐类型

- **HTTP**：低交互 HTTP 蜜罐，可配置 banner 和响应内容
- **SSH**：SSH 弱口令蜜罐，记录登录尝试
- **FTP**：FTP 匿名访问蜜罐
- **MySQL**：MySQL 认证蜜罐
- **Redis**：Redis 未授权访问蜜罐
- **MongoDB**：MongoDB 认证蜜罐
- **Nginx**：HTTP 服务器蜜罐
- **Apache**：HTTP 服务器蜜罐
- **Elasticsearch**：REST API 蜜罐
- **WebLogic**：WebLogic 探测蜜罐

## 部署建议

### 生产环境

1. 更改默认 JWT Secret 和 Agent Token
2. 配置 PostgreSQL 数据库
3. 使用反向代理（HTTPS）
4. 配置防火墙规则
5. 启用日志告警通知

### Agent 部署

```bash
# Linux 一键部署
curl -L "http://server:8080/api/downloads/agent/pitcher-agent-VERSION-linux-amd64" -o /usr/local/bin/pitcher-agent
chmod +x /usr/local/bin/pitcher-agent
nohup PITCHER_AGENT_TOKEN=your-token /usr/local/bin/pitcher-agent &
```