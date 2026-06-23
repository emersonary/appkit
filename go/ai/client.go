package ai

import "context"

// Client performs AI operations for a single provider.
type Client interface {
	Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error)
}
