package service

import (
	"fmt"
	"log"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
	"gorm.io/gorm"
)

type LogService struct{}

func NewLogService() *LogService {
	return &LogService{}
}

type CreateLogRequest struct {
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

type LogQuery struct {
	NodeID       string `json:"node_id"`
	ServiceID    uint   `json:"service_id"`
	ClientIP     string `json:"client_ip"`
	IsAttack     *bool  `json:"is_attack"`
	AttackType   string `json:"attack_type"`
	HoneypotType int    `json:"honeypot_type"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

func (s *LogService) Create(req CreateLogRequest) (*models.RequestLog, error) {
	geo := NewIPGeoService()
	geoInfo := geo.Lookup(req.ClientIP)

	logEntry := models.RequestLog{
		NodeID:       req.NodeID,
		ServiceID:    req.ServiceID,
		ClientIP:     req.ClientIP,
		Method:       req.Method,
		Path:         req.Path,
		UserAgent:    req.UserAgent,
		HoneypotType: req.HoneypotType,
		StatusCode:   req.StatusCode,
		RequestBody:  req.RequestBody,
		IsAttack:     req.IsAttack,
		AttackType:   req.AttackType,
		AttackDetail: req.AttackDetail,
		Protocol:     req.Protocol,
		GeoCountry:   geoInfo.Country,
		GeoCity:      geoInfo.City,
		GeoISP:       geoInfo.ISP,
	}
	if err := database.DB.Create(&logEntry).Error; err != nil {
		return nil, err
	}
	return &logEntry, nil
}

func (s *LogService) BatchCreate(logs []CreateLogRequest) error {
	if len(logs) == 0 {
		return nil
	}
	if len(logs) > 1000 {
		logs = logs[:1000]
	}
	log.Printf("[Log] BatchCreate count=%d", len(logs))
	return database.DB.Transaction(func(tx *gorm.DB) error {
		batch := make([]models.RequestLog, len(logs))
		for i, req := range logs {
			batch[i] = models.RequestLog{
				NodeID:       req.NodeID,
				ServiceID:    req.ServiceID,
				ClientIP:     req.ClientIP,
				Method:       req.Method,
				Path:         req.Path,
				UserAgent:    req.UserAgent,
				HoneypotType: req.HoneypotType,
				StatusCode:   req.StatusCode,
				RequestBody:  req.RequestBody,
				IsAttack:     req.IsAttack,
				AttackType:   req.AttackType,
				AttackDetail: req.AttackDetail,
				Protocol:     req.Protocol,
			}
		}
		return tx.CreateInBatches(batch, 100).Error
	})
}

func (s *LogService) Query(q LogQuery) ([]models.RequestLog, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}

	db := database.DB.Model(&models.RequestLog{})
	if q.NodeID != "" {
		db = db.Where("node_id = ?", q.NodeID)
	}
	if q.ServiceID > 0 {
		db = db.Where("service_id = ?", q.ServiceID)
	}
	if q.ClientIP != "" {
		db = db.Where("client_ip = ?", q.ClientIP)
	}
	if q.IsAttack != nil {
		db = db.Where("is_attack = ?", *q.IsAttack)
	}
	if q.AttackType != "" {
		db = db.Where("attack_type = ?", q.AttackType)
	}
	if q.HoneypotType > 0 {
		db = db.Where("honeypot_type = ?", q.HoneypotType)
	}
	if q.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, q.StartTime); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if q.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, q.EndTime); err == nil {
			db = db.Where("created_at <= ?", t)
		}
	}

	var total int64
	db.Count(&total)

	var logs []models.RequestLog
	err := db.Order("created_at DESC").
		Offset((q.Page - 1) * q.PageSize).
		Limit(q.PageSize).
		Find(&logs).Error
	return logs, total, err
}

type StatsSummary struct {
	TotalRequests    int64               `json:"total_requests"`
	HoneypotRequests int64               `json:"honeypot_requests"`
	AttackCount      int64               `json:"attack_count"`
	ByHoneypotType   map[int]int64       `json:"by_honeypot_type"`
	ByMethod         map[string]int64    `json:"by_method"`
	ByAttackType     map[string]int64    `json:"by_attack_type"`
	TopIPs           []IPStat            `json:"top_ips"`
	RecentAttacks    []models.RequestLog `json:"recent_attacks"`
}

type IPStat struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

func (s *LogService) GetSummary(hours int) (*StatsSummary, error) {
	if hours <= 0 {
		hours = 24
	}
	since := time.Now().Add(-time.Duration(hours) * time.Hour)
	summary := &StatsSummary{
		ByHoneypotType: make(map[int]int64),
		ByMethod:       make(map[string]int64),
		ByAttackType:   make(map[string]int64),
	}

	database.DB.Model(&models.RequestLog{}).Where("created_at >= ?", since).Count(&summary.TotalRequests)
	database.DB.Model(&models.RequestLog{}).Where("created_at >= ? AND honeypot_type > 0", since).Count(&summary.HoneypotRequests)
	database.DB.Model(&models.RequestLog{}).Where("created_at >= ? AND is_attack = true", since).Count(&summary.AttackCount)

	type groupResult struct {
		Key   string
		Count int64
	}
	var results []groupResult

	database.DB.Model(&models.RequestLog{}).
		Select("CAST(honeypot_type AS TEXT) as key, count(*) as count").
		Where("created_at >= ?", since).
		Group("honeypot_type").Scan(&results)
	for _, r := range results {
		var t int
		if _, err := fmt.Sscanf(r.Key, "%d", &t); err == nil {
			summary.ByHoneypotType[t] = r.Count
		}
	}

	results = nil
	database.DB.Model(&models.RequestLog{}).
		Select("method as key, count(*) as count").
		Where("created_at >= ?", since).
		Group("method").Scan(&results)
	for _, r := range results {
		summary.ByMethod[r.Key] = r.Count
	}

	results = nil
	database.DB.Model(&models.RequestLog{}).
		Select("attack_type as key, count(*) as count").
		Where("created_at >= ? AND is_attack = true AND attack_type != ''", since).
		Group("attack_type").Scan(&results)
	for _, r := range results {
		summary.ByAttackType[r.Key] = r.Count
	}

	results = nil
	database.DB.Model(&models.RequestLog{}).
		Select("client_ip as key, count(*) as count").
		Where("created_at >= ?", since).
		Group("client_ip").
		Order("count DESC").
		Limit(10).Scan(&results)
	for _, r := range results {
		summary.TopIPs = append(summary.TopIPs, IPStat{IP: r.Key, Count: r.Count})
	}

	database.DB.Where("created_at >= ? AND is_attack = true", since).
		Order("created_at DESC").Limit(10).Find(&summary.RecentAttacks)

	return summary, nil
}

func (s *LogService) GetTrend(honeypotType int, timeRange string) (map[string]interface{}, error) {
	now := time.Now()
	var labels []string
	var since time.Time
	var goFormat string
	var sqlFormat string
	isPg := database.IsPostgres()

	switch timeRange {
	case "1h":
		since = now.Add(-1 * time.Hour)
		goFormat = "15:04"
		if isPg {
			sqlFormat = "HH24:MI"
		} else {
			sqlFormat = "%H:%M"
		}
		for i := 11; i >= 0; i-- {
			t := now.Add(-time.Duration(i*5) * time.Minute)
			labels = append(labels, t.Format(goFormat))
		}
	case "24h":
		since = now.Add(-24 * time.Hour)
		goFormat = "15:04"
		if isPg {
			sqlFormat = "HH24:MI"
		} else {
			sqlFormat = "%H:%M"
		}
		for i := 11; i >= 0; i-- {
			t := now.Add(-time.Duration(i*2) * time.Hour)
			labels = append(labels, t.Format(goFormat))
		}
	case "30d":
		since = now.Add(-30 * 24 * time.Hour)
		goFormat = "01/02"
		if isPg {
			sqlFormat = "MM/DD"
		} else {
			sqlFormat = "%m/%d"
		}
		for i := 29; i >= 0; i-- {
			t := now.Add(-time.Duration(i) * 24 * time.Hour)
			labels = append(labels, t.Format(goFormat))
		}
	default:
		since = now.Add(-7 * 24 * time.Hour)
		goFormat = "01/02"
		if isPg {
			sqlFormat = "MM/DD"
		} else {
			sqlFormat = "%m/%d"
		}
		for i := 6; i >= 0; i-- {
			t := now.Add(-time.Duration(i) * 24 * time.Hour)
			labels = append(labels, t.Format(goFormat))
		}
	}

	type groupResult struct {
		Date  string
		Count int64
	}

	var dateExpr string
	if isPg {
		dateExpr = fmt.Sprintf("to_char(created_at, '%s')", sqlFormat)
	} else {
		dateExpr = fmt.Sprintf("strftime('%s', created_at)", sqlFormat)
	}

	queryDB := database.DB.Model(&models.RequestLog{}).Where("created_at >= ?", since)
	if honeypotType > 0 {
		queryDB = queryDB.Where("honeypot_type = ?", honeypotType)
	}

	var reqResults []groupResult
	queryDB.Select(dateExpr+" as date, count(*) as count").
		Group("date").Scan(&reqResults)

	attackDB := database.DB.Model(&models.RequestLog{}).Where("created_at >= ? AND is_attack = true", since)
	if honeypotType > 0 {
		attackDB = attackDB.Where("honeypot_type = ?", honeypotType)
	}
	var attackResults []groupResult
	attackDB.Select(dateExpr+" as date, count(*) as count").
		Group("date").Scan(&attackResults)

	reqMap := make(map[string]int64)
	for _, r := range reqResults {
		reqMap[r.Date] = r.Count
	}
	attackMap := make(map[string]int64)
	for _, r := range attackResults {
		attackMap[r.Date] = r.Count
	}

	data := make([]int64, len(labels))
	attacksData := make([]int64, len(labels))
	for i, label := range labels {
		data[i] = reqMap[label]
		attacksData[i] = attackMap[label]
	}

	return map[string]interface{}{
		"labels":       labels,
		"data":         data,
		"attacks_data": attacksData,
	}, nil
}

func (s *LogService) Export(q LogQuery) ([]models.RequestLog, error) {
	db := database.DB.Model(&models.RequestLog{})
	if q.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, q.StartTime); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if q.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, q.EndTime); err == nil {
			db = db.Where("created_at <= ?", t)
		}
	}
	var logs []models.RequestLog
	err := db.Order("created_at DESC").Limit(10000).Find(&logs).Error
	return logs, err
}
