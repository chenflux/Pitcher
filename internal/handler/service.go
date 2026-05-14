package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/chenflux/pitcher/internal/service"
	"github.com/gorilla/mux"
)

type ServiceHandler struct {
	svcService *service.HoneypotServiceService
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{svcService: service.NewHoneypotServiceService()}
}

func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	services, err := h.svcService.List(nodeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list services")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"services": services,
	})
}

func (h *ServiceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	svc, err := h.svcService.GetByID(uint(id))
	if err != nil {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"service": svc,
	})
}

func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := h.svcService.Create(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"service": svc,
	})
}

func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	var req service.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	svc, err := h.svcService.Update(uint(id), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"service": svc,
	})
}

func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	if err := h.svcService.Delete(uint(id)); err != nil {
		writeError(w, http.StatusNotFound, "service not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "service deleted",
	})
}

func (h *ServiceHandler) Start(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	h.svcService.UpdateStatus(uint(id), "running")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "service start command sent",
	})
}

func (h *ServiceHandler) Stop(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	h.svcService.UpdateStatus(uint(id), "stopped")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "service stop command sent",
	})
}

func (h *ServiceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service id")
		return
	}
	h.svcService.UpdateStatus(uint(id), "restarting")
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "service restart command sent",
	})
}
