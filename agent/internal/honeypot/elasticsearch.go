package honeypot

import (
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ElasticsearchHoneypot struct{}

func (h *ElasticsearchHoneypot) Name() string { return "Elasticsearch" }
func (h *ElasticsearchHoneypot) Type() int    { return 1 }

var esScannerPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(Nikto|Nessus|OpenVAS|Masscan|Zmap|sqlmap|Burp|Wfuzz|Dirb|Gobuster|ffuf)`),
}

var esAttackPatterns = map[string][]string{
	"directory_traversal": {"../", "/etc/passwd", "/proc/self/environ"},
	"sensitive_files":     {"/.git/", "/.env", "/robots.txt"},
	"exploit_paths":       {"/_plugin/marvel/", "/_plugin/head/", "/_river/", "/_snapshot/"},
}

func (h *ElasticsearchHoneypot) HandleHTTP(w http.ResponseWriter, r *http.Request, body []byte, cfg map[string]interface{}) (bool, *AttackInfo) {
	version := getString(cfg, "version", "7.4.2")
	clusterName := getString(cfg, "cluster_name", "elasticsearch")
	delayMs := getInt(cfg, "delay_ms", 0)

	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	path := r.URL.Path
	cleanPath := strings.TrimRight(strings.SplitN(path, "?", 2)[0], "/")

	attack := h.detectAttack(r, body)
	if attack != nil {
		attack.Type = cleanPath
	}

	statusCode := 200
	var data string

	switch {
	case cleanPath == "" || cleanPath == "/":
		data = `{"name":"` + h.nodeID() + `","cluster_name":"` + clusterName + `","cluster_uuid":"zaB2D2C_Qqabg7a3uzUgIw","version":{"number":"` + version + `","build_flavor":"default","build_type":"docker","build_hash":"2f90bbf7b93631e52bafb59b3b049cb44ec25e96","build_date":"2019-10-28T20:40:44.881551Z","lucene_version":"8.2.0"},"tagline":"You Know, for Search"}`
	case strings.HasPrefix(cleanPath, "/_nodes"):
		data = `{"cluster_name":"` + clusterName + `","nodes":{}}`
	case strings.HasPrefix(cleanPath, "/_cat/indices"):
		data = "yellow open read_me Zp3uE-8tQvGhjfQqQMW6-g 1 1 1 0 4.4kb 4.4kb"
	case strings.HasPrefix(cleanPath, "/_cat/health"):
		data = "1612345678 12:34:56 " + clusterName + " yellow 1 1 2 2 0 0 0 0 - 50.0%"
	case strings.HasPrefix(cleanPath, "/_cluster/health"):
		data = `{"cluster_name":"` + clusterName + `","status":"yellow","number_of_nodes":1}`
	case strings.HasPrefix(cleanPath, "/_search"):
		data = `{"took":0,"timed_out":false,"hits":{"total":{"value":0},"hits":[]}}`
	case strings.HasPrefix(cleanPath, "/_mapping"):
		data = `{}`
	default:
		data = `{"error":{"type":"index_not_found_exception","reason":"no such index"},"status":404}`
		statusCode = 404
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(data))
	return true, attack
}

func (h *ElasticsearchHoneypot) nodeID() string {
	return "b5c8bbb2116f"
}

func (h *ElasticsearchHoneypot) detectAttack(r *http.Request, body []byte) *AttackInfo {
	path := r.URL.Path

	for attackType, patterns := range esAttackPatterns {
		for _, p := range patterns {
			if strings.Contains(path, p) {
				return &AttackInfo{Type: attackType, Detail: map[string]string{"pattern": p}, Source: "path"}
			}
		}
	}

	ua := r.UserAgent()
	for _, re := range esScannerPatterns {
		if re.MatchString(ua) {
			return &AttackInfo{Type: "scanner_detection", Detail: map[string]string{"user_agent": ua}, Source: "header"}
		}
	}

	bodyStr := string(body)
	if bodyStr != "" {
		sqlPatterns := []string{"UNION SELECT", "' OR 1=1", "; DROP", "INSERT INTO"}
		for _, p := range sqlPatterns {
			if strings.Contains(strings.ToUpper(bodyStr), strings.ToUpper(p)) {
				return &AttackInfo{Type: "sql_injection", Detail: map[string]string{"pattern": p}, Source: "body"}
			}
		}
		cmdPatterns := []string{";cat", "|whoami", "&&ls", "`id`"}
		for _, p := range cmdPatterns {
			if strings.Contains(strings.ToLower(bodyStr), strings.ToLower(p)) {
				return &AttackInfo{Type: "command_injection", Detail: map[string]string{"pattern": p}, Source: "body"}
			}
		}
	}

	return nil
}
