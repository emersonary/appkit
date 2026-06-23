package ai

import (
	"fmt"
	"strings"
)

// ServiceType identifies which AI capability is requested.
type ServiceType string

const (
	ServiceTypeTranslation     ServiceType = "translation"
	ServiceTypeChat            ServiceType = "chat"
	ServiceTypeChatTranslation ServiceType = "chat_translation"
)

func ParseServiceType(raw string) (ServiceType, error) {
	switch ServiceType(strings.TrimSpace(raw)) {
	case ServiceTypeTranslation:
		return ServiceTypeTranslation, nil
	case ServiceTypeChat:
		return ServiceTypeChat, nil
	case ServiceTypeChatTranslation:
		return ServiceTypeChatTranslation, nil
	default:
		return "", fmt.Errorf("unknown ai service type %q", raw)
	}
}

func (t ServiceType) String() string {
	return string(t)
}

// KnownServiceTypes lists routable service types for validation defaults.
func KnownServiceTypes() []ServiceType {
	return []ServiceType{
		ServiceTypeTranslation,
		ServiceTypeChat,
		ServiceTypeChatTranslation,
	}
}
