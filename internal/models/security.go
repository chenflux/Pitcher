package internal_models

import "time"

type SecurityListType string

const (
	ListTypeBlacklist SecurityListType = "blacklist"
	ListTypeWhitelist SecurityListType = "whitelist"
	ListTypeGraylist  SecurityListType = "graylist"
)

type SecurityEntry struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	IPAddress string         `gorm:"size:45;index;not null" json:"ip_address"`
	ListType SecurityListType `gorm:"size:20;not null" json:"list_type"`
	Reason   string         `gorm:"size:255" json:"reason"`
	ExpiresAt *time.Time    `gorm:"index" json:"expires_at"`
	CreatedBy string         `gorm:"size:50" json:"created_by"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

func (SecurityEntry) TableName() string {
	return "security_entries"
}

type CaptchaChallenge struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Code      string    `gorm:"size:20;not null" json:"code"`
	Expression string   `gorm:"size:100;not null" json:"expression"`
	ClientIP  string    `gorm:"size:45;index" json:"client_ip"`
	Used      bool      `gorm:"default:false" json:"used"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (CaptchaChallenge) TableName() string {
	return "captcha_challenges"
}