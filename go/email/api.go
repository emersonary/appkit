package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// API talks to Stalwart's JMAP management API for mailbox provisioning.
type API struct {
	cfg    ProvisioningConfig
	client *http.Client
}

func NewAPI(cfg ProvisioningConfig) *API {
	if !cfg.Active() {
		return nil
	}
	return &API{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Configured reports whether Stalwart provisioning is active.
func (a *API) Configured() bool {
	return a != nil
}

// CreateUser provisions a Stalwart mailbox for localPart@mail_domain.
func (a *API) CreateUser(ctx context.Context, localPart string) error {
	if a == nil {
		return fmt.Errorf("email: stalwart api is not configured")
	}

	localPart = strings.TrimSpace(localPart)
	if localPart == "" {
		return fmt.Errorf("email: local mailbox name is required")
	}

	payload := map[string]any{
		"methodCalls": []any{
			[]any{
				"x:Account/set",
				map[string]any{
					"create": map[string]any{
						"new1": map[string]any{
							"@type":            "User",
							"name":             localPart,
							"domainId":         a.cfg.DomainID,
							"roles":            map[string]any{"@type": "User"},
							"permissions":      map[string]any{"@type": "Inherit"},
							"credentials":      map[string]any{"password": a.cfg.Password},
							"aliases":          map[string]any{},
							"quotas":           map[string]any{},
							"encryptionAtRest": map[string]any{"@type": "Disabled"},
						},
					},
				},
				"c1",
			},
		},
		"using": []string{
			"urn:ietf:params:jmap:core",
			"urn:stalwart:jmap",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email: marshal stalwart request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(a.cfg.JMAPURL, "/"), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build stalwart request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("email: stalwart request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("email: read stalwart response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email: stalwart api status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

// MailAddress returns the full address for a provisioned local part.
func (a *API) MailAddress(localPart string) string {
	if a == nil {
		return ""
	}
	localPart = strings.TrimSpace(localPart)
	domain := strings.TrimSpace(a.cfg.MailDomain)
	if localPart == "" || domain == "" {
		return ""
	}
	return localPart + "@" + domain
}
