package internal_models

import "time"

type RuleGroup struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Protocol    string `gorm:"size:20;not null" json:"protocol"`
	Description string `gorm:"size:500" json:"description"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	Rules       []Rule `gorm:"foreignKey:GroupID" json:"rules"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Rule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	GroupID   uint      `gorm:"index;not null" json:"group_id"`
	Type      string    `gorm:"size:30;not null" json:"type"`
	Pattern   string    `gorm:"size:500;not null" json:"pattern"`
	Action    string    `gorm:"size:20;default:alert" json:"action"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
