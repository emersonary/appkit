package email

import "go.uber.org/zap"

// Handler groups protocol clients and Stalwart API access.
type Handler struct {
	Client *Client
	API    *API
}

// NewHandler wires SMTP/IMAP/POP client access and optional Stalwart API.
func NewHandler(cfg Config, provisioning ProvisioningConfig, logger *zap.Logger) *Handler {
	return &Handler{
		Client: NewClient(cfg, logger),
		API:    NewAPI(provisioning),
	}
}
