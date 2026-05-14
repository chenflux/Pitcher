package honeypot

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type NginxHoneypot struct{}

func (h *NginxHoneypot) Name() string { return "Nginx" }
func (h *NginxHoneypot) Type() int    { return 6 }

var nginxSensitivePaths = []string{
	"/.git/", "/.env", "/.htaccess", "/.nginx",
	"/etc/passwd", "/proc/self", "/wp-admin", "/wp-login",
	"/admin", "/backup", "/config", "/debug", "/status",
	"/server-status", "/server-info", "/stub_status",
	"/.well-known/security.txt",
}

func (h *NginxHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	serverVersion := getString(cfg, "server_version", "nginx/1.20.1")
	delayMs := getInt(cfg, "delay_ms", 0)

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	w.Header().Set("Server", serverVersion)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	attack := detectNginxAttack(r)
	if r.URL.Path == "/" {
		w.WriteHeader(200)
		w.Write([]byte(nginxWelcomePage))
	} else if attack != nil {
		w.WriteHeader(403)
		w.Write([]byte(`<html><head><title>403 Forbidden</title></head><body><center><h1>403 Forbidden</h1></center><hr><center>` + serverVersion + `</center></body></html>`))
	} else {
		w.WriteHeader(404)
		w.Write([]byte(`<html><head><title>404 Not Found</title></head><body><center><h1>404 Not Found</h1></center><hr><center>` + serverVersion + `</center></body></html>`))
	}
	return true, attack
}

func detectNginxAttack(r *http.Request) *AttackInfo {
	path := r.URL.Path
	for _, p := range nginxSensitivePaths {
		if strings.Contains(path, p) {
			return &AttackInfo{Type: "sensitive_path", Detail: map[string]string{"path": path}, Source: "nginx"}
		}
	}
	ua := strings.ToLower(r.UserAgent())
	for _, scanner := range []string{"nikto", "nessus", "sqlmap", "burp", "dirb", "gobuster", "ffuf", "wfuzz", "nmap"} {
		if strings.Contains(ua, scanner) {
			return &AttackInfo{Type: "scanner", Detail: map[string]string{"user_agent": r.UserAgent()}, Source: "nginx"}
		}
	}
	return nil
}

const nginxWelcomePage = `<!DOCTYPE html>
<html>
<head><title>Welcome to nginx!</title>
<style>body{width:35em;margin:0 auto;font-family:Tahoma,Verdana,Arial,sans-serif}</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>
<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>
<p><em>Thank you for using nginx.</em></p>
</body>
</html>`

func getString(cfg map[string]interface{}, key, def string) string {
	if cfg == nil {
		return def
	}
	if v, ok := cfg[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func getInt(cfg map[string]interface{}, key string, def int) int {
	if cfg == nil {
		return def
	}
	if v, ok := cfg[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				return i
			}
		}
	}
	return def
}

func getBool(cfg map[string]interface{}, key string, def bool) bool {
	if cfg == nil {
		return def
	}
	if v, ok := cfg[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}
