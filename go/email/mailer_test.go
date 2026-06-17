package email

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestNewHandler_ClientUsesLogModeWhenSMTPDisabled(t *testing.T) {
	h := NewHandler(Config{}, ProvisioningConfig{}, nil)
	if h.Client == nil {
		t.Fatal("expected client")
	}
	if h.Client.SMTPConfigured() {
		t.Fatal("expected SMTP disabled")
	}
}

func TestNewHandler_ClientUsesSMTPWhenEnabled(t *testing.T) {
	h := NewHandler(Config{
		SMTP: SMTPConfig{Host: "smtp.example.com", Port: 587},
	}, ProvisioningConfig{}, nil)
	if !h.Client.SMTPConfigured() {
		t.Fatal("expected SMTP enabled")
	}
}

func TestNewHandler_APIOnlyWhenProvisioningActive(t *testing.T) {
	h := NewHandler(Config{}, ProvisioningConfig{Enabled: true}, nil)
	if h.API != nil {
		t.Fatal("expected nil API when provisioning fields incomplete")
	}

	h = NewHandler(Config{}, ProvisioningConfig{
		Enabled:    true,
		JMAPURL:    "https://mail.example.com/api",
		APIToken:   "token",
		DomainID:   "domain",
		Password:   "secret",
		MailDomain: "example.com",
	}, nil)
	if h.API == nil {
		t.Fatal("expected API when provisioning active")
	}
}

func TestClient_SendPlain_DevMode(t *testing.T) {
	c := NewClient(Config{}, zap.NewNop())
	if err := c.SendPlain(context.Background(), "a@b.com", "subject", "body"); err != nil {
		t.Fatalf("SendPlain: %v", err)
	}
}

func TestHandler_AccountMailer(t *testing.T) {
	h := NewHandler(Config{Brand: "Awaken Womb"}, ProvisioningConfig{}, zap.NewNop())
	m := h.AccountMailer()
	ctx := context.Background()
	if err := m.SendVerificationEmail(ctx, "a@b.com", "https://example.com/verify"); err != nil {
		t.Fatalf("SendVerificationEmail: %v", err)
	}
	if err := m.SendPasswordResetEmail(ctx, "a@b.com", "https://example.com/reset"); err != nil {
		t.Fatalf("SendPasswordResetEmail: %v", err)
	}
}
