package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/chenflux/pitcher/internal/config"
)

type SettingsHandler struct{}

func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

func (h *SettingsHandler) GetAgentToken(w http.ResponseWriter, r *http.Request) {
	token := config.GetAgentToken()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   token,
	})
}

func (h *SettingsHandler) RegenerateAgentToken(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 32)
	rand.Read(b)
	newToken := hex.EncodeToString(b)
	if err := config.SetAgentToken(newToken); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   newToken,
	})
}
