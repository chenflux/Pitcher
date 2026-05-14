package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chenflux/honeywatch/internal/service"
	"github.com/gorilla/mux"
)

type RuleHandler struct {
	ruleService *service.RuleService
}

func NewRuleHandler() *RuleHandler {
	return &RuleHandler{ruleService: service.NewRuleService()}
}

func (h *RuleHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	protocol := r.URL.Query().Get("protocol")
	groups, err := h.ruleService.ListGroups(protocol)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list rule groups")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"groups":  groups,
	})
}

func (h *RuleHandler) ListEnabled(w http.ResponseWriter, r *http.Request) {
	groups, err := h.ruleService.GetAllEnabled()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list rules")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"groups":  groups,
	})
}

func (h *RuleHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	group, err := h.ruleService.GetGroupByID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "rule group not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"group":   group,
	})
}

func (h *RuleHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req service.CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	group, err := h.ruleService.CreateGroup(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"group":   group,
	})
}

func (h *RuleHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var req service.UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	group, err := h.ruleService.UpdateGroup(uint(id), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"group":   group,
	})
}

func (h *RuleHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	if err := h.ruleService.DeleteGroup(uint(id)); err != nil {
		writeError(w, http.StatusNotFound, "rule group not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "rule group deleted",
	})
}

func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	var req service.CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rule, err := h.ruleService.CreateRule(uint(groupID), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"rule":    rule,
	})
}

func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	ruleID, err := strconv.ParseUint(mux.Vars(r)["rule_id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	var req service.UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	rule, err := h.ruleService.UpdateRule(uint(groupID), uint(ruleID), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"rule":    rule,
	})
}

func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid group id")
		return
	}
	ruleID, err := strconv.ParseUint(mux.Vars(r)["rule_id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	if err := h.ruleService.DeleteRule(uint(groupID), uint(ruleID)); err != nil {
		writeError(w, http.StatusNotFound, "rule not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "rule deleted",
	})
}
