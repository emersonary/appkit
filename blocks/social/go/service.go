package social

import (
	"context"
	"net/http"
)

// Service routes social publishing to configured platform clients.
type Service struct {
	cfg       SocialConfig
	templates *TemplateRenderer
	clients   map[PlatformID]PlatformClient
}

// NewService builds a social service from resolved config.
func NewService(cfg SocialConfig, templates *TemplateRenderer, httpClient *http.Client) (*Service, error) {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	clients, err := buildAllPlatformClients(cfg, templates, httpClient)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg:       cfg,
		templates: templates,
		clients:   clients,
	}, nil
}

func (s *Service) Config() SocialConfig {
	return s.cfg
}

// Client returns a wired platform client.
func (s *Service) Client(platformID PlatformID) (PlatformClient, error) {
	client, ok := s.clients[platformID]
	if !ok {
		return nil, ErrPlatformNotFound.With("platform", string(platformID))
	}
	return client, nil
}

// EnabledPlatforms returns platform ids with active clients.
func (s *Service) EnabledPlatforms() []PlatformID {
	out := make([]PlatformID, 0, len(s.clients))
	for id := range s.clients {
		out = append(out, id)
	}
	return out
}

// FormatPost renders the platform template without publishing.
func (s *Service) FormatPost(ctx context.Context, platformID PlatformID, input PostInput) (FormattedPost, error) {
	client, err := s.Client(platformID)
	if err != nil {
		return FormattedPost{}, err
	}
	return client.FormatPost(ctx, input)
}

// PublishToPlatforms formats and publishes to each requested platform.
func (s *Service) PublishToPlatforms(ctx context.Context, req PublishRequest) []PublishResult {
	platforms := req.PlatformIDs
	if len(platforms) == 0 {
		platforms = s.EnabledPlatforms()
	}

	results := make([]PublishResult, 0, len(platforms))
	for _, platformID := range platforms {
		client, err := s.Client(platformID)
		if err != nil {
			results = append(results, PublishResult{PlatformID: platformID, Err: err})
			continue
		}

		formatted, err := client.FormatPost(ctx, req.Input)
		if err != nil {
			results = append(results, PublishResult{PlatformID: platformID, Err: err})
			continue
		}

		createReq := CreatePostRequest{
			Input:        req.Input,
			Formatted:    formatted,
			DispatchMode: req.Dispatch,
		}
		result, err := client.CreatePost(ctx, createReq)
		results = append(results, PublishResult{
			PlatformID: platformID,
			Result:     result,
			Err:        err,
		})
	}
	return results
}
