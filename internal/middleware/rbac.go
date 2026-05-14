package middleware

import (
	"net/http"
	"strings"
)

func RBACMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, r := range allowedRoles {
		roleSet[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value("role").(string)
			if !roleSet[role] {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return RBACMiddleware(role)
}

func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole("admin")(next)
}

func getContextRole(r *http.Request) string {
	role, _ := r.Context().Value("role").(string)
	return strings.ToLower(role)
}
