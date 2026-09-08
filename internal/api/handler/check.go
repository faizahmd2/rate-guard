package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/faizahmd2/rate-limitter-service/internal/config"
	"github.com/faizahmd2/rate-limitter-service/internal/limiter"
)

type CheckHandler struct {
	limiterService *limiter.Service
}

func NewCheckHandler(
	limiterService *limiter.Service,
) *CheckHandler {
	return &CheckHandler{
		limiterService: limiterService,
	}
}

type checkRequest struct {
	Service  string `json:"service"`
	Resource string `json:"resource"`
	Key      string `json:"key"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *CheckHandler) Check(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req checkRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid JSON request.",
		)
		return
	}

	req.Service = strings.TrimSpace(req.Service)
	req.Resource = strings.TrimSpace(req.Resource)
	req.Key = strings.TrimSpace(req.Key)

	if req.Service == "" ||
		req.Resource == "" ||
		req.Key == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"service, resource and key are required.",
		)
		return
	}

	input := limiter.CheckInput{
		Service:  req.Service,
		Resource: req.Resource,
		Key:      req.Key,
	}

	decision, err := h.limiterService.Check(
		r.Context(),
		input,
	)
	if err != nil {
		handleCheckError(w, err)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		decision,
	)
}

func handleCheckError(
	w http.ResponseWriter,
	err error,
) {
	switch {
	case errors.Is(err, config.ErrRuleNotFound):
		writeError(
			w,
			http.StatusNotFound,
			"RULE_NOT_FOUND",
			"Rate limit rule not found.",
		)

	case errors.Is(err, limiter.ErrServiceUnavailable):
		writeError(
			w,
			http.StatusServiceUnavailable,
			"SERVICE_UNAVAILABLE",
			"RateGuard is temporarily unavailable.",
		)

	default:
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"An internal error occurred.",
		)
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	writeJSON(
		w,
		status,
		errorResponse{
			Code:    code,
			Message: message,
		},
	)
}

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
