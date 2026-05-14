package internal_models

import "time"

type HoneypotService struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	NodeID      string    `gorm:"index;size:36" json:"node_id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Type        int       `gorm:"not null" json:"type"`
	Protocol    string    `gorm:"size:20" json:"protocol"`
	Host        string    `gorm:"size:45;default:0.0.0.0" json:"host"`
	Port        int       `gorm:"not null" json:"port"`
	TLS         bool      `gorm:"default:false" json:"tls"`
	TLSCertPath string    `gorm:"size:255" json:"tls_cert_path"`
	TLSKeyPath  string    `gorm:"size:255" json:"tls_key_path"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	Config      string    `gorm:"type:text" json:"config"`
	Status      string    `gorm:"size:20;default:stopped" json:"status"`
	RuleGroups  string    `gorm:"type:text" json:"rule_groups"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var HoneypotTypeMap = map[int]string{
	0: "None",
	1: "Elasticsearch",
	2: "WebLogic",
	3: "MySQL",
	4: "Redis",
	5: "MongoDB",
	6: "Nginx",
	7: "Apache",
}

var HoneypotProtocolMap = map[int]string{
	1: "http",
	2: "http",
	3: "tcp",
	4: "tcp",
	5: "tcp",
	6: "http",
	7: "http",
}
