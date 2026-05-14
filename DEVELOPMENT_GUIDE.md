# Pitcher v2 - Go 重写开发指南

## 一、项目概述

将原 Python 单体蜜罐系统重写为 **管理中心(Go) + 前端(React) + Agent(Go)** 的分布式架构。

### 技术选型

| 组件 | 技术栈 |
|------|--------|
| **后端** | Go 1.22+ / gorilla/mux / GORM |
| **前端** | React 18 + TypeScript + Vite + Tailwind CSS |
| **数据库** | SQLite（默认）/ PostgreSQL |
| **Agent** | Go 1.22+ |

---

## 二、架构设计

```
┌─────────────────────────────────────────────────────┐
│                  React Frontend                      │
│         (Vite + TS + Tailwind + React Router)        │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP/REST
┌──────────────────────▼──────────────────────────────┐
│               Go Management Center                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │
│  │REST API  │ │Node Mgr  │ │Config Svc│ │Log Svc │ │
│  └──────────┘ └──────────┘ └──────────┘ └────────┘ │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │Auth Svc  │ │Rule Svc  │ │Stat Svc  │            │
│  └──────────┘ └──────────┘ └──────────┘            │
│  ┌──────────────────────────────────────┐            │
│  │         GORM (SQLite / PostgreSQL)    │            │
│  └──────────────────────────────────────┘            │
└──────────┬──────────────────────────────────────────┘
           │ HTTP API (JSON)
┌──────────▼──────────────────────────────────────────┐
│                Go Agent (per node)                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │Honeypot  │ │Heartbeat │ │Log Collec│            │
│  │ Engine   │ │ Reporter │ │ tor      │            │
│  └──────────┘ └──────────┘ └──────────┘            │
│  ┌──────────────────────────────────────┐            │
│  │  HTTP/TCP/Protocol Honeypot Servers  │            │
│  └──────────────────────────────────────┘            │
└─────────────────────────────────────────────────────┘
```

---

## 三、目录结构

