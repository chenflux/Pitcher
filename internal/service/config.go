package service

import (
	"encoding/json"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
)

type ConfigService struct{}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) List() ([]models.ConfigTemplate, error) {
	var configs []models.ConfigTemplate
	err := database.DB.Order("id ASC").Find(&configs).Error
	return configs, err
}

func (s *ConfigService) GetByID(id uint) (*models.ConfigTemplate, error) {
	var cfg models.ConfigTemplate
	if err := database.DB.First(&cfg, id).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

type CreateConfigRequest struct {
	Name       string `json:"name"`
	Description string `json:"description"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	ServiceIDs []uint `json:"service_ids"`
	NodeID     string `json:"node_id"`
}

func (s *ConfigService) Create(req CreateConfigRequest) (*models.ConfigTemplate, error) {
	var sidJSON []byte
	if len(req.ServiceIDs) > 0 {
		sidJSON, _ = json.Marshal(req.ServiceIDs)
	}
	cfg := models.ConfigTemplate{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Content:     req.Content,
		Version:     1,
		ServiceIDs:  string(sidJSON),
		NodeID:      req.NodeID,
	}
	if err := database.DB.Create(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

type UpdateConfigRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	Content     *string `json:"content"`
	ServiceIDs  *[]uint  `json:"service_ids"`
	NodeID      *string  `json:"node_id"`
}

func (s *ConfigService) Update(id uint, req UpdateConfigRequest) (*models.ConfigTemplate, error) {
	cfg, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Content != nil {
		updates["content"] = *req.Content
		updates["version"] = cfg.Version + 1
	}
	if req.ServiceIDs != nil {
		sidJSON, _ := json.Marshal(*req.ServiceIDs)
		updates["service_ids"] = string(sidJSON)
	}
	if req.NodeID != nil {
		updates["node_id"] = *req.NodeID
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		if err := database.DB.Model(cfg).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.GetByID(id)
}

func (s *ConfigService) GetByNodeID(nodeID string) ([]models.ConfigTemplate, error) {
	var configs []models.ConfigTemplate
	err := database.DB.Where("node_id = ? OR node_id = '' OR node_id IS NULL", nodeID).Order("id ASC").Find(&configs).Error
	return configs, err
}

func (s *ConfigService) Delete(id uint) error {
	return database.DB.Delete(&models.ConfigTemplate{}, id).Error
}
