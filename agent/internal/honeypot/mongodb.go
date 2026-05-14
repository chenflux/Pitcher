package honeypot

import (
	"net/http"
	"strings"
	"time"
)

type MongoDBHoneypot struct{}

func (h *MongoDBHoneypot) Name() string { return "MongoDB" }
func (h *MongoDBHoneypot) Type() int    { return 5 }

func (h *MongoDBHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	path := strings.TrimRight(strings.SplitN(r.URL.Path, "?", 2)[0], "/")
	delayMs := getInt(cfg, "delay_ms", 0)
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	switch path {
	case "", "/", "/ping":
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":1.0}`))
	case "/admin/$cmd":
		w.WriteHeader(200)
		w.Write([]byte(`{"isMaster":true,"maxBsonObjectSize":16777216,"ok":1.0}`))
	default:
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":1.0}`))
	}
	return true, nil
}
