# Pitcher

蜜罐管理系统，支持实时攻击监控与多平台 Agent。

## 特性

- 实时攻击监控（WebSocket）
- 跨平台 Agent（Windows / Linux / macOS）
- IP 地理位置归属
- 告警通知（邮件 / Webhook / 钉钉）
- 深色主题 UI，8 款配色

## 快速开始

### 构建

```bash
.\build.bat
```

### 运行服务

```bash
dist\pitcher-server-0.2.1-windows-amd64.exe
```

管理后台：http://localhost:8080 （账号 admin，密码 admin）

## 部署 Agent

### Windows

```powershell
Invoke-WebRequest -Uri "http://localhost:8080/api/downloads/agent/pitcher-agent-0.2.1-windows-amd64.exe" -OutFile "pitcher-agent.exe"
set PITCHER_AGENT_TOKEN=<your-token>
.\pitcher-agent.exe
```

### Linux

```bash
curl -L "http://localhost:8080/api/downloads/agent/pitcher-agent-0.2.1-linux-amd64" -o pitcher-agent
chmod +x pitcher-agent
PITCHER_AGENT_TOKEN=<your-token> ./pitcher-agent
```

## 配置

编辑 `data/config.json`：

```json
{
  "server": { "port": 8080, "data_port": 8090 },
  "database": { "type": "sqlite", "path": "data/pitcher.db" },
  "jwt": { "secret": "change-me", "expire_hours": 24 },
  "agent": { "token": "change-me", "heartbeat_interval": 30, "timeout": 60 }
}
```

## 技术栈

- **后端**：Go、gorilla/mux、GORM
- **前端**：React 18、TypeScript、Vite、Tailwind CSS
- **数据库**：SQLite（默认）、PostgreSQL

## 开源协议

MIT