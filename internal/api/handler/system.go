package handler

import (
	"net/http"

	"github.com/faizahmd2/rate-limitter-service/internal/auth"
	"github.com/faizahmd2/rate-limitter-service/internal/config"
	"github.com/faizahmd2/rate-limitter-service/internal/redis"
)

type SystemHandler struct {
	environment   string
	storageDriver string
	redisHost     string
	redisPort     string
	redisClient   *redis.Client
	ruleService   *config.RuleService
	authService   *auth.Service
}

func NewSystemHandler(
	environment string,
	storageDriver string,
	redisHost string,
	redisPort string,
	redisClient *redis.Client,
) *SystemHandler {
	return &SystemHandler{
		environment:   environment,
		storageDriver: storageDriver,
		redisHost:     redisHost,
		redisPort:     redisPort,
		redisClient:   redisClient,
	}
}

type systemResponse struct {
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Storage     struct {
		Driver string `json:"driver"`
		Status string `json:"status"`
	} `json:"storage"`
	Redis struct {
		Status string `json:"status"`
		Host   string `json:"host"`
	} `json:"redis"`
	Rules struct {
		Total  int `json:"total"`
		Active int `json:"active"`
	} `json:"rules"`
	Tokens struct {
		Active int `json:"active"`
	} `json:"tokens"`
}

func (h *SystemHandler) Get(w http.ResponseWriter, r *http.Request) {
	response := systemResponse{
		Version:     "0.1.0",
		Environment: h.environment,
	}

	response.Storage.Driver = h.storageDriver
	response.Storage.Status = "connected"

	redisStatus := "disconnected"

	if h.redisClient != nil {
		if err := h.redisClient.Ping(r.Context()); err == nil {
			redisStatus = "connected"
		}
	}

	response.Redis.Status = redisStatus
	response.Redis.Host = h.redisHost + ":" + h.redisPort

	writeJSON(w, http.StatusOK, response)
}
