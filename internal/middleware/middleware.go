package middleware

import (
	"fmt"
	"net/http"

	"github.com/Joshdike/subscriptions_aggregator/internal/pkg/errors"
	"github.com/Joshdike/subscriptions_aggregator/internal/utils"
)

// AdminSecretMiddleware is a middleware that checks if the incoming request has a valid secret-key header set to the value provided when creating the middleware.
// If the header is missing or invalid, it will return a 401 Unauthorized response.
func AdminSecretMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the value of the secret-key header
			authHeader := r.Header.Get("secret-key")

			// If the header is not set or is invalid, return a 401 Unauthorized response
			if authHeader != secret {
				err := fmt.Errorf("%w: secret-key header is missing or invalid", errors.ErrUnauthorized)
				utils.WriteError(w, err)
				return
			}

			// If the header is valid, call the next handler in the chain
			next.ServeHTTP(w, r)
		})
	}
}
