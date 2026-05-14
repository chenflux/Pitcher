package honeypot

import (
	"net/http"
	"time"
)

type RedisHoneypot struct{}

func (h *RedisHoneypot) Name() string { return "Redis" }
func (h *RedisHoneypot) Type() int    { return 4 }

func (h *RedisHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	version := getString(cfg, "version", "Redis server v=6.2.5 sha=00000000:0 malloc=jemalloc-5.1.0 bits=64 build=0\r\n")
	delayMs := getInt(cfg, "delay_ms", 0)
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(version))
	return true, nil
}
