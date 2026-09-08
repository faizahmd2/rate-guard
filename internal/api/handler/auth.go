package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/faizahmd2/rate-limitter-service/internal/auth"
)

type AuthHandler struct {
	service *auth.Service
	cookies *auth.CookieManager
}

func NewAuthHandler(
	service *auth.Service,
	cookies *auth.CookieManager,
) *AuthHandler {
	return &AuthHandler{
		service: service,
		cookies: cookies,
	}
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Setup(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req setupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request",
			http.StatusBadRequest,
		)
		return
	}

	err := h.service.Setup(
		r.Context(),
		req.Username,
		req.Password,
	)

	if err != nil {
		if errors.Is(err, auth.ErrSetupCompleted) {
			http.Error(
				w,
				"setup already completed",
				http.StatusForbidden,
			)
			return
		}

		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	h.cookies.Set(
		w,
		strings.TrimSpace(req.Username),
	)

	w.WriteHeader(http.StatusCreated)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(
			w,
			"invalid request",
			http.StatusBadRequest,
		)
		return
	}

	err := h.service.Login(
		r.Context(),
		strings.TrimSpace(req.Username),
		req.Password,
	)

	if err != nil {
		http.Error(
			w,
			"invalid credentials",
			http.StatusUnauthorized,
		)
		return
	}

	h.cookies.Set(
		w,
		strings.TrimSpace(req.Username),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	admin, err := h.service.GetAdmin(r.Context())

	if err != nil {
		if errors.Is(err, auth.ErrAdminNotConfigured) {
			writeJSON(w, http.StatusOK, map[string]bool{
				"setup_required": true,
				"authenticated":  false,
			})
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, authenticated := h.cookies.Username(r)

	writeJSON(w, http.StatusOK, map[string]bool{
		"setup_required": false,
		"authenticated":  authenticated,
	})

	_ = admin
}

func (h *AuthHandler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.cookies.Clear(w)

	w.WriteHeader(http.StatusNoContent)
}
