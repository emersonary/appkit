// Package transport wires health HTTP around health.Service.
//
//	svc := health.NewService(pool, nats, logger)
//	transport.New(svc, &transport.Mount{
//	    HTTPMux: mux,
//	    AppName: cfg.App.Name,
//	    Logger:  logger,
//	})
package transport

import (
	"net/http"

	"github.com/emersonary/appkit/health"
	"go.uber.org/zap"
)

// Block wires health transport around a Service.
type Block struct {
	Service *health.Service
	appName string
	logger  *zap.Logger
}

// Mount configures where health transport is registered.
type Mount struct {
	HTTPMux *http.ServeMux
	AppName string
	Logger  *zap.Logger
}

// New builds the health transport block and registers endpoints when mount is non-nil.
func New(svc *health.Service, mount *Mount) *Block {
	logger := zap.NewNop()
	appName := ""
	if mount != nil {
		if mount.Logger != nil {
			logger = mount.Logger.Named("transport")
		}
		appName = mount.AppName
	}

	b := &Block{
		Service: svc,
		appName: appName,
		logger:  logger,
	}
	if mount != nil {
		b.mount(mount)
	}
	return b
}

func (b *Block) mount(m *Mount) {
	if m.HTTPMux == nil || b.Service == nil {
		return
	}
	m.HTTPMux.HandleFunc("GET /health", b.handleHealth)
	m.HTTPMux.HandleFunc("GET /infra/health", b.handleInfraHealth)
}

func (b *Block) handleHealth(w http.ResponseWriter, r *http.Request) {
	b.logger.Debug("health check requested", zap.String("path", r.URL.Path))
	_, _ = w.Write([]byte(b.appName + " is running"))
}

func (b *Block) handleInfraHealth(w http.ResponseWriter, r *http.Request) {
	b.logger.Debug("infra health check requested", zap.String("path", r.URL.Path))

	if err := b.Service.IsInfraHealthy(); err != nil {
		b.logger.Warn("infra health check failed", zap.Error(err))
		http.Error(w, "infra error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	b.logger.Debug("infra health check succeeded")
	_, _ = w.Write([]byte("postgres and nats are connected"))
}
