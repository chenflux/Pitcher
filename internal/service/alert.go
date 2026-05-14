package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
)

type AlertService struct {
	channels map[uint]*AlertChannelImpl
	mu       sync.RWMutex
	rules    []models.AlertRule
}

type AlertChannelImpl struct {
	ID    uint
	Type  string
	Name  string
	Config map[string]string
	ThrottleMin int
	lastSent time.Time
}

func NewAlertService() *AlertService {
	svc := &AlertService{
		channels: make(map[uint]*AlertChannelImpl),
	}
	svc.loadChannels()
	return svc
}

func (s *AlertService) loadChannels() {
	var dbChannels []models.AlertChannel
	database.DB.Where("enabled = ?", true).Find(&dbChannels)

	s.mu.Lock()
	for _, ch := range dbChannels {
		cfg := make(map[string]string)
		if ch.Config != "" {
			json.Unmarshal([]byte(ch.Config), &cfg)
		}
		s.channels[ch.ID] = &AlertChannelImpl{
			ID:          ch.ID,
			Type:        ch.Type,
			Name:        ch.Name,
			Config:      cfg,
			ThrottleMin: ch.ThrottleMin,
			lastSent:    ch.LastSent,
		}
	}

	var dbRules []models.AlertRule
	database.DB.Where("enabled = ?", true).Find(&dbRules)
	s.rules = dbRules
	s.mu.Unlock()
}

func (s *AlertService) Reload() {
	s.loadChannels()
}

type AlertEvent struct {
	Type    string
	Message string
	Data    map[string]interface{}
}

func (s *AlertService) SendAttackAlert(attack map[string]interface{}) {
	event := AlertEvent{
		Type:    "attack",
		Message: fmt.Sprintf("Attack detected from %s: %s %s",
			attack["client_ip"], attack["method"], attack["path"]),
		Data:    attack,
	}
	s.processAlert(event)
}

func (s *AlertService) SendNewNodeAlert(node map[string]interface{}) {
	event := AlertEvent{
		Type:    "new_node",
		Message: fmt.Sprintf("New node registered: %s (%s)",
			node["hostname"], node["ip_address"]),
		Data: node,
	}
	s.processAlert(event)
}

func (s *AlertService) processAlert(event AlertEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rule := range s.rules {
		if s.shouldTrigger(rule, event) {
			s.sendToChannel(rule.ChannelID, event)
		}
	}
}

func (s *AlertService) shouldTrigger(rule models.AlertRule, event AlertEvent) bool {
	switch rule.Condition {
	case models.AlertConditionAttackCount:
		count, ok := event.Data["count"].(int)
		if !ok {
			return false
		}
		return count >= rule.Threshold
	case models.AlertConditionAttackType:
		attackType, ok := event.Data["attack_type"].(string)
		if !ok {
			return false
		}
		descLower := strings.ToLower(rule.Description)
		typeLower := strings.ToLower(attackType)
		return strings.Contains(descLower, typeLower)
	case models.AlertConditionNewNode:
		return event.Type == "new_node"
	}
	return false
}

func (s *AlertService) sendToChannel(channelID uint, event AlertEvent) {
	ch, ok := s.channels[channelID]
	if !ok {
		return
	}

	if time.Since(ch.lastSent) < time.Duration(ch.ThrottleMin)*time.Minute {
		return
	}

	switch ch.Type {
	case models.AlertChannelEmail:
		s.sendEmail(ch, event)
	case models.AlertChannelWebhook:
		s.sendWebhook(ch, event)
	case models.AlertChannelDingTalk:
		s.sendDingTalk(ch, event)
	}

	ch.lastSent = time.Now()
}

func (s *AlertService) sendEmail(ch *AlertChannelImpl, event AlertEvent) {
	to := ch.Config["to"]
	if to == "" {
		return
	}

	smtpHost := ch.Config["smtp_host"]
	smtpPort := ch.Config["smtp_port"]
	username := ch.Config["username"]
	password := ch.Config["password"]
	from := ch.Config["from"]
	if from == "" {
		from = username
	}

	if smtpHost == "" {
		smtpHost = "localhost"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}

	auth := smtp.PlainAuth("", username, password, smtpHost)
	toAddrs := strings.Split(to, ",")

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Pitcher Alert\r\n\r\n%s\r\n",
		from, to, event.Message)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	if ch.Config["use_tls"] == "true" {
		err := s.sendEmailTLS(addr, username, password, from, toAddrs, []byte(msg))
		if err != nil {
			log.Printf("[Alert] Email send failed: %v", err)
		}
	} else {
		err := smtp.SendMail(addr, auth, from, toAddrs, []byte(msg))
		if err != nil {
			log.Printf("[Alert] Email send failed: %v", err)
		}
	}
}

func (s *AlertService) sendEmailTLS(addr, username, password, from string, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: strings.Split(addr, ":")[0],
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, strings.Split(addr, ":")[0])
	if err != nil {
		return err
	}
	defer client.Close()

	if username != "" {
		auth := smtp.PlainAuth("", username, password, strings.Split(addr, ":")[0])
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(from); err != nil {
		return err
	}
	for _, t := range to {
		if err := client.Rcpt(t); err != nil {
			return err
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return client.Quit()
}

func (s *AlertService) sendWebhook(ch *AlertChannelImpl, event AlertEvent) {
	url := ch.Config["url"]
	if url == "" {
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type":      event.Type,
		"message":   event.Message,
		"data":      event.Data,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if token := ch.Config["auth_header"]; token != "" {
		req.Header.Set("Authorization", token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Alert] Webhook failed: %v", err)
		return
	}
	resp.Body.Close()
}

func (s *AlertService) sendDingTalk(ch *AlertChannelImpl, event AlertEvent) {
	webhook := ch.Config["webhook"]
	if webhook == "" {
		return
	}

	secret := ch.Config["secret"]

	var msg map[string]interface{}
	if secret != "" {
		timestamp := time.Now().UnixMilli()
		sign := s.dingTalkSign(secret, timestamp)
		webhook = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhook, timestamp, sign)
		msg = map[string]interface{}{
			"msgtype": "markdown",
			"markdown": map[string]string{
				"title": "Pitcher Alert",
				"text":  fmt.Sprintf("### Pitcher Alert\n\n%s\n\n**Data:**\n```json\n%s\n```",
					event.Message, s.json(event.Data)),
			},
		}
	} else {
		msg = map[string]interface{}{
			"msgtype": "text",
			"text": map[string]string{
				"content": fmt.Sprintf("Pitcher Alert: %s", event.Message),
			},
		}
	}

	payload, _ := json.Marshal(msg)
	req, _ := http.NewRequest("POST", webhook, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Alert] DingTalk failed: %v", err)
		return
	}
	resp.Body.Close()
}

func (s *AlertService) dingTalkSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	return url.QueryEscape(hmacSHA256(stringToSign, secret))
}

func (s *AlertService) json(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func hmacSHA256(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}