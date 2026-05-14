package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/chenflux/honeywatch/internal/hub"
	"github.com/chenflux/honeywatch/internal/service"
)

type LogHandler struct {
	logService   *service.LogService
	geoService   *service.IPGeoService
	alertService *service.AlertService
}

func NewLogHandler() *LogHandler {
	return &LogHandler{
		logService:  service.NewLogService(),
		geoService:  service.NewIPGeoService(),
		alertService: service.NewAlertService(),
	}
}

func (h *LogHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req service.CreateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	logEntry, err := h.logService.Create(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if logEntry.IsAttack {
		hub.PublishAttack(logEntry)
		log.Printf("[LogHandler] Attack detected from %s: %s %s", logEntry.ClientIP, logEntry.Method, logEntry.Path)
	}
	hub.BroadcastStats()
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"log":     logEntry,
	})
}

func (h *LogHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	var reqs []service.CreateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.logService.BatchCreate(reqs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	hub.BroadcastStats()
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"count":   len(reqs),
	})
}

func (h *LogHandler) Query(w http.ResponseWriter, r *http.Request) {
	q := service.LogQuery{
		NodeID:       r.URL.Query().Get("node_id"),
		ClientIP:     r.URL.Query().Get("client_ip"),
		AttackType:   r.URL.Query().Get("attack_type"),
		StartTime:    r.URL.Query().Get("start_time"),
		EndTime:      r.URL.Query().Get("end_time"),
		Page:         1,
		PageSize:     20,
	}
	if sid := r.URL.Query().Get("service_id"); sid != "" {
		if id, err := strconv.ParseUint(sid, 10, 64); err == nil {
			q.ServiceID = uint(id)
		}
	}
	if ht := r.URL.Query().Get("honeypot_type"); ht != "" {
		if t, err := strconv.Atoi(ht); err == nil {
			q.HoneypotType = t
		}
	}
	if attack := r.URL.Query().Get("is_attack"); attack != "" {
		isAttack := attack == "true"
		q.IsAttack = &isAttack
	}
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			q.Page = p
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if s, err := strconv.Atoi(ps); err == nil && s > 0 {
			q.PageSize = s
		}
	}

	logs, total, err := h.logService.Query(q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to query logs")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"logs":      logs,
		"total":     total,
		"page":      q.Page,
		"page_size": q.PageSize,
	})
}

func (h *LogHandler) Stats(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if h := r.URL.Query().Get("hours"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v > 0 {
			hours = v
		}
	}
	summary, err := h.logService.GetSummary(hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   summary,
	})
}

func (h *LogHandler) Trend(w http.ResponseWriter, r *http.Request) {
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "7d"
	}
	honeypotType := 0
	if ht := r.URL.Query().Get("honeypot_type"); ht != "" {
		if t, err := strconv.Atoi(ht); err == nil {
			honeypotType = t
		}
	}
	data, err := h.logService.GetTrend(honeypotType, timeRange)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get trend")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"trend_data": data,
	})
}

func (h *LogHandler) Export(w http.ResponseWriter, r *http.Request) {
	q := service.LogQuery{
		StartTime: r.URL.Query().Get("start_time"),
		EndTime:   r.URL.Query().Get("end_time"),
	}
	logs, err := h.logService.Export(q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to export logs")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=logs_export.json")
	json.NewEncoder(w).Encode(logs)
}
