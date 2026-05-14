package service

import (
	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
)

type RuleService struct{}

func NewRuleService() *RuleService {
	return &RuleService{}
}

func (s *RuleService) ListGroups(protocol string) ([]models.RuleGroup, error) {
	var groups []models.RuleGroup
	db := database.DB.Preload("Rules")
	if protocol != "" {
		db = db.Where("protocol = ?", protocol)
	}
	err := db.Order("id ASC").Find(&groups).Error
	return groups, err
}

func (s *RuleService) GetGroupByID(id uint) (*models.RuleGroup, error) {
	var group models.RuleGroup
	if err := database.DB.Preload("Rules").First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
}

func (s *RuleService) CreateGroup(req CreateGroupRequest) (*models.RuleGroup, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	group := models.RuleGroup{
		Name:        req.Name,
		Protocol:    req.Protocol,
		Description: req.Description,
		Enabled:     enabled,
	}
	if err := database.DB.Create(&group).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"`
	Protocol    *string `json:"protocol"`
	Description *string `json:"description"`
	Enabled     *bool   `json:"enabled"`
}

func (s *RuleService) UpdateGroup(id uint, req UpdateGroupRequest) (*models.RuleGroup, error) {
	group, err := s.GetGroupByID(id)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Protocol != nil {
		updates["protocol"] = *req.Protocol
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if len(updates) > 0 {
		if err := database.DB.Model(group).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.GetGroupByID(id)
}

func (s *RuleService) DeleteGroup(id uint) error {
	database.DB.Where("group_id = ?", id).Delete(&models.Rule{})
	return database.DB.Delete(&models.RuleGroup{}, id).Error
}

type CreateRuleRequest struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
	Action  string `json:"action"`
	Enabled *bool  `json:"enabled"`
}

func (s *RuleService) CreateRule(groupID uint, req CreateRuleRequest) (*models.Rule, error) {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	action := req.Action
	if action == "" {
		action = "alert"
	}
	rule := models.Rule{
		GroupID: groupID,
		Type:    req.Type,
		Pattern: req.Pattern,
		Action:  action,
		Enabled: enabled,
	}
	if err := database.DB.Create(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

type UpdateRuleRequest struct {
	Type    *string `json:"type"`
	Pattern *string `json:"pattern"`
	Action  *string `json:"action"`
	Enabled *bool   `json:"enabled"`
}

func (s *RuleService) UpdateRule(groupID, ruleID uint, req UpdateRuleRequest) (*models.Rule, error) {
	var rule models.Rule
	if err := database.DB.Where("id = ? AND group_id = ?", ruleID, groupID).First(&rule).Error; err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Pattern != nil {
		updates["pattern"] = *req.Pattern
	}
	if req.Action != nil {
		updates["action"] = *req.Action
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&rule).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	database.DB.Where("id = ? AND group_id = ?", ruleID, groupID).First(&rule)
	return &rule, nil
}

func (s *RuleService) DeleteRule(groupID, ruleID uint) error {
	return database.DB.Where("id = ? AND group_id = ?", ruleID, groupID).Delete(&models.Rule{}).Error
}

func (s *RuleService) GetRulesForService(groupIDs []uint) ([]models.Rule, error) {
	var rules []models.Rule
	err := database.DB.Where("group_id IN ? AND enabled = true", groupIDs).Find(&rules).Error
	return rules, err
}

func (s *RuleService) GetAllEnabled() ([]models.RuleGroup, error) {
	var groups []models.RuleGroup
	err := database.DB.Preload("Rules", "enabled = ?", true).
		Where("enabled = ?", true).
		Find(&groups).Error
	return groups, err
}
