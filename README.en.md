# Pitcher

Honeypot management system with real-time attack monitoring and multi-platform Agent support.

## Features

- Real-time attack monitoring (WebSocket)
- Cross-platform Agent (Windows / Linux / macOS)
- IP geolocation enrichment
- Alert notifications (Email / Webhook / DingTalk)
- Dark theme UI with 8 color schemes

## Quick Start

### Build

```bash
.\build.bat
```

### Run Server

```bash
dist\pitcher-server-0.2.1-windows-amd64.exe
```

Management UI: http://localhost:8080 (admin / admin)

## Deploy Agent

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

## Configuration

Edit `data/config.json`:

```json
{
  "server": { "port": 8080, "data_port": 8090 },
  "database": { "type": "sqlite", "path": "data/pitcher.db" },
  "jwt": { "secret": "change-me", "expire_hours": 24 },
  "agent": { "token": "change-me", "heartbeat_interval": 30, "timeout": 60 }
}
```

## Tech Stack

- **Backend**: Go, gorilla/mux, GORM
- **Frontend**: React 18, TypeScript, Vite, Tailwind CSS
- **Database**: SQLite (default), PostgreSQL

## License

MIT