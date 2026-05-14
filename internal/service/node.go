package service

import (
	"log"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
	"github.com/google/uuid"
)

type NodeService struct{}

func NewNodeService() *NodeService {
	return &NodeService{}
}

type RegisterNodeRequest struct {
	AgentID      string  `json:"agent_id"`
	Hostname     string  `json:"hostname"`
	IPAddress    string  `json:"ip_address"`
	AgentVersion string  `json:"agent_version"`
	OSName       string  `json:"os_name"`
	OSVersion    string  `json:"os_version"`
	OSArch       string  `json:"os_arch"`
	CPUCores     int     `json:"cpu_cores"`
	MemoryGB     float64 `json:"memory_gb"`
	DiskGB       float64 `json:"disk_gb"`
}

type HeartbeatRequest struct {
	Status          string  `json:"status"`
	CPUUsage        float64 `json:"cpu_usage"`
	MemoryUsage     float64 `json:"memory_usage"`
	DiskUsage       float64 `json:"disk_usage"`
	RunningServices []uint  `json:"running_services"`
}

func (s *NodeService) List() ([]models.Node, error) {
	var nodes []models.Node
	err := database.DB.Order("updated_at DESC").Find(&nodes).Error
	if err != nil {
		return nodes, err
	}

	now := time.Now()
	offlineThreshold := 90 * time.Second
	for i := range nodes {
		if now.Sub(nodes[i].LastHeartbeat) > offlineThreshold {
			nodes[i].Status = "offline"
		} else {
			nodes[i].Status = "online"
		}
	}
	return nodes, err
}

func (s *NodeService) GetByID(id string) (*models.Node, error) {
	var node models.Node
	if err := database.DB.Where("id = ?", id).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *NodeService) GetByAgentID(agentID string) (*models.Node, error) {
	var node models.Node
	if err := database.DB.Where("agent_id = ?", agentID).First(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *NodeService) Register(req RegisterNodeRequest) (*models.Node, error) {
	var node models.Node
	result := database.DB.Where("agent_id = ?", req.AgentID).First(&node)

	if result.RowsAffected > 0 {
		now := time.Now()
		node.Hostname = req.Hostname
		node.IPAddress = req.IPAddress
		node.AgentVersion = req.AgentVersion
		node.OSName = req.OSName
		node.OSVersion = req.OSVersion
		node.OSArch = req.OSArch
		node.CPUCores = req.CPUCores
		node.MemoryGB = req.MemoryGB
		node.DiskGB = req.DiskGB
		node.Status = "online"
		node.LastHeartbeat = now
		node.UpdatedAt = now
		database.DB.Save(&node)
		log.Printf("[Node] Re-registered agent_id=%s node_id=%s", req.AgentID, node.ID)
		return &node, nil
	}

	now := time.Now()
	node = models.Node{
		ID:            uuid.New().String(),
		AgentID:       req.AgentID,
		Hostname:      req.Hostname,
		IPAddress:     req.IPAddress,
		AgentVersion:  req.AgentVersion,
		OSName:        req.OSName,
		OSVersion:     req.OSVersion,
		OSArch:        req.OSArch,
		CPUCores:      req.CPUCores,
		MemoryGB:      req.MemoryGB,
		DiskGB:        req.DiskGB,
		Status:        "online",
		LastHeartbeat: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := database.DB.Create(&node).Error; err != nil {
		return nil, err
	}
	log.Printf("[Node] Registered new agent_id=%s node_id=%s host=%s ip=%s", req.AgentID, node.ID, req.Hostname, req.IPAddress)
	return &node, nil
}

func (s *NodeService) Heartbeat(nodeID string, req HeartbeatRequest) error {
	node, err := s.GetByID(nodeID)
	if err != nil {
		node, err = s.GetByAgentID(nodeID)
		if err != nil {
			return err
		}
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         req.Status,
		"cpu_usage":      req.CPUUsage,
		"memory_usage":   req.MemoryUsage,
		"disk_usage":     req.DiskUsage,
		"last_heartbeat": now,
		"updated_at":     now,
	}
	if req.Status == "" {
		updates["status"] = "online"
	}

	log.Printf("[Node] Heartbeat node_id=%s status=%s cpu=%.1f mem=%.1f", node.ID, updates["status"], req.CPUUsage, req.MemoryUsage)
	return database.DB.Model(node).Updates(updates).Error
}

func (s *NodeService) UpdateStatus(nodeID, status string) error {
	return database.DB.Model(&models.Node{}).Where("id = ?", nodeID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (s *NodeService) Delete(nodeID string) error {
	return database.DB.Where("id = ?", nodeID).Delete(&models.Node{}).Error
}