```
Pitcher/
├── _legacy/                    # 旧版Python代码归档
├── cmd/
│   └── server/
│       └── main.go             # 管理中心入口
├── cmd/
│   └── agent/
│       └── main.go             # Agent入口
├── internal/
│   ├── config/                 # 配置加载
│   │   └── config.go
│   ├── models/                 # GORM 数据模型
│   │   ├── node.go
│   │   ├── service.go
│   │   ├── request_log.go
│   │   ├── rule_group.go
│   │   ├── config.go
│   │   └── user.go
│   ├── handler/                # HTTP handler (API层)
│   │   ├── auth.go
│   │   ├── node.go
│   │   ├── service.go
│   │   ├── config.go
│   │   ├── log.go
│   │   ├── rule.go
│   │   └── stats.go
│   ├── service/                # 业务逻辑层
│   │   ├── auth.go
│   │   ├── node.go
│   │   ├── service.go
│   │   ├── config.go
│   │   ├── log.go
│   │   ├── rule.go
│   │   └── stats.go
│   ├── middleware/             # 中间件
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   └── database/               # 数据库初始化
│       └── database.go
├── agent/                      # Agent独立模块
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── honeypot/           # 蜜罐引擎
│   │   │   ├── engine.go       # 引擎接口+工厂
│   │   │   ├── elasticsearch.go
│   │   │   ├── weblogic.go
│   │   │   ├── mysql.go
│   │   │   ├── redis.go
│   │   │   ├── mongodb.go
│   │   │   ├── nginx.go
│   │   │   └── apache.go
│   │   ├── server/             # 协议服务器
│   │   │   ├── http_server.go
│   │   │   ├── tcp_server.go
│   │   │   └── mysql_server.go
│   │   ├── detector/           # 攻击检测
│   │   │   ├── detector.go
│   │   │   ├── path_detector.go
│   │   │   ├── header_detector.go
│   │   │   ├── body_detector.go
│   │   │   └── behavior_detector.go
│   │   ├── reporter/           # 心跳+日志上报
│   │   │   └── reporter.go
│   │   └── rule/               # 规则引擎
│   │       └── rule.go
│   └── go.mod
├── web/                        # React前端
│   ├── src/
│   │   ├── api/                # API客户端
│   │   ├── components/         # 通用组件
│   │   ├── pages/              # 页面
│   │   │   ├── Login.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Nodes.tsx
│   │   │   ├── Services.tsx
│   │   │   ├── Logs.tsx
│   │   │   ├── Rules.tsx
│   │   │   └── Settings.tsx
│   │   ├── hooks/              # 自定义hooks
│   │   ├── types/              # TypeScript类型
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── data/                       # 数据目录 (gitignored)
│   ├── db/
│   ├── logs/
│   └── configs/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 四、数据库模型设计 (GORM)

### 4.1 users 表

```go
type User struct {
    ID           uint   `gorm:"primaryKey"`
    Username     string `gorm:"uniqueIndex;size:50;not null"`
    PasswordHash string `gorm:"size:255;not null"`
    Role         string `gorm:"size:20;default:admin"` // admin, viewer
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### 4.2 nodes 表 (Agent注册的节点)

```go
type Node struct {
    ID            string `gorm:"primaryKey;size:36"` // UUID
    AgentID       string `gorm:"uniqueIndex;size:64;not null"`
    Hostname      string `gorm:"size:100"`
    IPAddress     string `gorm:"size:45;not null"`
    Status        string `gorm:"size:20;default:online"` // online, offline, warning
    AgentVersion  string `gorm:"size:20"`
    OSName        string `gorm:"size:50"`
    OSVersion     string `gorm:"size:50"`
    CPUUsage      float64
    MemoryUsage   float64
    DiskUsage     float64
    LastHeartbeat time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### 4.3 honeypot_services 表

```go
type HoneypotService struct {
    ID          uint   `gorm:"primaryKey"`
    NodeID      string `gorm:"index;size:36"` // 关联Node
    Name        string `gorm:"size:100;not null"`
    Type        int    `gorm:"not null"`       // 1-7 蜜罐类型
    Protocol    string `gorm:"size:20"`        // http, https, mysql, redis, mongodb
    Host        string `gorm:"size:45;default:0.0.0.0"`
    Port        int    `gorm:"not null"`
    TLS         bool   `gorm:"default:false"`
    TLSCertPath string `gorm:"size:255"`
    TLSKeyPath  string `gorm:"size:255"`
    Enabled     bool   `gorm:"default:true"`
    Config      string `gorm:"type:text"` // JSON配置
    Status      string `gorm:"size:20;default:stopped"` // running, stopped, error
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 4.4 request_logs 表 (蜜罐访问日志)

```go
type RequestLog struct {
    ID           uint   `gorm:"primaryKey"`
    NodeID       string `gorm:"index;size:36"`
    ServiceID    uint   `gorm:"index"`
    ClientIP     string `gorm:"index;size:45;not null"`
    Method       string `gorm:"size:10"`
    Path         string `gorm:"size:500"`
    UserAgent    string `gorm:"size:500"`
    HoneypotType int    `gorm:"index"`
    StatusCode   int
    RequestBody  string `gorm:"type:text"`
    IsAttack     bool   `gorm:"index;default:false"`
    AttackType   string `gorm:"size:50"`
    AttackDetail string `gorm:"type:text"` // JSON
    Protocol     string `gorm:"size:10"`
    CreatedAt    time.Time `gorm:"index"`
}
```

### 4.5 rule_groups 表

```go
type RuleGroup struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"size:100;not null"`
    Protocol    string `gorm:"size:20;not null"` // http, ssh, mysql, redis
    Description string `gorm:"size:500"`
    Enabled     bool   `gorm:"default:true"`
    Rules       []Rule `gorm:"foreignKey:GroupID"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Rule struct {
    ID        uint   `gorm:"primaryKey"`
    GroupID   uint   `gorm:"index;not null"`
    Type      string `gorm:"size:30;not null"` // path, query, header, body, user_agent, method, ip, command, auth_failure
    Pattern   string `gorm:"size:500;not null"`
    Action    string `gorm:"size:20;default:alert"` // alert, block
    Enabled   bool   `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 4.6 configs 表 (节点配置模板)

```go
type ConfigTemplate struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"size:100;not null"`
    Description string `gorm:"size:500"`
    Type        string `gorm:"size:30;default:default"`
    Content     string `gorm:"type:text;not null"` // JSON
    Version     int    `gorm:"default:1"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

---

## 五、API 设计

### 5.1 认证 API
- `POST   /api/auth/login`         - 登录，返回JWT
- `POST   /api/auth/logout`        - 登出
- `GET    /api/auth/profile`       - 获取当前用户信息
- `PUT    /api/auth/password`      - 修改密码

### 5.2 节点 API
- `GET    /api/nodes`              - 节点列表
- `GET    /api/nodes/{id}`         - 节点详情
- `POST   /api/nodes/register`    - Agent注册
- `POST   /api/nodes/{id}/heartbeat` - 心跳上报
- `PUT    /api/nodes/{id}/status`  - 更新状态
- `DELETE /api/nodes/{id}`         - 删除节点

### 5.3 服务 API
- `GET    /api/services`           - 服务列表
- `GET    /api/services/{id}`      - 服务详情
- `POST   /api/services`           - 创建服务
- `PUT    /api/services/{id}`      - 更新服务
- `DELETE /api/services/{id}`      - 删除服务
- `POST   /api/services/{id}/start`  - 启动
- `POST   /api/services/{id}/stop`   - 停止
- `POST   /api/services/{id}/restart`- 重启

### 5.4 配置 API
- `GET    /api/configs`            - 配置模板列表
- `GET    /api/configs/{id}`       - 配置详情
- `POST   /api/configs`            - 创建配置
- `PUT    /api/configs/{id}`       - 更新配置
- `DELETE /api/configs/{id}`       - 删除配置
- `POST   /api/configs/{id}/push`  - 推送到节点

### 5.5 日志 API
- `GET    /api/logs`               - 查询日志(支持分页/筛选)
- `GET    /api/logs/stats`         - 日志统计
- `GET    /api/logs/export`        - 导出日志

### 5.6 统计 API
- `GET    /api/stats/summary`      - 统计摘要
- `GET    /api/stats/trend`        - 趋势数据(1h/24h/7d/30d)
- `GET    /api/stats/top-ips`      - TOP攻击IP
- `GET    /api/stats/attack-types` - 攻击类型分布

### 5.7 规则 API
- `GET    /api/rules/groups`       - 规则组列表
- `GET    /api/rules/groups/{id}`  - 规则组详情
- `POST   /api/rules/groups`       - 创建规则组
- `PUT    /api/rules/groups/{id}`  - 更新规则组
- `DELETE /api/rules/groups/{id}`  - 删除规则组
- `POST   /api/rules/groups/{id}/rules`       - 添加规则
- `PUT    /api/rules/groups/{id}/rules/{rid}`  - 更新规则
- `DELETE /api/rules/groups/{id}/rules/{rid}`  - 删除规则

---

## 六、蜜罐引擎设计 (Agent端)

### 6.1 接口定义

```go
// agent/internal/honeypot/engine.go
type Honeypot interface {
    Name() string
    Type() int
    HandleHTTP(w http.ResponseWriter, r *http.Request) (bool, *AttackInfo)
}

type AttackInfo struct {
    Type    string            `json:"type"`
    Detail  map[string]string `json:"detail"`
    Source  string            `json:"source"` // path, header, body, behavior
}

// AttackDetector 攻击检测接口
type AttackDetector interface {
    Detect(req *http.Request, body []byte) *AttackInfo
}
```

### 6.2 蜜罐类型映射

| Type | 名称 | 协议 | 实现方式 |
|------|------|------|----------|
| 1 | Elasticsearch | HTTP | HTTP Handler, 模拟ES REST API |
| 2 | WebLogic | HTTP | HTTP Handler, 模拟管理控制台 |
| 3 | MySQL | TCP | 原生MySQL协议 (Handshake + Error) |
| 4 | Redis | TCP | Redis RESP协议 |
| 5 | MongoDB | TCP | MongoDB Wire Protocol |
| 6 | Nginx | HTTP | HTTP Handler, 模拟默认页 |
| 7 | Apache | HTTP | HTTP Handler, 模拟默认页 |

### 6.3 攻击检测器 (从旧代码迁移)

从 `old_src/core/honeypot.py` 提取的检测模式:

**Elasticsearch 攻击模式:**
- 目录遍历: `../`, `/etc/passwd`, `/proc/self/environ`
- 敏感文件: `/.git/`, `/.env`, `/robots.txt`
- 漏洞利用: `/_plugin/marvel/`, `/_plugin/head/`, `/_river/`
- User-Agent扫描器: Nikto, Nessus, sqlmap, Burp 等
- SQL注入 (请求体): UNION SELECT, OR 1=1
- 命令注入 (请求体): `;`, `|`, `&&`, 反引号
- 频率限制: 每分钟60次请求

**WebLogic 攻击模式:**
- 路径遍历: `/console/css/%252e%252e%252f`
- 漏洞利用: `/_async/AsyncResponseService`, `/wls-wsat/CoordinatorPortType`
- 敏感文件: `/etc/passwd`, `/.git/`

**通用规则 (从 config.json 迁移):**
- 10组预置规则: HTTP基础/高级防护、SQL注入、XSS、命令注入、目录遍历、漏洞利用路径、SSH暴力破解、MySQL防护、Redis防护

---

## 七、前端页面设计

### 7.1 页面路由

| 路由 | 页面 | 功能 |
|------|------|------|
| `/login` | Login | 登录页 |
| `/` | Dashboard | 仪表盘: 服务概览卡片、趋势图表、最近活动 |
| `/nodes` | Nodes | 节点管理: 在线节点列表、状态监控 |
| `/services` | Services | 服务管理: 蜜罐服务CRUD、启停控制 |
| `/logs` | Logs | 日志查看: 筛选、搜索、导出 |
| `/rules` | Rules | 规则中心: 规则组/规则管理 |
| `/settings` | Settings | 系统设置: 用户管理、主题、密码修改 |

### 7.2 通用组件
- `Layout` - 侧边栏+顶栏布局
- `StatusBadge` - 状态标签 (online/offline/running/stopped)
- `TrendChart` - 趋势图表 (基于recharts/chart.js)
- `DataTable` - 通用数据表格 (分页/排序/筛选)
- `ConfirmDialog` - 确认弹窗
- `ApiProvider` - API请求封装 + JWT管理

---

## 八、开发优先级 (Phase 1 - MVP)

### 第1步: Go后端骨架
1. go mod init, 安装 gorilla/mux, gorm, golang-jwt
2. 数据库模型 + 自动迁移
3. 配置加载 (YAML/JSON)
4. JWT认证中间件
5. 基础API路由

### 第2步: 核心API实现
1. Auth API (登录/登出/密码修改)
2. Node API (注册/心跳/列表)
3. Service API (CRUD/启停)
4. Log API (记录/查询/统计)
5. Rule API (规则组CRUD)
6. Stats API (摘要/趋势)

### 第3步: Agent核心
1. Agent注册 + 心跳
2. HTTP蜜罐服务器 (ES, WebLogic, Nginx, Apache)
3. TCP蜜罐服务器 (MySQL, Redis, MongoDB)
4. 攻击检测引擎
5. 日志上报

### 第4步: React前端
1. Vite项目初始化
2. 路由 + 布局组件
3. 登录页
4. 仪表盘 (统计卡片 + 趋势图)
5. 节点管理页
6. 服务管理页
7. 日志查看页
8. 规则中心页
9. 设置页

---

## 九、旧代码功能映射

| 旧文件 | 迁移到 | 说明 |
|--------|--------|------|
| `old_src/core/honeypot.py` HoneypotBase | `agent/internal/honeypot/engine.go` | 基类->接口 |
| `old_src/core/honeypot.py` ElasticsearchHoneypot | `agent/internal/honeypot/elasticsearch.go` | ES蜜罐 |
| `old_src/core/honeypot.py` WeblogicHoneypot | `agent/internal/honeypot/weblogic.go` | WebLogic蜜罐 |
| `old_src/core/honeypot.py` MySQLHoneypot + MySQLServer | `agent/internal/honeypot/mysql.go` | MySQL蜜罐+TCP协议 |
| `old_src/core/honeypot.py` RedisHoneypot + RedisServer | `agent/internal/honeypot/redis.go` | Redis蜜罐+TCP协议 |
| `old_src/core/honeypot.py` MongoDBHoneypot + MongoDBServer | `agent/internal/honeypot/mongodb.go` | MongoDB蜜罐+TCP协议 |
| `old_src/core/honeypot.py` NginxHoneypot | `agent/internal/honeypot/nginx.go` | Nginx蜜罐 |
| `old_src/core/honeypot.py` ApacheHoneypot | `agent/internal/honeypot/apache.go` | Apache蜜罐 |
| `old_src/core/honeypot.py` 攻击检测方法 | `agent/internal/detector/` | 攻击检测器 |
| `old_src/core/honeypot.py` 延迟/错误注入/指纹伪装 | `agent/internal/honeypot/engine.go` | 引擎通用能力 |
| `old_src/core/server.py` ServerFactory | `agent/internal/server/` | 服务器工厂 |
| `old_src/core/handler.py` 规则匹配 | `agent/internal/rule/rule.go` | 规则引擎 |
| `old_src/core/stats.py` StatsCollector | `internal/service/stats.go` | 统计服务 |
| `old_src/core/logger.py` Logger | `internal/service/log.go` + `agent/internal/reporter/` | 日志服务 |
| `old_src/core/service_manager.py` | `internal/service/service.go` | 服务管理 |
| `old_src/web/app.py` Flask路由 | `internal/handler/` | API handlers |
| `old_src/web/templates/*.html` | `web/src/pages/` | React页面 |
| `data/config.json` 规则数据 | 数据库 seed | 规则初始数据 |
