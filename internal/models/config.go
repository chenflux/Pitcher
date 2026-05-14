package internal_models

import "time"

type ConfigTemplate struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Type        string    `gorm:"size:30;default:default" json:"type"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Version     int       `gorm:"default:1" json:"version"`
	ServiceIDs  string    `gorm:"type:text" json:"service_ids"`
	NodeID      string    `gorm:"index;size:36" json:"node_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
