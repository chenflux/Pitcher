package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/chenflux/honeywatch/agent/internal/honeypot"
	"github.com/chenflux/honeywatch/agent/internal/server"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type AgentConfig struct {
	ServerURL    string `json:"server_url"`
	AgentID      string `json:"agent_id"`
	AgentToken   string `json:"agent_token"`
	Hostname     string `json:"hostname"`
	IPAddress    string `json:"ip_address"`
	PollInterval int    `json:"poll_interval"`
}

type LogEntry struct {
	NodeID       string `json:"node_id"`
	ServiceID    uint   `json:"service_id"`
	ClientIP     string `json:"client_ip"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	UserAgent    string `json:"user_agent"`
	HoneypotType int    `json:"honeypot_type"`
	StatusCode   int    `json:"status_code"`
	RequestBody  string `json:"request_body"`
	IsAttack     bool   `json:"is_attack"`
	AttackType   string `json:"attack_type"`
	AttackDetail string `json:"attack_detail"`
	Protocol     string `json:"protocol"`
}

type serviceRef struct {
	httpSrv *http.Server
	tcpSrv  *server.TCPServer
}

type cachedRule struct {
	Type    string         `json:"type"`
	Pattern string         `json:"pattern"`
	Action  string         `json:"action"`
	regex   *regexp.Regexp
}

type cachedRuleGroup struct {
	Protocol string       `json:"protocol"`
	Rules    []cachedRule `json:"rules"`
}

var (
	cfg         *AgentConfig
	nodeID      string
	nodeIDMu    sync.RWMutex
	logQueue    []LogEntry
	logQueueMu  sync.Mutex
	activeSvcs  map[uint]*serviceRef
	activeMu    sync.RWMutex
	ruleCache   []cachedRuleGroup
	ruleMu      sync.RWMutex
	configCache map[uint]AgentConfigTemplate
	configMu    sync.RWMutex
)

type AgentConfigTemplate struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	Version    int    `json:"version"`
	ServiceIDs []uint `json:"service_ids"`
}

const maxLogQueue = 10000

func main() {
	fmt.Println("============================================================")
	fmt.Println("  HoneyWatch Agent v2.0")
	fmt.Println("============================================================")

	agentID := os.Getenv("HONEYWATCH_AGENT_ID")
	if agentID == "" {
		agentID = loadAgentID()
		if agentID == "" {
			agentID = uuid.New().String()
			saveAgentID(agentID)
		}
	}

	cfg = &AgentConfig{
		ServerURL:    getEnv("HONEYWATCH_SERVER", "http://localhost:8090"),
		AgentID:      agentID,
		AgentToken:   os.Getenv("HONEYWATCH_AGENT_TOKEN"),
		Hostname:     getEnv("HONEYWATCH_HOSTNAME", mustHostname()),
		PollInterval: 30,
	}

	if cfg.AgentToken == "" {
		log.Fatal("[FATAL] HONEYWATCH_AGENT_TOKEN environment variable is required")
	}

	if ip := getEnv("HONEYWATCH_IP", ""); ip != "" {
		cfg.IPAddress = ip
	} else {
		cfg.IPAddress = autoDetectIP()
	}

	fmt.Printf("[1/5] Agent ID: %s\n", cfg.AgentID)
	fmt.Printf("[2/5] Server: %s\n", cfg.ServerURL)

	activeSvcs = make(map[uint]*serviceRef)

	register()

	fmt.Println("[3/5] Fetching rules...")
	fetchRules()

	go heartbeatLoop()
	go logFlushLoop()
	go fetchConfigLoop()
	go ruleRefreshLoop()

	fmt.Println("[4/5] Starting services...")
	fmt.Println("[5/5] Agent running")
	fmt.Println("============================================================")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down agent...")
	flushLogs()
	activeMu.Lock()
	for _, ref := range activeSvcs {
		if ref.httpSrv != nil {
			ref.httpSrv.Close()
		}
		if ref.tcpSrv != nil {
			ref.tcpSrv.Stop()
		}
	}
	activeMu.Unlock()
}

func getNodeID() string {
	nodeIDMu.RLock()
	defer nodeIDMu.RUnlock()
	return nodeID
}

func setNodeID(id string) {
	nodeIDMu.Lock()
	defer nodeIDMu.Unlock()
	nodeID = id
}

func agentGet(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Agent-Token", cfg.AgentToken)
	return http.DefaultClient.Do(req)
}

func agentPost(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Token", cfg.AgentToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func register() {
	reqBody := map[string]interface{}{
		"agent_id":      cfg.AgentID,
		"hostname":      cfg.Hostname,
		"ip_address":    cfg.IPAddress,
		"agent_version": "2.0.0",
		"os_name":       runtime.GOOS,
		"os_arch":       runtime.GOARCH,
	}
	body, _ := json.Marshal(reqBody)
	log.Printf("[Register] POST %s/data/nodes/register agent_id=%s", cfg.ServerURL, cfg.AgentID)
	resp, err := agentPost(cfg.ServerURL+"/data/nodes/register", body)
	if err != nil {
		log.Printf("[Register] HTTP error: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[Register] Response status=%d", resp.StatusCode)

	if resp.StatusCode != 201 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Printf("[Register] Non-201 response: %s", string(bodyBytes))
		return
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	nodeMap, hasNode := result["node"].(map[string]interface{})
	if !hasNode {
		log.Printf("[Register] ERROR: no 'node' in response: %v", result)
		return
	}
	nodeIDStr, hasID := nodeMap["id"].(string)
	if !hasID || nodeIDStr == "" {
		log.Printf("[Register] ERROR: empty node ID in response: %v", nodeMap)
		return
	}
	setNodeID(nodeIDStr)
	log.Printf("[Register] Success node_id=%s agent_id=%s", nodeIDStr, cfg.AgentID)
}

func heartbeatLoop() {
	ticker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		nid := getNodeID()
		if nid == "" {
			register()
			continue
		}
		cpuUsage, memUsage, diskUsage := collectResourceUsage()
		body, _ := json.Marshal(map[string]interface{}{
			"status":       "online",
			"cpu_usage":    cpuUsage,
			"memory_usage": memUsage,
			"disk_usage":   diskUsage,
		})
		url := fmt.Sprintf("%s/data/nodes/%s/heartbeat", cfg.ServerURL, nid)
		resp, err := agentPost(url, body)
		if err != nil {
			log.Printf("[Heartbeat] Failed: %v", err)
			continue
		}
		if resp.StatusCode == 404 {
			log.Printf("[Heartbeat] Node %s not found, re-registering", nid)
			setNodeID("")
			resp.Body.Close()
			register()
			continue
		}
		resp.Body.Close()
	}
}

func logFlushLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		flushLogs()
	}
}

func flushLogs() {
	logQueueMu.Lock()
	if len(logQueue) == 0 {
		logQueueMu.Unlock()
		return
	}
	logs := make([]LogEntry, len(logQueue))
	copy(logs, logQueue)
	logQueue = nil
	logQueueMu.Unlock()

	log.Printf("[FlushLogs] Sending %d logs", len(logs))
	body, _ := json.Marshal(logs)
	resp, err := agentPost(cfg.ServerURL+"/data/logs/batch", body)
	if err != nil {
		log.Printf("[FlushLogs] Failed: %v", err)
		logQueueMu.Lock()
		remaining := maxLogQueue - len(logQueue)
		if remaining < len(logs) {
			logs = logs[:remaining]
		}
		logQueue = append(logs, logQueue...)
		logQueueMu.Unlock()
		return
	}
	resp.Body.Close()
}

func enqueueLog(entry LogEntry) {
	logQueueMu.Lock()
	defer logQueueMu.Unlock()
	if len(logQueue) >= maxLogQueue {
		logQueue = logQueue[len(logQueue)/2:]
	}
	logQueue = append(logQueue, entry)
}

func fetchConfigLoop() {
	fetchAndStartServices()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		fetchAndStartServices()
	}
}

func fetchAndStartServices() {
	nid := getNodeID()
	if nid == "" {
		return
	}

	fetchConfigsForNode(nid)

	url := fmt.Sprintf("%s/data/services?node_id=%s", cfg.ServerURL, nid)
	resp, err := agentGet(url)
	if err != nil {
		log.Printf("Fetch services failed: %v", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Services []struct {
			ID      uint   `json:"id"`
			Type    int    `json:"type"`
			Name    string `json:"name"`
			Host    string `json:"host"`
			Port    int    `json:"port"`
			Enabled bool   `json:"enabled"`
			Status  string `json:"status"`
			Config  string `json:"config"`
		} `json:"services"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	activeIDs := make(map[uint]bool)
	for _, svc := range result.Services {
		activeIDs[svc.ID] = true
		if !svc.Enabled {
			continue
		}
		activeMu.RLock()
		_, exists := activeSvcs[svc.ID]
		activeMu.RUnlock()
		if exists {
			continue
		}
		if svc.Port == 0 {
			continue
		}
		hp := honeypot.Create(svc.Type)
		if hp == nil {
			continue
		}

		cfgContent := svc.Config
		if cfgContent == "" {
			cfgContent = getConfigForService(svc.ID)
		} else {
			override := getConfigForService(svc.ID)
			if override != "" {
				var base, ov map[string]interface{}
				if json.Unmarshal([]byte(cfgContent), &base) == nil && json.Unmarshal([]byte(override), &ov) == nil {
					for k, v := range ov {
						base[k] = v
					}
					if merged, err := json.Marshal(base); err == nil {
						cfgContent = string(merged)
					}
				}
			}
		}

		var cfgMap map[string]interface{}
		if cfgContent != "" {
			json.Unmarshal([]byte(cfgContent), &cfgMap)
		}

		switch svc.Type {
		case 3:
			go startTCPService(svc.ID, svc.Host, svc.Port, hp.Name(), "mysql", cfgMap)
		case 4:
			go startTCPService(svc.ID, svc.Host, svc.Port, hp.Name(), "redis", cfgMap)
		case 5:
			go startTCPService(svc.ID, svc.Host, svc.Port, hp.Name(), "mongodb", cfgMap)
		default:
			go startHTTPService(svc.ID, svc.Host, svc.Port, hp, cfgMap)
		}
	}

	activeMu.RLock()
	for id, ref := range activeSvcs {
		if !activeIDs[id] {
			activeMu.RUnlock()
			activeMu.Lock()
			if ref.httpSrv != nil {
				ref.httpSrv.Close()
			}
			if ref.tcpSrv != nil {
				ref.tcpSrv.Stop()
			}
			delete(activeSvcs, id)
			activeMu.Unlock()
			activeMu.RLock()
			log.Printf("Stopped service %d", id)
		}
	}
	activeMu.RUnlock()
}

