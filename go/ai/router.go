package ai

import "strings"

// Router resolves capability operations to wired providers.
type Router struct {
	routes    RouteTable
	providers map[string]Provider
}

func newRouter(routes RouteTable, providers map[string]Provider) *Router {
	return &Router{
		routes:    routes,
		providers: providers,
	}
}

func (r *Router) ProviderFor(capability Capability, operation Operation) (Provider, error) {
	providerID, err := r.ProviderIDFor(capability, operation)
	if err != nil {
		return nil, err
	}
	provider, ok := r.providers[providerID]
	if !ok {
		return nil, ErrProviderNotFound.With("provider", providerID)
	}
	if !providerSupportsOperation(provider.Driver(), capability, operation) {
		return nil, ErrOperationNotSupported.With(
			"capability", string(capability),
			"operation", string(operation),
			"provider", providerID,
			"driver", provider.Driver(),
		)
	}
	return provider, nil
}

func (r *Router) ProviderIDFor(capability Capability, operation Operation) (string, error) {
	if _, err := parseOperation(capability, string(operation)); err != nil {
		return "", err
	}
	providerID, ok := r.routes.resolve(capability, operation)
	if !ok {
		return "", ErrOperationNotRouted.With(
			"capability", string(capability),
			"operation", string(operation),
		)
	}
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return "", ErrOperationNotRouted.With(
			"capability", string(capability),
			"operation", string(operation),
		)
	}
	if _, ok := r.providers[providerID]; !ok {
		return "", ErrProviderNotFound.With("provider", providerID)
	}
	return providerID, nil
}

func (r *Router) Summary() map[string]string {
	return r.routes.summary()
}
