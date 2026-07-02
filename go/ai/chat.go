package ai

import "context"

// Chat routes chat operations to configured providers.
type Chat struct {
	router *Router
}

func newChat(router *Router) *Chat {
	return &Chat{router: router}
}

func (c *Chat) Send(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	provider, err := c.router.ProviderFor(CapabilityChat, OpSend)
	if err != nil {
		return ChatResponse{}, err
	}
	backend := provider.Chat()
	if backend == nil {
		return ChatResponse{}, ErrOperationNotSupported.With(
			"capability", string(CapabilityChat),
			"operation", string(OpSend),
			"provider", provider.ID(),
		)
	}
	return backend.Send(ctx, req)
}

// ProviderFor reports which provider handles a chat operation.
func (c *Chat) ProviderFor(operation Operation) (string, error) {
	return c.router.ProviderIDFor(CapabilityChat, operation)
}
