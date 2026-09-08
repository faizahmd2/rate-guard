package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/faizahmd2/rate-guard/internal/auth"
)

func APIToken(tokens *auth.TokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")

			if header == "" {
				http.Error(w, "missing authorization", http.StatusUnauthorized)
				return
			}

			parts := strings.Fields(header)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			if !strings.HasPrefix(token, "rg_") {
				http.Error(w, "invalid authorization", http.StatusUnauthorized)
				return
			}

			hash := sha256.Sum256([]byte(token))
			tokenHash := hex.EncodeToString(hash[:])

			if !tokens.Has(tokenHash) {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
