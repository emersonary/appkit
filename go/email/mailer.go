package email

import (
	"context"
	"fmt"

	"github.com/emersonary/appkit/accounts"
)

// AccountMailer adapts Handler.Client for the accounts block.
func (h *Handler) AccountMailer() accounts.Mailer {
	if h == nil || h.Client == nil {
		return accounts.NoopMailer{}
	}
	return &accountMailer{client: h.Client}
}

type accountMailer struct {
	client *Client
}

func (m *accountMailer) SendVerificationEmail(ctx context.Context, toEmail, verifyURL string) error {
	brand := m.client.cfg.Brand
	if brand == "" {
		brand = "your account"
	}
	subject := fmt.Sprintf("Verify your %s", brand)
	body := fmt.Sprintf(
		"Hello,\r\n\r\nPlease verify your email address by opening this link:\r\n\r\n%s\r\n\r\nIf you did not create an account, you can ignore this message.\r\n",
		verifyURL,
	)
	return m.client.SendPlain(ctx, toEmail, subject, body)
}

func (m *accountMailer) SendPasswordResetEmail(ctx context.Context, toEmail, resetURL string) error {
	brand := m.client.cfg.Brand
	if brand == "" {
		brand = "your account"
	}
	subject := fmt.Sprintf("Reset your %s password", brand)
	body := fmt.Sprintf(
		"Hello,\r\n\r\nWe received a request to reset your password. Open this link to choose a new password:\r\n\r\n%s\r\n\r\nThis link expires in one hour. If you did not request a reset, you can ignore this message.\r\n",
		resetURL,
	)
	return m.client.SendPlain(ctx, toEmail, subject, body)
}
