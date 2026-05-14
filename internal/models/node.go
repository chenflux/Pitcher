package internal_models

import "time"

type Node struct {
	ID            string    `gorm:"primaryKey;size:36" json:"id"`
	AgentID       string    `gorm:"uniqueIndex;size:64;not null" json:"agent_id"`
	Hostname      string    `gorm:"size:100" json:"hostname"`
	IPAddress     string    `gorm:"size:45;not null" json:"ip_address"`
	Status        string    `gorm:"size:20;default:online" json:"status"`
	AgentVersion  string    `gorm:"size:20" json:"agent_version"`
	OSName        string    `gorm:"size:50" json:"os_name"`
	OSVersion     string    `gorm:"size:50" json:"os_version"`
	OSArch        string    `gorm:"size:20" json:"os_arch"`
	CPUCores      int       `json:"cpu_cores"`
	MemoryGB      float64   `json:"memory_gb"`
	DiskGB        float64   `json:"disk_gb"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   float64   `json:"memory_usage"`
	DiskUsage     float64   `json:"disk_usage"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
