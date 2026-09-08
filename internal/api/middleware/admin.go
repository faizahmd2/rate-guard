package middleware

import (
	"net/http"

	"github.com/faizahmd2/rate-guard/internal/auth"
)

func AdminAuth(
	cookies *auth.CookieManager,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if _, ok := cookies.Username(r); !ok {
				http.Error(
					w,
					"unauthorized",
					http.StatusUnauthorized,
				)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
