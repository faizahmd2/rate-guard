package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

func APIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				writeUnauthorized(w)
				return
			}

			const prefix = "Bearer "

			if !strings.HasPrefix(authHeader, prefix) {
				writeUnauthorized(w)
				return
			}

			providedKey := strings.TrimSpace(
				strings.TrimPrefix(authHeader, prefix),
			)

			if providedKey == "" {
				writeUnauthorized(w)
				return
			}

			if subtle.ConstantTimeCompare(
				[]byte(providedKey),
				[]byte(expectedKey),
			) != 1 {
				writeUnauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusUnauthorized)

	_, _ = w.Write([]byte(`{
		"code": "UNAUTHORIZED",
		"message": "Invalid or missing API key."
	}`))
}
