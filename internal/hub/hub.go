package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/chenflux/honeywatch/internal/database"
	internal_models "github.com/chenflux/honeywatch/internal/models"
	"gorm.io/gorm"
)

type EventType string

const (
	EventNewAttack   EventType = "new_attack"
	EventNewLog      EventType = "new_log"
	EventNodeStatus  EventType = "node_status"
	EventStatsUpdate EventType = "stats_update"
)

type Event struct {
	Type      EventType       `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type AttackEventData struct {
	ID           uint   `json:"id"`
	NodeID       string `json:"node_id"`
	ServiceID    uint   `json:"service_id"`
	ClientIP     string `json:"client_ip"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	AttackType   string `json:"attack_type"`
	AttackDetail string `json:"attack_detail"`
	UserAgent    string `json:"user_agent"`
	HoneypotType int    `json:"honeypot_type"`
	CreatedAt    string `json:"created_at"`
}

type StatsEventData struct {
	TotalRequests   int64 `json:"total_requests"`
	AttackCount     int64 `json:"attack_count"`
	OnlineNodes     int   `json:"online_nodes"`
	RunningServices int   `json:"running_services"`
}

type subEntry struct {
	types []EventType
	done  bool
}

type Hub struct {
	mu         sync.Mutex
	subs       map[chan Event]*subEntry
}

var GlobalHub *Hub

func init() {
	GlobalHub = NewHub()
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[chan Event]*subEntry),
	}
}

func (h *Hub) Register(ch chan Event, types []EventType) {
	h.mu.Lock()
	h.subs[ch] = &subEntry{types: types}
	h.mu.Unlock()
}

func (h *Hub) Unregister(ch chan Event) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

func (h *Hub) Publish(event Event) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return
	}
	event.Data = data

	h.mu.Lock()
	defer h.mu.Unlock()

	for ch, entry := range h.subs {
		if entry.done {
			continue
		}
		for _, t := range entry.types {
			if t == event.Type || t == "" {
				select {
				case ch <- event:
				default:
					entry.done = true
					delete(h.subs, ch)
				}
				break
			}
		}
	}
}

func (h *Hub) PublishAttack(reqLog *internal_models.RequestLog) {
	data, _ := json.Marshal(AttackEventData{
		ID:           reqLog.ID,
		NodeID:       reqLog.NodeID,
		ServiceID:    reqLog.ServiceID,
		ClientIP:     reqLog.ClientIP,
		Method:       reqLog.Method,
		Path:         reqLog.Path,
		AttackType:   reqLog.AttackType,
		AttackDetail: reqLog.AttackDetail,
		UserAgent:    reqLog.UserAgent,
		HoneypotType: reqLog.HoneypotType,
		CreatedAt:    reqLog.CreatedAt.Format(time.RFC3339),
	})
	h.Publish(Event{
		Type:      EventNewAttack,
		Timestamp: time.Now(),
		Data:      data,
	})
}

func (h *Hub) PublishStats(stats *StatsEventData) {
	data, _ := json.Marshal(stats)
	h.Publish(Event{
		Type:      EventStatsUpdate,
		Timestamp: time.Now(),
		Data:      data,
	})
}

func (h *Hub) BroadcastStats() {
	go func() {
		if stats, err := h.getCurrentStats(); err == nil {
			h.PublishStats(stats)
		}
	}()
}

func (h *Hub) getCurrentStats() (*StatsEventData, error) {
	stats := &StatsEventData{}
	since := time.Now().Add(-24 * time.Hour)

	if err := database.DB.Model(&internal_models.RequestLog{}).Where("created_at >= ?", since).Count(&stats.TotalRequests).Error; err != nil {
		return nil, err
	}
	if err := database.DB.Model(&internal_models.RequestLog{}).Where("created_at >= ? AND is_attack = true", since).Count(&stats.AttackCount).Error; err != nil {
		return nil, err
	}

	var onlineNodes int64
	if err := database.DB.Model(&internal_models.Node{}).Where("status = ?", "online").Count(&onlineNodes).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	stats.OnlineNodes = int(onlineNodes)

	var runningServices int64
	if err := database.DB.Model(&internal_models.HoneypotService{}).Where("status = ?", "running").Count(&runningServices).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	stats.RunningServices = int(runningServices)

	return stats, nil
}

func BroadcastStats() {
	GlobalHub.BroadcastStats()
}

func PublishAttack(reqLog *internal_models.RequestLog) {
	GlobalHub.PublishAttack(reqLog)
	log.Printf("[Hub] Published attack event for log ID %d from IP %s", reqLog.ID, reqLog.ClientIP)
}