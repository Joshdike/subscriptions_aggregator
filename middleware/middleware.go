package middleware

import (
	"net/http"
)

func AdminSecretMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("secret-key")
			if authHeader != secret {
				http.Error(w, "secret-key header is missing or invalid", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
