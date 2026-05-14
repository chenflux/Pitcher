package honeypot

import (
	"net/http"
	"strings"
	"time"
)

type ApacheHoneypot struct{}

func (h *ApacheHoneypot) Name() string { return "Apache" }
func (h *ApacheHoneypot) Type() int    { return 7 }

var apacheSensitivePaths = []string{
	"/.git/", "/.env", "/.htaccess", "/.htpasswd",
	"/etc/passwd", "/proc/self", "/wp-admin", "/wp-login",
	"/admin", "/backup", "/phpmyadmin", "/pma",
	"/server-status", "/server-info", "/manager/html",
	"/cgi-bin/", "/icons/", "/manual/",
}

func (h *ApacheHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	serverVersion := getString(cfg, "server_version", "Apache/2.4.48 (Ubuntu)")
	delayMs := getInt(cfg, "delay_ms", 0)

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	w.Header().Set("Server", serverVersion)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	attack := detectApacheAttack(r)
	if r.URL.Path == "/" {
		w.WriteHeader(200)
		w.Write([]byte(apacheWelcomePage))
	} else if attack != nil {
		w.WriteHeader(403)
		w.Write([]byte(`<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN"><html><head><title>403 Forbidden</title></head><body><h1>Forbidden</h1><p>You don't have permission to access this resource.</p></body></html>`))
	} else {
		w.WriteHeader(404)
		w.Write([]byte(`<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML 2.0//EN"><html><head><title>404 Not Found</title></head><body><h1>Not Found</h1><p>The requested URL was not found on this server.</p></body></html>`))
	}
	return true, attack
}

func detectApacheAttack(r *http.Request) *AttackInfo {
	path := r.URL.Path
	for _, p := range apacheSensitivePaths {
		if strings.Contains(path, p) {
			return &AttackInfo{Type: "sensitive_path", Detail: map[string]string{"path": path}, Source: "apache"}
		}
	}
	ua := strings.ToLower(r.UserAgent())
	for _, scanner := range []string{"nikto", "nessus", "sqlmap", "burp", "dirb", "gobuster", "ffuf", "wfuzz", "nmap"} {
		if strings.Contains(ua, scanner) {
			return &AttackInfo{Type: "scanner", Detail: map[string]string{"user_agent": r.UserAgent()}, Source: "apache"}
		}
	}
	return nil
}

const apacheWelcomePage = `<!DOCTYPE html>
<html lang="en">
<head>
<title>Apache2 Ubuntu Default Page: It works</title>
<style>
body{background-color:#141414;color:#fff;font-family:'Ubuntu',sans-serif;font-size:14px}
.container{width:80%;margin:0 auto}
h1{color:#fff;background-color:#000}
</style>
</head>
<body>
<div class="container">
<h1>It works!</h1>
<p>This is the default welcome page used to test the correct operation of the Apache2 server after installation on Ubuntu systems. It is based on the equivalent page on Debian, hence the very similar look.</p>
</div>
</body>
</html>`
