package email

import "strings"

// Config holds outbound mail settings (SMTP).
type Config struct {
	From  string     `mapstructure:"from" json:"from"`
	Brand string     `mapstructure:"brand" json:"brand"`
	SMTP  SMTPConfig `mapstructure:"smtp" json:"smtp"`
}

// SMTPConfig configures SMTP submission.
type SMTPConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"-"`
}

func (s SMTPConfig) Enabled() bool {
	return strings.TrimSpace(s.Host) != "" && s.Port > 0
}

// ProvisioningConfig configures Stalwart mailbox provisioning via JMAP.
type ProvisioningConfig struct {
	Enabled    bool   `mapstructure:"enabled" json:"enabled"`
	JMAPURL    string `mapstructure:"jmap_url" json:"jmap_url"`
	APIToken   string `mapstructure:"api_token" json:"-"`
	DomainID   string `mapstructure:"domain_id" json:"domain_id"`
	Password   string `mapstructure:"password" json:"-"`
	MailDomain string `mapstructure:"mail_domain" json:"mail_domain"`
}

func (p ProvisioningConfig) Active() bool {
	return p.Enabled &&
		strings.TrimSpace(p.JMAPURL) != "" &&
		strings.TrimSpace(p.APIToken) != "" &&
		strings.TrimSpace(p.DomainID) != "" &&
		strings.TrimSpace(p.Password) != "" &&
		strings.TrimSpace(p.MailDomain) != ""
}
