package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
)

type HoneypotServiceService struct{}

func NewHoneypotServiceService() *HoneypotServiceService {
	return &HoneypotServiceService{}
}

type CreateServiceRequest struct {
	NodeID     string `json:"node_id"`
	Name       string `json:"name"`
	Type       int    `json:"type"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	TLS        bool   `json:"tls"`
	TLSCertPath string `json:"tls_cert_path"`
	TLSKeyPath  string `json:"tls_key_path"`
	Enabled    bool   `json:"enabled"`
	Config     string `json:"config"`
	RuleGroups string `json:"rule_groups"`
}

type UpdateServiceRequest struct {
	Name       *string `json:"name"`
	Type       *int    `json:"type"`
	Host       *string `json:"host"`
	Port       *int    `json:"port"`
	TLS        *bool   `json:"tls"`
	TLSCertPath *string `json:"tls_cert_path"`
	TLSKeyPath  *string `json:"tls_key_path"`
	Enabled    *bool   `json:"enabled"`
	Config     *string `json:"config"`
	RuleGroups *string `json:"rule_groups"`
}

func (s *HoneypotServiceService) List(nodeID string) ([]models.HoneypotService, error) {
	var services []models.HoneypotService
	q := database.DB.Order("id ASC")
	if nodeID != "" {
		q = q.Where("node_id = ?", nodeID)
	}
	err := q.Find(&services).Error
	return services, err
}

func (s *HoneypotServiceService) GetByID(id uint) (*models.HoneypotService, error) {
	var svc models.HoneypotService
	if err := database.DB.First(&svc, id).Error; err != nil {
		return nil, err
	}
	return &svc, nil
}

func (s *HoneypotServiceService) Create(req CreateServiceRequest) (*models.HoneypotService, error) {
	if req.Port == 80 || req.Port == 22 || req.Port == 443 {
		return nil, fmt.Errorf("port %d is forbidden", req.Port)
	}

	protocol := models.HoneypotProtocolMap[req.Type]
	if protocol == "" {
		protocol = "http"
	}

	svc := models.HoneypotService{
		NodeID:      req.NodeID,
		Name:        req.Name,
		Type:        req.Type,
		Protocol:    protocol,
		Host:        req.Host,
		Port:        req.Port,
		TLS:         req.TLS,
		TLSCertPath: req.TLSCertPath,
		TLSKeyPath:  req.TLSKeyPath,
		Enabled:     req.Enabled,
		Config:      req.Config,
		Status:      "stopped",
		RuleGroups:  req.RuleGroups,
	}
	if svc.Host == "" {
		svc.Host = "0.0.0.0"
	}

	if err := database.DB.Create(&svc).Error; err != nil {
		return nil, err
	}
	return &svc, nil
}

func (s *HoneypotServiceService) Update(id uint, req UpdateServiceRequest) (*models.HoneypotService, error) {
	svc, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Type != nil {
		updates["type"] = *req.Type
		updates["protocol"] = models.HoneypotProtocolMap[*req.Type]
	}
	if req.Host != nil {
		updates["host"] = *req.Host
	}
	if req.Port != nil {
		if *req.Port == 80 || *req.Port == 22 || *req.Port == 443 {
			return nil, fmt.Errorf("port %d is forbidden", *req.Port)
		}
		updates["port"] = *req.Port
	}
	if req.TLS != nil {
		updates["tls"] = *req.TLS
	}
	if req.TLSCertPath != nil {
		updates["tls_cert_path"] = *req.TLSCertPath
	}
	if req.TLSKeyPath != nil {
		updates["tls_key_path"] = *req.TLSKeyPath
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Config != nil {
		updates["config"] = *req.Config
	}
	if req.RuleGroups != nil {
		updates["rule_groups"] = *req.RuleGroups
	}

	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(svc).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	return s.GetByID(id)
}

func (s *HoneypotServiceService) Delete(id uint) error {
	return database.DB.Delete(&models.HoneypotService{}, id).Error
}

func (s *HoneypotServiceService) UpdateStatus(id uint, status string) error {
	return database.DB.Model(&models.HoneypotService{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (s *HoneypotServiceService) GetServiceConfig(svc *models.HoneypotService) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"id":       svc.ID,
		"name":     svc.Name,
		"type":     svc.Type,
		"protocol": svc.Protocol,
		"host":     svc.Host,
		"port":     svc.Port,
		"tls":      svc.TLS,
		"enabled":  svc.Enabled,
	}
	if svc.Config != "" {
		var extra map[string]interface{}
		if err := json.Unmarshal([]byte(svc.Config), &extra); err == nil {
			for k, v := range extra {
				config[k] = v
			}
		}
	}
	if svc.RuleGroups != "" {
		var groups []uint
		if err := json.Unmarshal([]byte(svc.RuleGroups), &groups); err == nil {
			config["rule_groups"] = groups
		}
	}
	return config, nil
}
