package health

import (
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
	appName string
}

func NewHandler(service *Service, logger *zap.Logger, appName string) *Handler {
	return &Handler{
		service: service,
		logger:  logger.Named("handler"),
		appName: appName,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /infra/health", h.InfraHealth)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("health check requested", zap.String("path", r.URL.Path))
	_, _ = w.Write([]byte(h.appName + " is running"))
}

func (h *Handler) InfraHealth(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("infra health check requested", zap.String("path", r.URL.Path))

	if err := h.service.IsInfraHealthy(); err != nil {
		h.logger.Warn("infra health check failed", zap.Error(err))
		http.Error(w, "infra error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Debug("infra health check succeeded")
	_, _ = w.Write([]byte("postgres and nats are connected"))
}
