package internal_models

import "time"

type AlertChannel struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Type       string    `gorm:"size:20;not null" json:"type"`
	Enabled    bool      `gorm:"default:true" json:"enabled"`
	Config     string    `gorm:"type:text" json:"config"`
	ThrottleMin int      `gorm:"default:5" json:"throttle_min"`
	LastSent   time.Time `json:"last_sent"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AlertChannel) TableName() string { return "alert_channels" }

type AlertRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Condition   string    `gorm:"size:50" json:"condition"`
	Threshold   int       `json:"threshold"`
	WindowMin   int       `json:"window_min"`
	ChannelID   uint      `gorm:"index" json:"channel_id"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (AlertRule) TableName() string { return "alert_rules" }

const (
	AlertConditionAttackCount = "attack_count"
	AlertConditionAttackType  = "attack_type"
	AlertConditionSourceIP    = "source_ip"
	AlertConditionNewNode     = "new_node"

	AlertChannelEmail   = "email"
	AlertChannelWebhook = "webhook"
	AlertChannelDingTalk = "dingtalk"
)