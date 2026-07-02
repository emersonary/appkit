package ai

import (
	"context"
	"strings"
	"unicode"
)

type localProvider struct {
	id string
}

func newLocalProvider(id string) *localProvider {
	return &localProvider{id: strings.TrimSpace(id)}
}

func (p *localProvider) ID() string     { return p.id }
func (p *localProvider) Driver() string { return defaultLocalDriver }

func (p *localProvider) Translation() TranslationBackend { return p }
func (p *localProvider) Chat() ChatBackend               { return nil }

func (p *localProvider) Detect(ctx context.Context, text string) (DetectResponse, error) {
	_ = ctx
	text = strings.TrimSpace(text)
	if text == "" {
		return DetectResponse{}, ErrInvalidRequest.With("text", "required")
	}

	language, confidence := detectLanguageLocal(text)
	return DetectResponse{
		Language:   language,
		Confidence: confidence,
		Provider:   p.id,
	}, nil
}

func (p *localProvider) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	_ = ctx
	_ = req
	return TranslateResponse{}, ErrOperationNotSupported.With(
		"capability", string(CapabilityTranslation),
		"operation", string(OpTranslate),
		"provider", p.id,
		"driver", p.Driver(),
	)
}

func detectLanguageLocal(text string) (language string, confidence float64) {
	for _, r := range text {
		switch {
		case unicode.In(r, unicode.Arabic):
			return "ar", 0.8
		case unicode.In(r, unicode.Hebrew):
			return "he", 0.8
		case unicode.In(r, unicode.Han):
			return "zh", 0.75
		case unicode.In(r, unicode.Hiragana, unicode.Katakana):
			return "ja", 0.8
		case unicode.In(r, unicode.Hangul):
			return "ko", 0.8
		case unicode.In(r, unicode.Cyrillic):
			return "ru", 0.75
		case unicode.In(r, unicode.Devanagari):
			return "hi", 0.75
		}
	}

	lower := strings.ToLower(text)
	switch {
	case strings.ContainsAny(lower, "ãõáéíóúàâêôç"):
		return "pt", 0.7
	case strings.ContainsAny(lower, "ñ¿¡"):
		return "es", 0.7
	case strings.ContainsAny(lower, "äöüß"):
		return "de", 0.65
	case strings.ContainsAny(lower, "àâçéèêëîïôùû"):
		return "fr", 0.6
	default:
		return "en", 0.55
	}
}
