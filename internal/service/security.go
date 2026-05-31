package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
)

type SecurityService struct{}

func NewSecurityService() *SecurityService {
	return &SecurityService{}
}

type LoginAttempt struct {
	Count         int
	FirstAttempt  time.Time
	LastAttempt   time.Time
	CaptchaRequired bool
	LockedUntil   time.Time
}

var loginAttempts = make(map[string]*LoginAttempt)
var loginMu sync.RWMutex

const (
	MaxLoginAttempts     = 5
	WindowMinutes        = 15
	LockDurationMinutes  = 15
	GraylistTTLMinutes   = 30
	CaptchaThreshold     = 3
)

type CaptchaResult struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Expression string `json:"expression"`
}

func (s *SecurityService) GenerateCaptcha(clientIP string) (*CaptchaResult, error) {
	a := rand.Intn(10)
	b := rand.Intn(10)
	op := "+"
	if rand.Float32() > 0.5 {
		if a < b {
			a, b = b, a
		}
		op = "-"
	}
	expr := fmt.Sprintf("%d %s %d", a, op, b)
	code := fmt.Sprintf("%d", evaluateExpr(a, op, b))

	expiresAt := time.Now().Add(5 * time.Minute)
	c := models.CaptchaChallenge{
		Code:       code,
		Expression: expr,
		ClientIP:   clientIP,
		ExpiresAt:  expiresAt,
	}

	if err := database.DB.Create(&c).Error; err != nil {
		return nil, err
	}

	return &CaptchaResult{
		ID:    fmt.Sprintf("%d", c.ID),
		Code:  expr,
	}, nil
}

func evaluateExpr(a int, op string, b int) int {
	if op == "+" {
		return a + b
	}
	return a - b
}

func (s *SecurityService) VerifyCaptcha(id, code, clientIP string) bool {
	var captcha models.CaptchaChallenge
	var cid uint
	if _, err := fmt.Sscanf(id, "%d", &cid); err != nil {
		return false
	}

	if err := database.DB.Where("id = ? AND client_ip = ? AND used = false AND expires_at > ?",
		cid, clientIP, time.Now()).First(&captcha).Error; err != nil {
		return false
	}

	if captcha.Code != code {
		return false
	}

	captcha.Used = true
	database.DB.Save(&captcha)
	return true
}

func (s *SecurityService) CheckIPList(clientIP string) (listType models.SecurityListType, reason string, ttlSeconds int) {
	var entry models.SecurityEntry

	if err := database.DB.Where("ip_address = ? AND (expires_at IS NULL OR expires_at > ?)",
		clientIP, time.Now()).First(&entry).Error; err != nil {
		return "", "", 0
	}

	ttl := 0
	if entry.ExpiresAt != nil {
		ttl = int(time.Until(*entry.ExpiresAt).Seconds())
		if ttl < 0 {
			database.DB.Delete(&entry)
			return "", "", 0
		}
	}
	return entry.ListType, entry.Reason, ttl
}

func (s *SecurityService) AddToList(clientIP string, listType models.SecurityListType, reason, createdBy string, durationMinutes int) error {
	entry := models.SecurityEntry{
		IPAddress: clientIP,
		ListType:  listType,
		Reason:    reason,
		CreatedBy: createdBy,
	}

	if durationMinutes > 0 {
		t := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
		entry.ExpiresAt = &t
	}

	return database.DB.Create(&entry).Error
}

func (s *SecurityService) RemoveFromList(clientIP string, listType models.SecurityListType) error {
	return database.DB.Where("ip_address = ? AND list_type = ?", clientIP, listType).
		Delete(&models.SecurityEntry{}).Error
}

func (s *SecurityService) ListEntries(listType string, page, pageSize int) ([]models.SecurityEntry, int64, error) {
	var entries []models.SecurityEntry
	var total int64

	query := database.DB.Model(&models.SecurityEntry{})
	if listType != "" {
		query = query.Where("list_type = ?", listType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (s *SecurityService) RecordLoginAttempt(clientIP string, success bool) (captchaRequired bool, remainingAttempts int, lockedUntil int64, graylistTTL int) {
	loginMu.Lock()
	defer loginMu.Unlock()

	now := time.Now()
	attempt, exists := loginAttempts[clientIP]

	if !exists || now.Sub(attempt.FirstAttempt) > WindowMinutes*time.Minute {
		attempt = &LoginAttempt{
			FirstAttempt: now,
			LastAttempt:  now,
			Count:        0,
		}
		loginAttempts[clientIP] = attempt
	}

	if success {
		delete(loginAttempts, clientIP)
		return false, MaxLoginAttempts, 0, 0
	}

	attempt.Count++
	attempt.LastAttempt = now

	remainingAttempts = MaxLoginAttempts - attempt.Count
	if remainingAttempts < 0 {
		remainingAttempts = 0
	}

	if attempt.Count >= MaxLoginAttempts {
		expiresAt := now.Add(LockDurationMinutes * time.Minute)
		attempt.LockedUntil = expiresAt
		lockedUntil = expiresAt.Unix() * 1000

		go s.AddToList(clientIP, models.ListTypeGraylist,
			fmt.Sprintf("login_bruteforce_%d_attempts", attempt.Count),
			"system", GraylistTTLMinutes)

		return true, 0, lockedUntil, GraylistTTLMinutes * 60
	}

	if attempt.Count >= CaptchaThreshold {
		attempt.CaptchaRequired = true
		captchaRequired = true
	}

	return captchaRequired, remainingAttempts, 0, 0
}

func (s *SecurityService) IsCaptchaRequired(clientIP string) bool {
	loginMu.RLock()
	defer loginMu.RUnlock()

	if attempt, exists := loginAttempts[clientIP]; exists {
		return attempt.CaptchaRequired && time.Since(attempt.FirstAttempt) <= WindowMinutes*time.Minute
	}
	return false
}

func (s *SecurityService) GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return extractIP(r.RemoteAddr)
}

func extractIP(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		host := addr[:idx]
		if strings.HasPrefix(host, "[") {
			idx = strings.LastIndex(host, "]")
			if idx > 0 {
				return host[1:idx]
			}
		}
		return host
	}
	return addr
}

func (s *SecurityService) CleanupExpiredAttempts() {
	loginMu.Lock()
	defer loginMu.Unlock()

	now := time.Now()
	for ip, attempt := range loginAttempts {
		if now.Sub(attempt.FirstAttempt) > WindowMinutes*time.Minute {
			delete(loginAttempts, ip)
		}
	}
}

func (s *SecurityService) CleanupExpiredEntries() {
	database.DB.Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Delete(&models.SecurityEntry{})
}

func (s *SecurityService) IsWhitelisted(clientIP string) bool {
	var count int64
	database.DB.Model(&models.SecurityEntry{}).
		Where("ip_address = ? AND list_type = ? AND (expires_at IS NULL OR expires_at > ?)",
			clientIP, models.ListTypeWhitelist, time.Now()).
		Count(&count)
	return count > 0
}

func (s *SecurityService) IsBlacklisted(clientIP string) bool {
	var count int64
	database.DB.Model(&models.SecurityEntry{}).
		Where("ip_address = ? AND list_type = ? AND (expires_at IS NULL OR expires_at > ?)",
			clientIP, models.ListTypeBlacklist, time.Now()).
		Count(&count)
	return count > 0
}