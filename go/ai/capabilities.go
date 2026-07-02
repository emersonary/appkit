package ai

// Capability identifies a high-level AI feature area.
type Capability string

const (
	CapabilityTranslation Capability = "translation"
	CapabilityChat        Capability = "chat"
)

// Operation identifies a method within a capability.
type Operation string

const (
	OpDetect    Operation = "detect"
	OpTranslate Operation = "translate"
	OpSend      Operation = "send"
)

// OperationKey is a routable capability method.
type OperationKey struct {
	Capability Capability
	Operation  Operation
}

func (k OperationKey) String() string {
	return string(k.Capability) + "." + string(k.Operation)
}

// KnownCapabilities lists supported capability areas.
func KnownCapabilities() []Capability {
	return []Capability{
		CapabilityTranslation,
		CapabilityChat,
	}
}

// KnownOperations lists valid operations for a capability.
func KnownOperations(capability Capability) []Operation {
	switch capability {
	case CapabilityTranslation:
		return []Operation{OpDetect, OpTranslate}
	case CapabilityChat:
		return []Operation{OpSend}
	default:
		return nil
	}
}

func parseCapability(raw string) (Capability, error) {
	switch Capability(raw) {
	case CapabilityTranslation, CapabilityChat:
		return Capability(raw), nil
	default:
		return "", invalidConfig("capability", "unknown capability "+quote(raw))
	}
}

func parseOperation(capability Capability, raw string) (Operation, error) {
	op := Operation(raw)
	for _, known := range KnownOperations(capability) {
		if op == known {
			return op, nil
		}
	}
	return "", invalidConfig("operation", "unknown operation "+quote(raw)+" for capability "+quote(string(capability)))
}

func quote(s string) string {
	return `"` + s + `"`
}
