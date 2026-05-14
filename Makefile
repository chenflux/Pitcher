.PHONY: server agent frontend build dev clean

build: server agent frontend

server:
	go build -o bin/server.exe ./cmd/server/

agent:
	cd agent && go build -o ../bin/agent.exe ./cmd/

frontend:
	cd web && npm run build

dev-server:
	go run ./cmd/server/

dev-frontend:
	cd web && npm run dev

clean:
	rm -rf bin/ web/dist/

init-db:
	@echo "Database will be auto-created on first run"
