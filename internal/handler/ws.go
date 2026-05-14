package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/chenflux/pitcher/internal/hub"
	"github.com/chenflux/pitcher/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSHandler struct{}

func NewWSHandler() *WSHandler {
	return &WSHandler{}
}

type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type SubscribePayload struct {
	Types []string `json:"types"`
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	tokenStr := ""
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr = parts[1]
		}
	}
	if tokenStr == "" {
		tokenStr = r.URL.Query().Get("token")
	}
	if tokenStr == "" {
		http.Error(w, "missing authorization", http.StatusUnauthorized)
		return
	}

	_, err := middleware.ParseToken(tokenStr)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	ch := make(chan hub.Event, 100)
	hub.GlobalHub.Register(ch, []hub.EventType{hub.EventNewAttack, hub.EventNewLog, hub.EventNodeStatus, hub.EventStatsUpdate})
	defer hub.GlobalHub.Unregister(ch)

	closed := make(chan struct{})

	go func() {
		defer close(closed)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var wsMsg WSMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				continue
			}

			switch wsMsg.Type {
			case "ping":
				conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"pong","timestamp":"`+time.Now().Format(time.RFC3339)+`"}`))
			case "subscribe":
				var sub SubscribePayload
				if err := json.Unmarshal(wsMsg.Payload, &sub); err == nil {
					types := make([]hub.EventType, len(sub.Types))
					for i, t := range sub.Types {
						types[i] = hub.EventType(t)
					}
					hub.GlobalHub.Register(ch, types)
				}
			}
		}
	}()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				return
			}
			eventMsg, err := json.Marshal(map[string]interface{}{
				"type":      event.Type,
				"timestamp": event.Timestamp.Format(time.RFC3339),
				"data":      event.Data,
			})
			if err != nil {
				continue
			}
			if err := conn.WriteMessage(websocket.TextMessage, eventMsg); err != nil {
				return
			}
		case <-closed:
			return
		}
	}
}

func RegisterWSRoutes(r *mux.Router) {
	wsHandler := NewWSHandler()
	r.Handle("/ws/events", http.HandlerFunc(wsHandler.HandleWS))
}