package database

import (
	"fmt"
	"log"

	"github.com/chenflux/honeywatch/internal/config"
	models "github.com/chenflux/honeywatch/internal/models"
	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func IsPostgres() bool {
	if DB == nil {
		return false
	}
	_, ok := DB.Dialector.(*postgres.Dialector)
	return ok
}

func Init(cfg *config.AppConfig) error {
	var err error
	var dialector gorm.Dialector

	switch cfg.Database.Type {
	case "postgres":
		dsn := cfg.Database.Dsn
		if dsn == "" {
			return fmt.Errorf("postgres DSN is required")
		}
		dialector = postgres.Open(dsn)
	case "sqlite":
		fallthrough
	default:
		dbPath := cfg.Database.Path
		if dbPath == "" {
			dbPath = "data/honeywatch.db"
		}
		dialector = sqlite.Open(dbPath)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Node{},
		&models.HoneypotService{},
		&models.RequestLog{},
		&models.RuleGroup{},
		&models.Rule{},
		&models.ConfigTemplate{},
		&models.AlertChannel{},
		&models.AlertRule{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_created_attack ON request_logs(created_at, is_attack)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_created_type ON request_logs(created_at, honeypot_type)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_ip_created ON request_logs(client_ip, created_at)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_node_created ON request_logs(node_id, created_at)")

	if err := seedData(); err != nil {
		log.Printf("Warning: seed data failed: %v", err)
	}

	return nil
}

func seedData() error {
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		user := models.User{
			Username:     "admin",
			PasswordHash: string(hash),
			Role:         "admin",
		}
		if err := DB.Create(&user).Error; err != nil {
			return fmt.Errorf("create default user: %w", err)
		}
		log.Println("Created default admin user (username: admin, password: admin)")
	}

	DB.Model(&models.RuleGroup{}).Count(&count)
	if count == 0 {
		seedRuleGroups()
	}

	return nil
}

func seedRuleGroups() {
	groups := []struct {
		Name        string
		Protocol    string
		Description string
		Rules       []struct {
			Type    string
			Pattern string
			Action  string
		}
	}{
		{
			Name: "HTTP 基础防护策略", Protocol: "http",
			Description: "HTTP 服务的基础安全防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"path", "/admin/*", "alert"},
				{"path", "/backup/*", "block"},
				{"path", "/.env", "block"},
				{"path", "/robots.txt", "alert"},
			},
		},
		{
			Name: "SQL 注入防护策略", Protocol: "http",
			Description: "针对 SQL 注入攻击的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"query", "' OR 1=1 --", "block"},
				{"query", "UNION SELECT", "block"},
				{"query", "INSERT INTO", "alert"},
				{"query", "; DROP", "block"},
			},
		},
		{
			Name: "XSS 防护策略", Protocol: "http",
			Description: "针对跨站脚本攻击的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"query", "<script>", "block"},
				{"query", "javascript:", "block"},
				{"query", "onload=", "block"},
				{"query", "onerror=", "block"},
				{"query", "<iframe", "alert"},
			},
		},
		{
			Name: "命令注入防护策略", Protocol: "http",
			Description: "针对命令注入攻击的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"query", "; cat /etc/passwd", "block"},
				{"query", "| whoami", "block"},
				{"query", "&& ls -la", "block"},
			},
		},
		{
			Name: "目录遍历防护策略", Protocol: "http",
			Description: "针对目录遍历攻击的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"path", "../", "block"},
				{"path", "..\\", "block"},
				{"path", "/etc/passwd", "block"},
				{"path", "/proc/self/environ", "block"},
			},
		},
		{
			Name: "常见漏洞利用路径防护", Protocol: "http",
			Description: "针对常见漏洞利用路径的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"path", "/_plugin/marvel/", "block"},
				{"path", "/_plugin/head/", "block"},
				{"path", "/_river/", "block"},
				{"path", "/wls-wsat/", "block"},
				{"path", "/console/css/%252e%252e%252f", "block"},
			},
		},
		{
			Name: "MySQL 服务防护策略", Protocol: "mysql",
			Description: "针对 MySQL 服务的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"query", "' OR 1=1 --", "alert"},
				{"query", "UNION SELECT", "alert"},
				{"query", "DROP DATABASE", "alert"},
			},
		},
		{
			Name: "Redis 服务防护策略", Protocol: "redis",
			Description: "针对 Redis 服务的防护策略组",
			Rules: []struct {
				Type    string
				Pattern string
				Action  string
			}{
				{"command", "CONFIG SET", "block"},
				{"command", "FLUSHALL", "block"},
				{"command", "FLUSHDB", "block"},
				{"command", "SLAVEOF", "alert"},
			},
		},
	}

	for _, g := range groups {
		group := models.RuleGroup{
			Name:        g.Name,
			Protocol:    g.Protocol,
			Description: g.Description,
			Enabled:     true,
		}
		if err := DB.Create(&group).Error; err != nil {
			log.Printf("Warning: create rule group %s failed: %v", g.Name, err)
			continue
		}
		for _, r := range g.Rules {
			rule := models.Rule{
				GroupID: group.ID,
				Type:    r.Type,
				Pattern: r.Pattern,
				Action:  r.Action,
				Enabled: true,
			}
			DB.Create(&rule)
		}
	}
	log.Printf("Seeded %d rule groups", len(groups))
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
