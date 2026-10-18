package messenger_test

import (
	"testing"

	"github.com/emersonary/appkit/messenger"
)

func TestConfigValidateDefaults(t *testing.T) {
	cfg := messenger.Config{}
	cfg.Schema = ""
	cfg.AccountSchema = ""
	cfg.TenantSchema = ""

	app := messenger.AppConfig{Enabled: true, Schema: "", AccountSchema: "", TenantSchema: ""}
	resolved, err := messenger.ResolveBlockConfig(app)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Schema != "messenger" {
		t.Fatalf("expected default schema messenger, got %q", resolved.Schema)
	}
}
