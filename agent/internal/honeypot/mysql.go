package honeypot

import (
	"net/http"
)

type MySQLHoneypot struct{}

func (h *MySQLHoneypot) Name() string { return "MySQL" }
func (h *MySQLHoneypot) Type() int    { return 3 }

func (h *MySQLHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	banner := getString(cfg, "banner", "MySQL v5.7.33 Community Server (GPL)\r\n")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(banner))
	return true, nil
}
