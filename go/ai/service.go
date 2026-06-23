package ai

import (
	"context"
	"strings"
)

// Service routes AI operations to configured providers.
type Service struct {
	cfg     AIConfig
	clients map[string]Client
}

func NewService(cfg AIConfig) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	clients := make(map[string]Client, len(cfg.Providers))
	for id, providerCfg := range cfg.Providers {
		if !providerCfg.isEnabled() {
			continue
		}

		client, err := buildProviderClient(id, providerCfg)
		if err != nil {
			return nil, err
		}
		clients[id] = client
	}

	return &Service{
		cfg:     cfg,
		clients: clients,
	}, nil
}

func buildProviderClient(providerID string, cfg ProviderConfig) (Client, error) {
	switch cfg.driver(providerID) {
	case defaultOpenAIDriver:
		return newOpenAIClient(openAIClientConfig{
			ProviderID: providerID,
			APIKey:     cfg.resolvedAPIKey(),
			BaseURL:    strings.TrimRight(cfg.BaseURL, "/"),
			Model:      cfg.DefaultModel,
			Timeout:    cfg.Timeout,
		})
	default:
		return nil, invalidConfigf("providers.%s.driver", providerID, "unsupported driver %q", cfg.driver(providerID))
	}
}

func (s *Service) Config() AIConfig {
	return s.cfg
}

// ClientFor returns the provider client for a service type.
func (s *Service) ClientFor(serviceType ServiceType) (Client, error) {
	providerID, ok := s.cfg.route(serviceType)
	if !ok {
		return nil, ErrServiceNotRouted.With("service_type", string(serviceType))
	}

	client, ok := s.clients[providerID]
	if !ok {
		return nil, ErrProviderNotFound.With("provider", providerID)
	}
	return client, nil
}

// Translate uses the routed translation provider.
func (s *Service) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	client, err := s.ClientFor(ServiceTypeTranslation)
	if err != nil {
		return TranslateResponse{}, err
	}
	return client.Translate(ctx, req)
}

// ProviderFor reports which provider handles a service type.
func (s *Service) ProviderFor(serviceType ServiceType) (string, error) {
	providerID, ok := s.cfg.route(serviceType)
	if !ok {
		return "", ErrServiceNotRouted.With("service_type", string(serviceType))
	}
	return providerID, nil
}

// Supports reports whether a service type is routed and its provider is wired.
func (s *Service) Supports(serviceType ServiceType) bool {
	providerID, err := s.ProviderFor(serviceType)
	if err != nil {
		return false
	}
	_, ok := s.clients[providerID]
	return ok
}

// RouteSummary returns service type to provider id mappings.
func (s *Service) RouteSummary() map[ServiceType]string {
	out := make(map[ServiceType]string, len(s.cfg.Routes))
	for rawType, providerID := range s.cfg.Routes {
		serviceType, err := ParseServiceType(rawType)
		if err != nil {
			continue
		}
		out[serviceType] = strings.TrimSpace(providerID)
	}
	return out
}
