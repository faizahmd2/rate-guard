package web

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

func Handler() (http.Handler, error) {
	files, err := Files()
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(files))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(r.URL.Path, "/")

		// API routes are handled by the main router.
		if strings.HasPrefix(requestPath, "api/") ||
			strings.HasPrefix(requestPath, "v1/") ||
			requestPath == "health" {
			http.NotFound(w, r)
			return
		}

		// Serve an existing static file.
		if requestPath != "" {
			cleanPath := path.Clean(requestPath)

			if _, err := fs.Stat(files, cleanPath); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// React SPA fallback.
		//
		// Don't send /index.html through http.FileServer because
		// it has special redirect behavior for index.html.
		index, err := fs.ReadFile(files, "index.html")
		if err != nil {
			http.Error(
				w,
				"frontend index not found",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(index)
	}), nil
}
