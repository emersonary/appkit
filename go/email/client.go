package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// Client handles POP/IMAP/SMTP mailbox protocols. SMTP sending is implemented;
// IMAP and POP are reserved for future inbound mail support.
type Client struct {
	cfg    Config
	logger *zap.Logger
}

func NewClient(cfg Config, logger *zap.Logger) *Client {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Client{cfg: cfg, logger: logger}
}

// SMTPConfigured reports whether real SMTP delivery is configured.
func (c *Client) SMTPConfigured() bool {
	return c.cfg.SMTP.Enabled()
}

// SendPlain sends a plain-text message over SMTP, or logs it in dev mode.
func (c *Client) SendPlain(ctx context.Context, to, subject, body string) error {
	_ = ctx
	if !c.cfg.SMTP.Enabled() {
		c.logger.Info(
			"email (dev mode — configure email.smtp to send real mail)",
			zap.String("to", to),
			zap.String("subject", subject),
		)
		return nil
	}

	from := strings.TrimSpace(c.cfg.From)
	if from == "" {
		return fmt.Errorf("email: from address is required")
	}

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", c.cfg.SMTP.Host, c.cfg.SMTP.Port)
	auth := smtp.PlainAuth("", c.cfg.SMTP.Username, c.cfg.SMTP.Password, c.cfg.SMTP.Host)
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
