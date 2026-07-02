package ai

import (
	"context"
	"strings"
)

// Service routes AI capability operations to configured providers.
type Service struct {
	cfg         AIConfig
	router      *Router
	translation *Translation
	chat        *Chat
}

func NewService(cfg AIConfig) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	providers := make(map[string]Provider, len(cfg.Providers))
	for id, providerCfg := range cfg.Providers {
		if !providerCfg.isEnabled() {
			continue
		}
		provider, err := buildProvider(id, providerCfg)
		if err != nil {
			return nil, err
		}
		providers[id] = provider
	}

	router := newRouter(cfg.routeTable, providers)
	return &Service{
		cfg:         cfg,
		router:      router,
		translation: newTranslation(router),
		chat:        newChat(router),
	}, nil
}

func buildProvider(providerID string, cfg ProviderConfig) (Provider, error) {
	switch cfg.normalizedDriver(providerID) {
	case defaultOpenAIDriver:
		return newOpenAIProvider(openAIProviderConfig{
			ProviderID: providerID,
			APIKey:     cfg.resolvedAPIKey(),
			BaseURL:    strings.TrimRight(cfg.BaseURL, "/"),
			Model:      cfg.DefaultModel,
			Timeout:    cfg.Timeout,
		})
	case defaultLocalDriver:
		return newLocalProvider(providerID), nil
	default:
		return nil, invalidConfigf("providers.%s.driver", providerID, "unsupported driver %q", cfg.driver(providerID))
	}
}

func (s *Service) Config() AIConfig {
	return s.cfg
}

func (s *Service) Translation() *Translation {
	return s.translation
}

func (s *Service) Chat() *Chat {
	return s.chat
}

// RouteSummary returns capability.operation to provider id mappings.
func (s *Service) RouteSummary() map[string]string {
	return s.router.Summary()
}

// Supports reports whether an operation is routed and supported.
func (s *Service) Supports(capability Capability, operation Operation) bool {
	_, err := s.router.ProviderFor(capability, operation)
	return err == nil
}

// ProviderFor reports which provider handles a capability operation.
func (s *Service) ProviderFor(capability Capability, operation Operation) (string, error) {
	return s.router.ProviderIDFor(capability, operation)
}

// Translate is a convenience wrapper for Translation().Translate.
func (s *Service) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	return s.translation.Translate(ctx, req)
}
