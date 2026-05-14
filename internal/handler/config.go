package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chenflux/pitcher/internal/service"
	"github.com/gorilla/mux"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{configService: service.NewConfigService()}
}

func (h *ConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	configs, err := h.configService.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list configs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"configs": configs,
	})
}

func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid config id")
		return
	}
	cfg, err := h.configService.GetByID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "config not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"config":  cfg,
	})
}

func (h *ConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "name and content required")
		return
	}
	cfg, err := h.configService.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"config":  cfg,
	})
}

func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid config id")
		return
	}
	var req service.UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	cfg, err := h.configService.Update(uint(id), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"config":  cfg,
	})
}

func (h *ConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid config id")
		return
	}
	if err := h.configService.Delete(uint(id)); err != nil {
		writeError(w, http.StatusNotFound, "config not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "config deleted",
	})
}

func (h *ConfigHandler) Push(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid config id")
		return
	}
	cfg, err := h.configService.GetByID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "config not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"message":   "config push queued",
		"config_id": cfg.ID,
	})
}

func (h *ConfigHandler) ListByNodeID(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		writeError(w, http.StatusBadRequest, "node_id required")
		return
	}
	configs, err := h.configService.GetByNodeID(nodeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list configs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"configs":  configs,
	})
}
