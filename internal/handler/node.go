package handler

import (
	"encoding/json"
	"net/http"

	"github.com/chenflux/honeywatch/internal/service"
	"github.com/gorilla/mux"
)

type NodeHandler struct {
	nodeService *service.NodeService
}

func NewNodeHandler() *NodeHandler {
	return &NodeHandler{nodeService: service.NewNodeService()}
}

func (h *NodeHandler) List(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.nodeService.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list nodes")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"nodes":   nodes,
	})
}

func (h *NodeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	node, err := h.nodeService.GetByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"node":    node,
	})
}

func (h *NodeHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AgentID == "" || req.IPAddress == "" {
		writeError(w, http.StatusBadRequest, "agent_id and ip_address required")
		return
	}

	node, err := h.nodeService.Register(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"node":    node,
	})
}

func (h *NodeHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req service.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.nodeService.Heartbeat(id, req); err != nil {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "heartbeat updated",
	})
}

func (h *NodeHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.nodeService.UpdateStatus(id, req.Status); err != nil {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "status updated",
	})
}

func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.nodeService.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, "node not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "node deleted",
	})
}
