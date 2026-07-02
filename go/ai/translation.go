package ai

import "context"

// Translation routes translation operations to configured providers.
type Translation struct {
	router *Router
}

func newTranslation(router *Router) *Translation {
	return &Translation{router: router}
}

func (t *Translation) Detect(ctx context.Context, text string) (DetectResponse, error) {
	provider, err := t.router.ProviderFor(CapabilityTranslation, OpDetect)
	if err != nil {
		return DetectResponse{}, err
	}
	backend := provider.Translation()
	if backend == nil {
		return DetectResponse{}, ErrOperationNotSupported.With(
			"capability", string(CapabilityTranslation),
			"operation", string(OpDetect),
			"provider", provider.ID(),
		)
	}
	return backend.Detect(ctx, text)
}

func (t *Translation) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	provider, err := t.router.ProviderFor(CapabilityTranslation, OpTranslate)
	if err != nil {
		return TranslateResponse{}, err
	}
	backend := provider.Translation()
	if backend == nil {
		return TranslateResponse{}, ErrOperationNotSupported.With(
			"capability", string(CapabilityTranslation),
			"operation", string(OpTranslate),
			"provider", provider.ID(),
		)
	}
	return backend.Translate(ctx, req)
}

// ProviderFor reports which provider handles a translation operation.
func (t *Translation) ProviderFor(operation Operation) (string, error) {
	return t.router.ProviderIDFor(CapabilityTranslation, operation)
}
