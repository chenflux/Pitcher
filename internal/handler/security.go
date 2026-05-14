package handler

import (
	"encoding/json"
	"net/http"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
	"github.com/chenflux/pitcher/internal/service"
)

type SecurityHandler struct {
	securityService *service.SecurityService
}

func NewSecurityHandler() *SecurityHandler {
	return &SecurityHandler{securityService: service.NewSecurityService()}
}

func (h *SecurityHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	listType := r.URL.Query().Get("list_type")
	page := 1
	pageSize := 50

	entries, total, err := h.securityService.ListEntries(listType, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list entries")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"entries":   entries,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *SecurityHandler) AddEntry(w http.ResponseWriter, r *http.Request) {
	_, username, _, ok := getContextUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IPAddress       string `json:"ip_address"`
		ListType        string `json:"list_type"`
		Reason          string `json:"reason"`
		DurationMinutes int    `json:"duration_minutes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.IPAddress == "" || req.ListType == "" {
		writeError(w, http.StatusBadRequest, "ip_address and list_type required")
		return
	}

	duration := req.DurationMinutes
	if duration <= 0 {
		duration = -1
	}

	if err := h.securityService.AddToList(req.IPAddress, models.SecurityListType(req.ListType), req.Reason, username, duration); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add entry")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "entry added",
	})
}

func (h *SecurityHandler) RemoveEntry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IPAddress string `json:"ip_address"`
		ListType  string `json:"list_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.IPAddress == "" || req.ListType == "" {
		writeError(w, http.StatusBadRequest, "ip_address and list_type required")
		return
	}

	if err := h.securityService.RemoveFromList(req.IPAddress, models.SecurityListType(req.ListType)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove entry")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "entry removed",
	})
}

func (h *SecurityHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var blacklistCount, whitelistCount, graylistCount int64

	database.DB.Model(&models.SecurityEntry{}).Where("list_type = ?", "blacklist").Count(&blacklistCount)
	database.DB.Model(&models.SecurityEntry{}).Where("list_type = ?", "whitelist").Count(&whitelistCount)
	database.DB.Model(&models.SecurityEntry{}).Where("list_type = ?", "graylist").Count(&graylistCount)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats": map[string]int64{
			"blacklist": blacklistCount,
			"whitelist": whitelistCount,
			"graylist":  graylistCount,
		},
	})
}