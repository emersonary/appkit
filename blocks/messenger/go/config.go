package messenger

import "strings"

const (
	UserTypeHuman  = "H"
	UserTypeAgent  = "A"
	UserTypeSystem = "S"

	ChannelKindDirect = "D"
	ChannelKindGroup  = "G"

	MessageKindUser   = "U"
	MessageKindSystem = "S"

	MemberRoleMember = "member"
	MemberRoleAdmin  = "admin"
)

type Config struct {
	Schema        string `yaml:"schema" mapstructure:"schema"`
	AccountSchema string `yaml:"account_schema" mapstructure:"account_schema"`
	TenantSchema  string `yaml:"tenant_schema" mapstructure:"tenant_schema"`
	AgentAPIKey   string `yaml:"agent_api_key" mapstructure:"agent_api_key" json:"-"`
}

func (c *Config) normalize() {
	c.Schema = strings.TrimSpace(c.Schema)
	if c.Schema == "" {
		c.Schema = "messenger"
	}
	c.AccountSchema = strings.TrimSpace(c.AccountSchema)
	if c.AccountSchema == "" {
		c.AccountSchema = "account"
	}
	c.TenantSchema = strings.TrimSpace(c.TenantSchema)
	if c.TenantSchema == "" {
		c.TenantSchema = "tenant"
	}
}

func (c Config) Validate() error {
	for _, ident := range []string{c.Schema, c.AccountSchema, c.TenantSchema} {
		if err := validateIdent(ident); err != nil {
			return ErrInvalidInput.With("field", ident, "reason", err.Error())
		}
	}
	return nil
}

func (c Config) accountsTable() string {
	return qualifiedName(c.AccountSchema, "accounts")
}

func (c Config) tenantsTable() string {
	return qualifiedName(c.TenantSchema, "tenants")
}