func fetchConfigsForNode(nodeID string) {
	url := fmt.Sprintf("%s/data/configs/node?node_id=%s", cfg.ServerURL, nodeID)
	resp, err := agentGet(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Configs []AgentConfigTemplate `json:"configs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	configMu.Lock()
	configCache = make(map[uint]AgentConfigTemplate)
	for _, c := range result.Configs {
		for _, sid := range c.ServiceIDs {
			configCache[sid] = c
		}
	}
	configMu.Unlock()
	log.Printf("[Config] Loaded %d configs for node %s", len(result.Configs), nodeID)
}

func getConfigForService(svcID uint) string {
	configMu.RLock()
	defer configMu.RUnlock()
	if c, ok := configCache[svcID]; ok {
		return c.Content
	}
	return ""
}

func ruleRefreshLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		fetchRules()
	}
}

func fetchRules() {
	url := cfg.ServerURL + "/data/rules"
	resp, err := agentGet(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Groups []struct {
			Protocol string `json:"protocol"`
			Rules    []struct {
				Type    string `json:"type"`
				Pattern string `json:"pattern"`
				Action  string `json:"action"`
			} `json:"rules"`
		} `json:"groups"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	groups := make([]cachedRuleGroup, 0, len(result.Groups))
	for _, g := range result.Groups {
		cg := cachedRuleGroup{Protocol: g.Protocol}
		for _, r := range g.Rules {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				continue
			}
			cg.Rules = append(cg.Rules, cachedRule{
				Type:    r.Type,
				Pattern: r.Pattern,
				Action:  r.Action,
				regex:   re,
			})
		}
		if len(cg.Rules) > 0 {
			groups = append(groups, cg)
		}
	}

	ruleMu.Lock()
	ruleCache = groups
	ruleMu.Unlock()
	log.Printf("Loaded %d rule groups", len(groups))
}

