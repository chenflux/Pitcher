package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type ServerConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DataPort int    `json:"data_port"`
}

type DatabaseConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
	Dsn  string `json:"dsn"`
}

type JWTConfig struct {
	Secret     string `json:"secret"`
	ExpireHours int   `json:"expire_hours"`
}

type AgentConfig struct {
	HeartbeatInterval int    `json:"heartbeat_interval"`
	Timeout           int    `json:"timeout"`
	Token             string `json:"token"`
}

type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Agent    AgentConfig    `json:"agent"`
}

var (
	GlobalConfig *AppConfig
	once         sync.Once
	mu           sync.RWMutex
)

func Load(configPath string) (*AppConfig, error) {
	var cfg *AppConfig
	var loadErr error

	once.Do(func() {
		cfg = &AppConfig{
			Server: ServerConfig{
				Host:     "0.0.0.0",
				Port:     8080,
				DataPort: 8090,
			},
			Database: DatabaseConfig{
				Type: "sqlite",
				Path: "data/honeywatch.db",
			},
			JWT: JWTConfig{
				Secret:      "honeywatch-secret-change-me",
				ExpireHours: 24,
			},
			Agent: AgentConfig{
				HeartbeatInterval: 30,
				Timeout:           60,
				Token:             "honeywatch-agent-token-change-me",
			},
		}

		if configPath == "" {
			configPath = "data/config.json"
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			if !os.IsNotExist(err) {
				loadErr = fmt.Errorf("read config: %w", err)
				return
			}
		} else {
			if err := json.Unmarshal(data, cfg); err != nil {
				loadErr = fmt.Errorf("parse config: %w", err)
				return
			}
		}

		if envSecret := os.Getenv("HONEYWATCH_JWT_SECRET"); envSecret != "" {
			cfg.JWT.Secret = envSecret
		}
		if envDBType := os.Getenv("HONEYWATCH_DB_TYPE"); envDBType != "" {
			cfg.Database.Type = envDBType
		}
		if envDBDsn := os.Getenv("HONEYWATCH_DB_DSN"); envDBDsn != "" {
			cfg.Database.Dsn = envDBDsn
		}

		dir := filepath.Dir(cfg.Database.Path)
		if dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}

		GlobalConfig = cfg
	})

	if loadErr != nil {
		return nil, loadErr
	}
	return GlobalConfig, nil
}

func (c *AppConfig) Save(configPath string) error {
	if configPath == "" {
		configPath = "data/config.json"
	}
	dir := filepath.Dir(configPath)
	if dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(configPath, data, 0600)
}

func SetAgentToken(token string) error {
	mu.Lock()
	defer mu.Unlock()
	GlobalConfig.Agent.Token = token
	return GlobalConfig.Save("")
}

func GetAgentToken() string {
	mu.RLock()
	defer mu.RUnlock()
	return GlobalConfig.Agent.Token
}
