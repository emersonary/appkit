package config

import (
	"encoding/json"
	"net/http"
)

// PayloadFunc returns the JSON body for a config endpoint.
type PayloadFunc func() any

// Handler serves application config as JSON.
type Handler struct {
	payload PayloadFunc
}

// NewHandler builds a config JSON handler. Product config structs should use json:"-" on secret fields.
func NewHandler(payload PayloadFunc) *Handler {
	return &Handler{payload: payload}
}

// RegisterRoutes mounts the handler on route (e.g. "GET /infra/config").
func (h *Handler) RegisterRoutes(mux *http.ServeMux, route string) {
	mux.HandleFunc(route, h.Serve)
}

// Serve writes the configured payload as JSON.
func (h *Handler) Serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(h.payload()); err != nil {
		http.Error(w, "encode config", http.StatusInternalServerError)
	}
}