func evaluateRules(protocol, method, path, query, userAgent, body string) (bool, string) {
	ruleMu.RLock()
	defer ruleMu.RUnlock()

	for _, g := range ruleCache {
		if g.Protocol != protocol && g.Protocol != "http" {
			continue
		}
		for _, r := range g.Rules {
			var target string
			switch r.Type {
			case "path":
				target = path
			case "query":
				target = query
			case "header", "user_agent":
				target = userAgent
			case "body":
				target = body
			case "command":
				target = body
			default:
				continue
			}
			if r.regex != nil && r.regex.MatchString(target) {
				return r.Action == "block", r.Type + ":" + r.Pattern
			}
		}
	}
	return false, ""
}

func startHTTPService(id uint, host string, port int, hp honeypot.Honeypot, cfg map[string]interface{}) {
	addr := fmt.Sprintf("%s:%d", host, port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(io.LimitReader(r.Body, 1<<20))
		}

		handled, attack := hp.HandleHTTP(w, r, body, cfg)

		if handled {
			attackType := ""
			attackDetail := ""
			isAttack := false
			if attack != nil {
				attackType = attack.Type
				isAttack = true
				if d, err := json.Marshal(attack.Detail); err == nil {
					attackDetail = string(d)
				}
			}

			blocked, ruleMatch := evaluateRules("http", r.Method, r.URL.Path, r.URL.RawQuery, r.UserAgent(), string(body))
			if ruleMatch != "" && !isAttack {
				isAttack = true
				attackType = "rule_match"
				attackDetail = ruleMatch
			}
			if blocked {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"forbidden"}`))
			}

			enqueueLog(LogEntry{
				NodeID:       getNodeID(),
				ServiceID:    id,
				ClientIP:     stripPort(r.RemoteAddr),
				Method:       r.Method,
				Path:         r.URL.Path,
				UserAgent:    r.UserAgent(),
				HoneypotType: hp.Type(),
				RequestBody:  truncate(string(body), 4096),
				IsAttack:     isAttack,
				AttackType:   attackType,
				AttackDetail: truncate(attackDetail, 2048),
				Protocol:     "http",
			})
		}
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	activeMu.Lock()
	activeSvcs[id] = &serviceRef{httpSrv: srv}
	activeMu.Unlock()
	log.Printf("[HTTP] Service %d (%s) listening on %s", id, hp.Name(), addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Service %d error: %v", id, err)
	}
}

func startTCPService(id uint, host string, port int, name string, protocol string, cfg map[string]interface{}) {
	var handler func(net.Conn) (string, bool, string)

	switch protocol {
	case "mysql":
		proto := server.NewMySQLProto(cfg)
		handler = func(conn net.Conn) (string, bool, string) {
			clientIP := stripPort(conn.RemoteAddr().String())
			detail, isAttack, _ := proto.Handle(conn)
			blocked, ruleMatch := evaluateRules("mysql", "CONNECT", "", "", "", detail)
			if ruleMatch != "" && !isAttack {
				isAttack = true
			}
			enqueueLog(LogEntry{
				NodeID:       getNodeID(),
				ServiceID:    id,
				ClientIP:     clientIP,
				Method:       "CONNECT",
				HoneypotType: 3,
				StatusCode:   1045,
				Protocol:     "mysql",
				IsAttack:     blocked || isAttack,
				AttackType:   ruleMatch,
				AttackDetail: truncate(detail, 2048),
			})
			if blocked {
				log.Printf("[TCP/mysql] Blocked connection from %s", clientIP)
			}
			return detail, blocked, ""
		}
	case "redis":
		proto := server.NewRedisProto(cfg)
		handler = func(conn net.Conn) (string, bool, string) {
			clientIP := stripPort(conn.RemoteAddr().String())
			detail, isAttack, _ := proto.Handle(conn)
			blocked, ruleMatch := evaluateRules("redis", "CONNECT", "", "", "", detail)
			if ruleMatch != "" && !isAttack {
				isAttack = true
			}
			enqueueLog(LogEntry{
				NodeID:       getNodeID(),
				ServiceID:    id,
				ClientIP:     clientIP,
				Method:       "CONNECT",
				HoneypotType: 4,
				StatusCode:   1,
				Protocol:     "redis",
				IsAttack:     blocked || isAttack,
				AttackType:   ruleMatch,
				AttackDetail: truncate(detail, 2048),
			})
			if blocked {
				log.Printf("[TCP/redis] Blocked connection from %s", clientIP)
			}
			return detail, blocked, ""
		}
	case "mongodb":
		proto := server.NewMongoDBProto(cfg)
		handler = func(conn net.Conn) (string, bool, string) {
			clientIP := stripPort(conn.RemoteAddr().String())
			detail, isAttack, _ := proto.Handle(conn)
			blocked, ruleMatch := evaluateRules("mongodb", "CONNECT", "", "", "", detail)
			if ruleMatch != "" && !isAttack {
				isAttack = true
			}
			enqueueLog(LogEntry{
				NodeID:       getNodeID(),
				ServiceID:    id,
				ClientIP:     clientIP,
				Method:       "CONNECT",
				HoneypotType: 5,
				StatusCode:   0,
				Protocol:     "mongodb",
				IsAttack:     blocked || isAttack,
				AttackType:   ruleMatch,
				AttackDetail: truncate(detail, 2048),
			})
			if blocked {
				log.Printf("[TCP/mongodb] Blocked connection from %s", clientIP)
			}
			return detail, blocked, ""
		}
	default:
		log.Printf("Unknown TCP protocol: %s", protocol)
		return
	}

	tcpSrv := server.NewTCPServerWithHandler(host, port, handler)
	activeMu.Lock()
	activeSvcs[id] = &serviceRef{tcpSrv: tcpSrv}
	activeMu.Unlock()
	log.Printf("[TCP/%s] Service %d (%s) listening on %s:%d", protocol, id, name, host, port)
	if err := tcpSrv.Start(); err != nil {
		log.Printf("TCP service %d error: %v", id, err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

var agentIDFile string

func init() {
	exe, err := os.Executable()
	if err != nil {
		agentIDFile = "data/agent.id"
		return
	}
	agentIDFile = filepath.Join(filepath.Dir(exe), "data", "agent.id")
}

func loadAgentID() string {
	data, err := os.ReadFile(agentIDFile)
	if err != nil || len(data) == 0 {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveAgentID(id string) {
	dir := filepath.Dir(agentIDFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("[FATAL] Cannot create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(agentIDFile, []byte(id), 0644); err != nil {
		log.Fatalf("[FATAL] Cannot write agent ID to %s: %v", agentIDFile, err)
	}
	log.Printf("[Config] Persisted agent_id=%s to %s", id, agentIDFile)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func stripPort(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func autoDetectIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("[Config] autoDetectIP failed: %v, defaulting to 127.0.0.1", err)
		return "127.0.0.1"
	}
	defer conn.Close()
	addr := conn.LocalAddr().String()
	ip := stripPort(addr)
	log.Printf("[Config] autoDetectIP=%s", ip)
	return ip
}

func collectResourceUsage() (float64, float64, float64) {
	cpuUsage := getCPUTotal()
	memUsage := getMemUsage()
	diskUsage := getDiskUsage()
	return cpuUsage, memUsage, diskUsage
}

func getCPUTotal() float64 {
	percentages, err := cpu.Percent(0, false)
	if err != nil || len(percentages) == 0 {
		return 0
	}
	return percentages[0]
}

func getMemUsage() float64 {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return 0
	}
	return stat.UsedPercent
}

func getDiskUsage() float64 {
	exe, err := os.Executable()
	if err != nil {
		return 0
	}
	stat, err := disk.Usage(filepath.Dir(exe))
	if err != nil {
		return 0
	}
	return stat.UsedPercent
}
