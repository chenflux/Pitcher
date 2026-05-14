package internal_models

import "time"

type RequestLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	NodeID       string    `gorm:"index;size:36" json:"node_id"`
	ServiceID    uint      `gorm:"index" json:"service_id"`
	ClientIP     string    `gorm:"size:45;not null" json:"client_ip"`
	Method       string    `gorm:"size:10" json:"method"`
	Path         string    `gorm:"size:500" json:"path"`
	UserAgent    string    `gorm:"size:500" json:"user_agent"`
	HoneypotType int       `gorm:"index" json:"honeypot_type"`
	StatusCode   int       `json:"status_code"`
	RequestBody  string    `gorm:"type:text" json:"request_body"`
	IsAttack     bool      `gorm:"index;default:false" json:"is_attack"`
	AttackType   string    `gorm:"size:50" json:"attack_type"`
	AttackDetail string    `gorm:"type:text" json:"attack_detail"`
	Protocol     string    `gorm:"size:10" json:"protocol"`
	GeoCountry   string    `gorm:"size:50" json:"geo_country"`
	GeoCity      string    `gorm:"size:100" json:"geo_city"`
	GeoISP       string    `gorm:"size:200" json:"geo_isp"`
	CreatedAt    time.Time `gorm:"index" json:"created_at"`
}

func (RequestLog) TableName() string { return "request_logs" }

type RequestLogCompositeIdx struct{} // marker for migration
