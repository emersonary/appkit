package config

import (
	"github.com/emersonary/appkit/accounts"
	"github.com/emersonary/appkit/currency"
	"github.com/emersonary/appkit/email"
	"github.com/emersonary/appkit/tenants"
)

// Blocks holds optional appkit block configs from the main YAML file.
type Blocks struct {
	Accounts         accounts.AppConfig       `mapstructure:"accounts" json:"accounts"`
	Tenants          tenants.AppConfig        `mapstructure:"tenants" json:"tenants"`
	Currency         currency.AppConfig       `mapstructure:"currency" json:"currency"`
	Email            email.Config             `mapstructure:"email" json:"email"`
	MailProvisioning email.ProvisioningConfig `mapstructure:"mail_provisioning" json:"mail_provisioning"`
}
