package accounts

import "context"

type Mailer interface {
	SendVerificationEmail(ctx context.Context, email, verifyURL string) error
	SendPasswordResetEmail(ctx context.Context, email, resetURL string) error
}

type NoopMailer struct{}

func (NoopMailer) SendVerificationEmail(context.Context, string, string) error { return nil }
func (NoopMailer) SendPasswordResetEmail(context.Context, string, string) error { return nil }
