package handler

import (
	"encoding/json"
	"net/http"

	"github.com/chenflux/pitcher/internal/service"
)

type AuthHandler struct {
	authService     *service.AuthService
	securityService *service.SecurityService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService:     service.NewAuthService(),
		securityService: service.NewSecurityService(),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password required")
		return
	}

	clientIP := h.securityService.GetClientIP(r)

	if listType, _, ttl := h.securityService.CheckIPList(clientIP); listType != "" {
		if listType == "blacklist" {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
		if listType == "graylist" {
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"success":      false,
				"error":        "too many requests",
				"graylist_ttl": ttl,
			})
			return
		}
	}

	resp, err := h.authService.Login(req)
	if err != nil {
		captchaRequired, remainingAttempts, lockedUntil, graylistTTL := h.securityService.RecordLoginAttempt(clientIP, false)

		result := map[string]interface{}{
			"success":           false,
			"error":             err.Error(),
			"captcha_required":  captchaRequired,
			"remaining_attempts": remainingAttempts,
		}

		if lockedUntil > 0 {
			result["locked_until"] = lockedUntil
		}
		if graylistTTL > 0 {
			result["graylist_ttl"] = graylistTTL
		}

		if lockedUntil > 0 || remainingAttempts <= 0 {
			writeJSON(w, http.StatusTooManyRequests, result)
			return
		}

		writeJSON(w, http.StatusUnauthorized, result)
		return
	}

	h.securityService.RecordLoginAttempt(clientIP, true)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   resp.Token,
		"user":    resp.UserInfo,
	})
}

func (h *AuthHandler) GetCaptcha(w http.ResponseWriter, r *http.Request) {
	clientIP := h.securityService.GetClientIP(r)

	captcha, err := h.securityService.GenerateCaptcha(clientIP)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate captcha")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"id":      captcha.ID,
		"code":    captcha.Code,
	})
}

func getContextUser(r *http.Request) (uint, string, string, bool) {
	userID, ok1 := r.Context().Value("user_id").(uint)
	username, ok2 := r.Context().Value("username").(string)
	role, ok3 := r.Context().Value("role").(string)
	return userID, username, role, ok1 && ok2 && ok3
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, _, ok := getContextUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"user":    user,
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _, _, ok := getContextUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "password changed",
	})
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.authService.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"users":   users,
	})
}