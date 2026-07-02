package ai

import "context"

// Provider is a configured AI backend (OpenAI, local, …).
type Provider interface {
	ID() string
	Driver() string
	Translation() TranslationBackend
	Chat() ChatBackend
}

// TranslationBackend performs translation capability operations for one provider.
type TranslationBackend interface {
	Detect(ctx context.Context, text string) (DetectResponse, error)
	Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error)
}

// ChatBackend performs chat capability operations for one provider.
type ChatBackend interface {
	Send(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

var driverOperations = map[string]map[Capability][]Operation{
	defaultOpenAIDriver: {
		CapabilityTranslation: {OpDetect, OpTranslate},
		CapabilityChat:        {OpSend},
	},
	defaultLocalDriver: {
		CapabilityTranslation: {OpDetect},
	},
}

func providerSupportsOperation(driver string, capability Capability, operation Operation) bool {
	opsByCap, ok := driverOperations[driver]
	if !ok {
		return false
	}
	ops, ok := opsByCap[capability]
	if !ok {
		return false
	}
	for _, op := range ops {
		if op == operation {
			return true
		}
	}
	return false
}
