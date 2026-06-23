package social

import "strings"

// DispatchMode controls whether the server calls the platform API or returns a client job.
type DispatchMode string

const (
	DispatchServer DispatchMode = "server"
	DispatchClient DispatchMode = "client"
)

// ParseDispatchMode normalizes a dispatch mode string.
func ParseDispatchMode(raw string) (DispatchMode, error) {
	switch DispatchMode(strings.TrimSpace(strings.ToLower(raw))) {
	case "", DispatchServer:
		return DispatchServer, nil
	case DispatchClient:
		return DispatchClient, nil
	default:
		return "", invalidConfigf("dispatch", "unsupported dispatch mode %q (use server or client)", raw)
	}
}
