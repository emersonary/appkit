package ai

import (
	"fmt"
	"strings"
)

// CapabilityRoute maps operations within one capability to providers.
type CapabilityRoute struct {
	Default    string
	Operations map[string]string
}

// RouteTable holds capability-level defaults and per-operation overrides.
type RouteTable struct {
	GlobalDefault string
	Capabilities  map[Capability]CapabilityRoute
}

func (t RouteTable) IsEmpty() bool {
	return len(t.Capabilities) == 0 && strings.TrimSpace(t.GlobalDefault) == ""
}

// ParseRouteMap builds a RouteTable from YAML/JSON route config.
func ParseRouteMap(raw map[string]any) (RouteTable, error) {
	table := RouteTable{Capabilities: make(map[Capability]CapabilityRoute)}
	if len(raw) == 0 {
		return table, nil
	}

	for key, value := range raw {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if key == "_default" {
			providerID, err := asProviderID(value)
			if err != nil {
				return RouteTable{}, invalidConfig("routes._default", err.Error())
			}
			table.GlobalDefault = providerID
			continue
		}

		capability, err := parseCapability(key)
		if err != nil {
			return RouteTable{}, err
		}

		route, err := parseCapabilityRoute(value)
		if err != nil {
			return RouteTable{}, invalidConfigf("routes.%s", key, "%s", err.Error())
		}
		table.Capabilities[capability] = route
	}

	return table, nil
}

func parseCapabilityRoute(value any) (CapabilityRoute, error) {
	switch v := value.(type) {
	case string:
		providerID, err := asProviderID(v)
		if err != nil {
			return CapabilityRoute{}, err
		}
		return CapabilityRoute{Default: providerID, Operations: map[string]string{}}, nil
	case map[string]any:
		return parseCapabilityRouteMap(v)
	case map[any]any:
		normalized := make(map[string]any, len(v))
		for k, val := range v {
			normalized[fmt.Sprint(k)] = val
		}
		return parseCapabilityRouteMap(normalized)
	default:
		return CapabilityRoute{}, fmt.Errorf("expected string or map, got %T", value)
	}
}

func parseCapabilityRouteMap(raw map[string]any) (CapabilityRoute, error) {
	route := CapabilityRoute{Operations: map[string]string{}}

	if defaultRaw, ok := raw["default"]; ok {
		providerID, err := asProviderID(defaultRaw)
		if err != nil {
			return CapabilityRoute{}, fmt.Errorf("default: %w", err)
		}
		route.Default = providerID
	}

	opsRaw, ok := raw["operations"]
	if !ok {
		if route.Default == "" {
			return CapabilityRoute{}, fmt.Errorf("default or operations is required")
		}
		return route, nil
	}

	opsMap, err := asStringMap(opsRaw)
	if err != nil {
		return CapabilityRoute{}, fmt.Errorf("operations: %w", err)
	}
	route.Operations = opsMap

	if route.Default == "" && len(route.Operations) == 0 {
		return CapabilityRoute{}, fmt.Errorf("default or operations is required")
	}
	return route, nil
}

func asProviderID(value any) (string, error) {
	s, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected provider id string, got %T", value)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("provider id is required")
	}
	return s, nil
}

func asStringMap(value any) (map[string]string, error) {
	switch v := value.(type) {
	case map[string]string:
		out := make(map[string]string, len(v))
		for k, val := range v {
			k = strings.TrimSpace(k)
			val = strings.TrimSpace(val)
			if k == "" || val == "" {
				continue
			}
			out[k] = val
		}
		return out, nil
	case map[string]any:
		out := make(map[string]string, len(v))
		for k, raw := range v {
			id, err := asProviderID(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", k, err)
			}
			out[strings.TrimSpace(k)] = id
		}
		return out, nil
	case map[any]any:
		normalized := make(map[string]any, len(v))
		for k, val := range v {
			normalized[fmt.Sprint(k)] = val
		}
		return asStringMap(normalized)
	default:
		return nil, fmt.Errorf("expected map, got %T", value)
	}
}

func (t RouteTable) resolve(capability Capability, operation Operation) (string, bool) {
	route, ok := t.Capabilities[capability]
	if ok {
		if providerID, ok := route.Operations[string(operation)]; ok {
			providerID = strings.TrimSpace(providerID)
			if providerID != "" {
				return providerID, true
			}
		}
		if providerID := strings.TrimSpace(route.Default); providerID != "" {
			return providerID, true
		}
	}
	if providerID := strings.TrimSpace(t.GlobalDefault); providerID != "" {
		return providerID, true
	}
	return "", false
}

func (t RouteTable) summary() map[string]string {
	out := make(map[string]string)
	for capability, route := range t.Capabilities {
		if defaultID := strings.TrimSpace(route.Default); defaultID != "" {
			out[string(capability)+".default"] = defaultID
		}
		for op, providerID := range route.Operations {
			out[string(capability)+"."+op] = providerID
		}
	}
	if global := strings.TrimSpace(t.GlobalDefault); global != "" {
		out["_default"] = global
	}
	return out
}
