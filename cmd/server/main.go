package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chenflux/honeywatch/internal/config"
	"github.com/chenflux/honeywatch/internal/database"
	"github.com/chenflux/honeywatch/internal/handler"
	"github.com/chenflux/honeywatch/internal/hub"
	"github.com/chenflux/honeywatch/internal/middleware"
	"github.com/chenflux/honeywatch/internal/service"
	"github.com/gorilla/mux"
)

const (
	Version   = "0.2.1"
	BuildDate = "20260513"
)

func findAgentFile(fn string) string {
	for _, dir := range []string{"dist", "bin"} {
		fp := filepath.Join(dir, fn)
		if _, err := os.Stat(fp); err == nil {
			return fp
		}
	}
	return ""
}

func main() {
	rootDir := findRootDir()
	os.Chdir(rootDir)

	fmt.Println("============================================================")
	fmt.Println("  HoneyWatch Management Center v2.0")
	fmt.Println("============================================================")
	fmt.Printf("Working directory: %s\n", rootDir)

	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("[1/4] Config loaded: manage=:%d, data=:%d, db=%s\n",
		cfg.Server.Port, cfg.Server.DataPort, cfg.Database.Type)

	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	fmt.Println("[2/4] Database initialized")
	_ = hub.GlobalHub

	authHandler := handler.NewAuthHandler()
	nodeHandler := handler.NewNodeHandler()
	serviceHandler := handler.NewServiceHandler()
	logHandler := handler.NewLogHandler()
	ruleHandler := handler.NewRuleHandler()
	configHandler := handler.NewConfigHandler()
	settingsHandler := handler.NewSettingsHandler()

	mgmtRouter := buildMgmtRouter(cfg, authHandler, nodeHandler, serviceHandler,
		logHandler, ruleHandler, configHandler, settingsHandler)
	dataRouter := buildDataRouter(nodeHandler, logHandler)

	fmt.Println("[3/4] Routes registered (manage + data plane)")

	startCleanupScheduler()

	mgmtAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	dataAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.DataPort)

	fmt.Printf("[4/4] Manage API: %s | Data API: %s\n", mgmtAddr, dataAddr)
	fmt.Println("============================================================")

	go func() {
		srv := &http.Server{
			Handler:      dataRouter,
			Addr:         dataAddr,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		log.Fatal(srv.ListenAndServe())
	}()

	srv := &http.Server{
		Handler:      mgmtRouter,
		Addr:         mgmtAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func buildMgmtRouter(
	cfg *config.AppConfig,
	authHandler *handler.AuthHandler,
	nodeHandler *handler.NodeHandler,
	serviceHandler *handler.ServiceHandler,
	logHandler *handler.LogHandler,
	ruleHandler *handler.RuleHandler,
	configHandler *handler.ConfigHandler,
	settingsHandler *handler.SettingsHandler,
) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RequestLogMiddleware)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()

	loginLimiter := middleware.RateLimitMiddleware(10, time.Minute)
	api.Handle("/auth/login", loginLimiter(http.HandlerFunc(authHandler.Login))).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/auth/profile", authHandler.GetProfile).Methods("GET")
	protected.HandleFunc("/auth/password", authHandler.ChangePassword).Methods("PUT")

	protected.HandleFunc("/nodes", nodeHandler.List).Methods("GET")
	protected.HandleFunc("/nodes/{id}", nodeHandler.Get).Methods("GET")
	protected.HandleFunc("/logs", logHandler.Query).Methods("GET")
	protected.HandleFunc("/logs/stats", logHandler.Stats).Methods("GET")
	protected.HandleFunc("/logs/trend", logHandler.Trend).Methods("GET")
	protected.HandleFunc("/logs/export", logHandler.Export).Methods("GET")

	handler.RegisterWSRoutes(r)

	api.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"version":     Version,
			"build_date":  BuildDate,
			"agent_token": cfg.Agent.Token,
		})
	}).Methods("GET")

	api.HandleFunc("/downloads/agents", func(w http.ResponseWriter, r *http.Request) {
		pattern := "dist/honeywatch-agent-" + Version + "-*"
		exes, _ := filepath.Glob(pattern)
		if len(exes) == 0 {
			exes, _ = filepath.Glob("bin/honeywatch-agent-" + Version + "-*")
		}
		type entry struct {
			Filename string `json:"filename"`
			OS       string `json:"os"`
			Arch     string `json:"arch"`
			Size     int64  `json:"size"`
		}
		var list []entry
		for _, f := range exes {
			info, _ := os.Stat(f)
			fn := filepath.Base(f)
			os, arch, _ := parseAgentFilename(fn)
			list = append(list, entry{
				Filename: fn,
				OS:       os,
				Arch:     arch,
				Size:     info.Size(),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"version": Version,
			"agents":  list,
		})
	}).Methods("GET")

	api.HandleFunc("/downloads/agent/{filename}", func(w http.ResponseWriter, r *http.Request) {
		fn := mux.Vars(r)["filename"]
		fp := findAgentFile(fn)
		if fp == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+fn)
		http.ServeFile(w, r, fp)
	}).Methods("GET")

	protected.HandleFunc("/rules/groups", ruleHandler.ListGroups).Methods("GET")
	protected.HandleFunc("/rules/groups/{id}", ruleHandler.GetGroup).Methods("GET")

	protected.HandleFunc("/services", serviceHandler.List).Methods("GET")
	protected.HandleFunc("/services/{id}", serviceHandler.Get).Methods("GET")
	protected.HandleFunc("/configs", configHandler.List).Methods("GET")
	protected.HandleFunc("/configs/{id}", configHandler.Get).Methods("GET")

	admin := protected.PathPrefix("").Subrouter()
	admin.Use(middleware.RequireAdmin)

	admin.HandleFunc("/auth/users", authHandler.ListUsers).Methods("GET")
	admin.HandleFunc("/settings/agent-token", settingsHandler.GetAgentToken).Methods("GET")
	admin.HandleFunc("/settings/agent-token", settingsHandler.RegenerateAgentToken).Methods("PUT")
	admin.HandleFunc("/nodes/{id}", nodeHandler.Delete).Methods("DELETE")
	admin.HandleFunc("/services", serviceHandler.Create).Methods("POST")
	admin.HandleFunc("/services/{id}", serviceHandler.Update).Methods("PUT")
	admin.HandleFunc("/services/{id}", serviceHandler.Delete).Methods("DELETE")
	admin.HandleFunc("/services/{id}/start", serviceHandler.Start).Methods("POST")
	admin.HandleFunc("/services/{id}/stop", serviceHandler.Stop).Methods("POST")
	admin.HandleFunc("/services/{id}/restart", serviceHandler.Restart).Methods("POST")
	admin.HandleFunc("/rules/groups", ruleHandler.CreateGroup).Methods("POST")
	admin.HandleFunc("/rules/groups/{id}", ruleHandler.UpdateGroup).Methods("PUT")
	admin.HandleFunc("/rules/groups/{id}", ruleHandler.DeleteGroup).Methods("DELETE")
	admin.HandleFunc("/rules/groups/{id}/rules", ruleHandler.CreateRule).Methods("POST")
	admin.HandleFunc("/rules/groups/{id}/rules/{rule_id}", ruleHandler.UpdateRule).Methods("PUT")
	admin.HandleFunc("/rules/groups/{id}/rules/{rule_id}", ruleHandler.DeleteRule).Methods("DELETE")
	admin.HandleFunc("/configs", configHandler.Create).Methods("POST")
	admin.HandleFunc("/configs/{id}", configHandler.Update).Methods("PUT")
	admin.HandleFunc("/configs/{id}", configHandler.Delete).Methods("DELETE")
	protected.HandleFunc("/configs/{id}/push", configHandler.Push).Methods("POST")

	r.PathPrefix("/").Handler(spaHandler{staticDir: "web/dist"})

	return r
}

