package service

import (
	"errors"
	"time"

	"github.com/chenflux/pitcher/internal/database"
	models "github.com/chenflux/pitcher/internal/models"
	"github.com/chenflux/pitcher/internal/middleware"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Captcha   string `json:"captcha"`
	CaptchaID string `json:"captcha_id"`
}

type LoginResponse struct {
	Token    string      `json:"token"`
	UserInfo models.User `json:"user"`
}

func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:    token,
		UserInfo: user,
	}, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	if len(newPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range newPassword {
		switch {
		case c >= 'A' && c <= 'Z': hasUpper = true
		case c >= 'a' && c <= 'z': hasLower = true
		case c >= '0' && c <= '9': hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain uppercase, lowercase and digits")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return database.DB.Model(&user).Update("password_hash", string(hash)).Error
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) CreateUser(username, password, role string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) ListUsers() ([]models.User, error) {
	var users []models.User
	err := database.DB.Find(&users).Error
	return users, err
}

func (s *AuthService) CleanupOfflineNodes() {
	timeout := time.Duration(60) * time.Second
	threshold := time.Now().Add(-timeout)
	database.DB.Model(&models.Node{}).
		Where("status = ? AND last_heartbeat < ?", "online", threshold).
		Update("status", "offline")

	ttl := 24 * time.Hour
	ttlThreshold := time.Now().Add(-ttl)
	database.DB.Where("status = ? AND updated_at < ?", "offline", ttlThreshold).
		Delete(&models.Node{})
}
