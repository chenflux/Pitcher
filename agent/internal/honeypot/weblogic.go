package honeypot

import (
	"net/http"
	"strings"
	"time"
)

type WebLogicHoneypot struct{}

func (h *WebLogicHoneypot) Name() string { return "WebLogic" }
func (h *WebLogicHoneypot) Type() int    { return 2 }

var weblogicExploits = []string{
	"/console/css/%252e%252e%252f",
	"/_async/AsyncResponseService",
	"/wls-wsat/CoordinatorPortType",
	"/wls-wsat/RegistrationPortTypeRPC",
	"/console/login/LoginForm.jsp",
	"/console/j_security_check",
}

func (h *WebLogicHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	version := getString(cfg, "version", "Enterprise Edition 12c")
	delayMs := getInt(cfg, "delay_ms", 0)

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	path := r.URL.Path
	attack := h.detectAttack(path)

	if path == "/" || path == "/login" {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(200)
			w.Write([]byte(h.buildLoginPage(version)))
		} else {
			w.Header().Set("Location", "/login")
			w.WriteHeader(302)
		}
		return true, attack
	}

	if strings.Contains(path, "/j_security_check") {
		w.Header().Set("Location", "/console")
		w.WriteHeader(302)
		return true, attack
	}

	if strings.Contains(path, "/console") {
		w.Header().Set("Location", "/login")
		w.WriteHeader(302)
		return true, attack
	}

	for _, exploit := range weblogicExploits {
		if strings.Contains(path, exploit) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(`<h2>Error 500--Internal Server Error</h2>`))
			return true, &AttackInfo{Type: "weblogic_exploit", Detail: map[string]string{"path": path}, Source: "path"}
		}
	}

	return false, attack
}

func (h *WebLogicHoneypot) buildLoginPage(version string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>WebLogic Server Login</title>
<style>
body{font-family:Arial,sans-serif;background-color:#f0f0f0;margin:0;padding:0}
.container{max-width:400px;margin:100px auto;background-color:white;padding:20px;border-radius:8px;box-shadow:0 0 10px rgba(0,0,0,.1)}
h1{text-align:center;color:#333}
.form-group{margin-bottom:15px}
label{display:block;margin-bottom:5px;color:#666}
input[type="text"],input[type="password"]{width:100%;padding:8px;border:1px solid #ddd;border-radius:4px}
button{width:100%;padding:10px;background-color:#1a73e8;color:white;border:none;border-radius:4px;cursor:pointer}
button:hover{background-color:#1557b0}
</style>
</head>
<body>
<div class="container">
<h1>WebLogic Server</h1>
<p style="text-align:center;color:#666">` + version + `</p>
<form action="/j_security_check" method="POST">
<div class="form-group"><label>Username:</label><input type="text" name="j_username" required></div>
<div class="form-group"><label>Password:</label><input type="password" name="j_password" required></div>
<button type="submit">Login</button>
</form>
</div>
</body>
</html>`
}

func (h *WebLogicHoneypot) detectAttack(path string) *AttackInfo {
	for _, exploit := range weblogicExploits {
		if strings.Contains(path, exploit) {
			return &AttackInfo{Type: "weblogic_exploit", Detail: map[string]string{"pattern": exploit}, Source: "path"}
		}
	}
	return nil
}
