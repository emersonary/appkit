package email

import "go.uber.org/zap"

// Service groups protocol clients and Stalwart API access.
type Service struct {
	Client *Client
	API    *API
	logger *zap.Logger
}

// NewService wires SMTP/IMAP/POP client access and optional Stalwart API.
func NewService(cfg Config, provisioning ProvisioningConfig, logger *zap.Logger) *Service {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Service{
		Client: NewClient(cfg, logger),
		API:    NewAPI(provisioning),
		logger: logger,
	}
}
