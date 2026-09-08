package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/faizahmd2/rate-guard/internal/config"
	"github.com/go-chi/chi/v5"
)

type RuleHandler struct {
	service *config.RuleService
}

func NewRuleHandler(
	service *config.RuleService,
) *RuleHandler {
	return &RuleHandler{
		service: service,
	}
}

type createRuleRequest struct {
	Service   string          `json:"service"`
	Resource  string          `json:"resource"`
	Algorithm string          `json:"algorithm"`
	Config    json.RawMessage `json:"config"`
}

type updateRuleRequest struct {
	Algorithm string          `json:"algorithm"`
	Config    json.RawMessage `json:"config"`
	Status    string          `json:"status"`
}

type ruleResponse struct {
	ID        string          `json:"id"`
	Service   string          `json:"service"`
	Resource  string          `json:"resource"`
	Algorithm string          `json:"algorithm"`
	Config    json.RawMessage `json:"config"`
	Status    string          `json:"status"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func (h *RuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRuleRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid JSON request.",
			},
		)

		return
	}

	rule, err := h.service.CreateRule(
		r.Context(),
		config.CreateRuleInput{
			Service:   request.Service,
			Resource:  request.Resource,
			Algorithm: request.Algorithm,
			Config:    request.Config,
		},
	)
	if err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		ruleResponse{
			ID:        rule.ID,
			Service:   rule.Service,
			Resource:  rule.Resource,
			Algorithm: rule.Algorithm,
			Config:    json.RawMessage(rule.Config),
			Status:    rule.Status,
			CreatedAt: rule.CreatedAt.Format(time.RFC3339),
			UpdatedAt: rule.UpdatedAt.Format(time.RFC3339),
		},
	)
}

func (h *RuleHandler) List(
	w http.ResponseWriter,
	r *http.Request,
) {
	rules, err := h.service.GetRules(
		r.Context(),
	)
	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve rules.",
			},
		)

		return
	}

	response := make(
		[]ruleResponse,
		0,
		len(rules),
	)

	for _, rule := range rules {
		response = append(
			response,
			toRuleResponse(rule),
		)
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

func (h *RuleHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	service := chi.URLParam(r, "service")
	resource := chi.URLParam(r, "resource")

	rule, err := h.service.GetRule(
		r.Context(),
		service,
		resource,
	)
	if err != nil {
		if errors.Is(
			err,
			config.ErrRuleNotFound,
		) {
			writeJSON(
				w,
				http.StatusNotFound,
				errorResponse{
					Code:    "RULE_NOT_FOUND",
					Message: "Rate limit rule not found.",
				},
			)

			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve rule.",
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		toRuleResponse(rule),
	)
}

func toRuleResponse(
	rule *config.RateLimitRule,
) ruleResponse {
	return ruleResponse{
		ID:        rule.ID,
		Service:   rule.Service,
		Resource:  rule.Resource,
		Algorithm: rule.Algorithm,
		Config:    json.RawMessage(rule.Config),
		Status:    rule.Status,
		CreatedAt: rule.CreatedAt.Format(time.RFC3339),
		UpdatedAt: rule.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *RuleHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	service := chi.URLParam(r, "service")
	resource := chi.URLParam(r, "resource")

	var request updateRuleRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid JSON request.",
			},
		)
		return
	}

	rule, err := h.service.UpdateRule(
		r.Context(),
		service,
		resource,
		config.UpdateRuleInput{
			Algorithm: request.Algorithm,
			Config:    request.Config,
			Status:    request.Status,
		},
	)

	if err != nil {
		if errors.Is(err, config.ErrRuleNotFound) {
			writeJSON(
				w,
				http.StatusNotFound,
				errorResponse{
					Code:    "RULE_NOT_FOUND",
					Message: "Rate limit rule not found.",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusBadRequest,
			errorResponse{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		toRuleResponse(rule),
	)
}

func (h *RuleHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	service := chi.URLParam(r, "service")
	resource := chi.URLParam(r, "resource")

	err := h.service.DeleteRule(
		r.Context(),
		service,
		resource,
	)

	if err != nil {
		if errors.Is(err, config.ErrRuleNotFound) {
			writeJSON(
				w,
				http.StatusNotFound,
				errorResponse{
					Code:    "RULE_NOT_FOUND",
					Message: "Rate limit rule not found.",
				},
			)
			return
		}

		writeJSON(
			w,
			http.StatusInternalServerError,
			errorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to delete rule.",
			},
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
