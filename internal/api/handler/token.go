package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/faizahmd2/rate-guard/internal/auth"
)

type TokenHandler struct {
	service    *auth.Service
	tokenStore *auth.TokenStore
}

func NewTokenHandler(
	service *auth.Service,
	tokenStore *auth.TokenStore,
) *TokenHandler {
	return &TokenHandler{
		service:    service,
		tokenStore: tokenStore,
	}
}

type createTokenRequest struct {
	Name   string `json:"name"`
	Client string `json:"client"`
}

type tokenResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Client     string  `json:"client"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	LastUsedAt *string `json:"last_used_at,omitempty"`
}

type createTokenResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Client    string `json:"client"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

func (h *TokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, plaintext, err := h.service.CreateToken(
		r.Context(),
		req.Name,
		req.Client,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.tokenStore.Add(token.TokenHash)

	writeJSON(w, http.StatusCreated, createTokenResponse{
		ID:        token.ID,
		Name:      token.Name,
		Client:    token.Client,
		Token:     plaintext,
		CreatedAt: token.CreatedAt.UTC().Format(time.RFC3339),
	})
}

func (h *TokenHandler) List(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.service.ListTokens(r.Context())
	if err != nil {
		http.Error(w, "failed to load tokens", http.StatusInternalServerError)
		return
	}

	response := make([]tokenResponse, 0, len(tokens))

	for _, token := range tokens {
		var lastUsedAt *string

		if token.LastUsedAt != nil {
			value := token.LastUsedAt.UTC().Format(time.RFC3339)
			lastUsedAt = &value
		}

		response = append(response, tokenResponse{
			ID:         token.ID,
			Name:       token.Name,
			Client:     token.Client,
			Status:     token.Status,
			CreatedAt:  token.CreatedAt.UTC().Format(time.RFC3339),
			LastUsedAt: lastUsedAt,
		})
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *TokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "token id is required", http.StatusBadRequest)
		return
	}

	// Get the active tokens before revocation so we know
	// which hash needs to be removed from memory.
	tokens, err := h.service.ListTokens(r.Context())
	if err != nil {
		http.Error(w, "failed to load tokens", http.StatusInternalServerError)
		return
	}

	var tokenHash string

	for _, token := range tokens {
		if token.ID == id {
			tokenHash = token.TokenHash
			break
		}
	}

	if tokenHash == "" {
		http.Error(w, "token not found", http.StatusNotFound)
		return
	}

	if err := h.service.RevokeToken(r.Context(), id); err != nil {
		http.Error(w, "failed to revoke token", http.StatusInternalServerError)
		return
	}

	h.tokenStore.Remove(tokenHash)

	w.WriteHeader(http.StatusNoContent)
}
