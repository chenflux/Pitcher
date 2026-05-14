package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/chenflux/honeywatch/internal/config"
)

func AgentAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Agent-Token")
		if subtle.ConstantTimeCompare([]byte(token), []byte(config.GlobalConfig.Agent.Token)) != 1 {
			http.Error(w, `{"error":"invalid agent token"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