func buildDataRouter(
	nodeHandler *handler.NodeHandler,
	logHandler *handler.LogHandler,
) *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.RequestLogMiddleware)

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","plane":"data","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	}).Methods("GET")

	data := r.PathPrefix("/data").Subrouter()
	data.Use(middleware.AgentAuthMiddleware)

	data.HandleFunc("/nodes/register", nodeHandler.Register).Methods("POST")
	data.HandleFunc("/nodes/{id}/heartbeat", nodeHandler.Heartbeat).Methods("POST")
	data.HandleFunc("/nodes/{id}/status", nodeHandler.UpdateStatus).Methods("PUT")

	data.HandleFunc("/logs", logHandler.Create).Methods("POST")
	data.HandleFunc("/logs/batch", logHandler.BatchCreate).Methods("POST")

	svcHandler := handler.NewServiceHandler()
	data.HandleFunc("/services", svcHandler.List).Methods("GET")

	ruleHandler2 := handler.NewRuleHandler()
	data.HandleFunc("/rules", ruleHandler2.ListEnabled).Methods("GET")

	configHandler2 := handler.NewConfigHandler()
	data.HandleFunc("/configs", configHandler2.List).Methods("GET")
	data.HandleFunc("/configs/node", configHandler2.ListByNodeID).Methods("GET")

	data.HandleFunc("/agents/download/{filename}", func(w http.ResponseWriter, r *http.Request) {
		fn := mux.Vars(r)["filename"]
		fp := findAgentFile(fn)
		if fp == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+fn)
		http.ServeFile(w, r, fp)
	}).Methods("GET")

	return r
}

func startCleanupScheduler() {
	go func() {
		authService := service.NewAuthService()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			authService.CleanupOfflineNodes()
			hub.BroadcastStats()
		}
	}()
}

type spaHandler struct {
	staticDir string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	fp := h.staticDir + path
	if info, err := os.Stat(fp); err == nil && !info.IsDir() {
		http.ServeFile(w, r, fp)
		return
	}
	http.ServeFile(w, r, h.staticDir+"/index.html")
}

func findRootDir() string {
	exePath, _ := os.Executable()
	candidates := []string{
		filepath.Dir(exePath),
		filepath.Join(filepath.Dir(exePath), ".."),
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	for _, dir := range candidates {
		abs, _ := filepath.Abs(dir)
		if _, err := os.Stat(filepath.Join(abs, "web", "dist", "index.html")); err == nil {
			return abs
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

func parseAgentFilename(fn string) (os string, arch string, ext string) {
	os = "unknown"
	arch = "unknown"
	if strings.Contains(fn, "-windows-") {
		os = "windows"
		arch = strings.TrimPrefix(fn, "honeywatch-agent-"+Version+"-windows-")
		arch = strings.TrimSuffix(arch, ".exe")
		ext = ".exe"
	} else if strings.Contains(fn, "-linux-") {
		os = "linux"
		arch = strings.TrimPrefix(fn, "honeywatch-agent-"+Version+"-linux-")
		ext = ""
	} else if strings.Contains(fn, "-darwin-") {
		os = "darwin"
		arch = strings.TrimPrefix(fn, "honeywatch-agent-"+Version+"-darwin-")
		ext = ""
	}
	return
}
