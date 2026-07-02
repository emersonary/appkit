package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func testAIConfig(t *testing.T) AIConfig {
	t.Helper()
	os.Setenv("OPENAI_API_KEY", "test-key")
	t.Cleanup(func() { os.Unsetenv("OPENAI_API_KEY") })

	return AIConfig{
		Enabled: true,
		Providers: map[string]ProviderConfig{
			"openai": {
				Driver:       "openai",
				BaseURL:      "https://example.test/v1",
				DefaultModel: "gpt-4o-mini",
			},
			"local": {
				Driver: "local",
			},
		},
		Routes: map[string]any{
			"translation": map[string]any{
				"default": "openai",
				"operations": map[string]any{
					"detect": "local",
				},
			},
			"chat": map[string]any{
				"default": "openai",
			},
		},
	}
}

func TestAIConfigRequiresTranslationRoute(t *testing.T) {
	cfg := testAIConfig(t)
	cfg.Routes = map[string]any{
		"chat": map[string]any{"default": "openai"},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing translation.translate route")
	}
}

func TestAIConfigRequiresAPIKeyForOpenAI(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	cfg := testAIConfig(t)

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing api key")
	}
}

func TestAIConfigRejectsUnsupportedOperationRoute(t *testing.T) {
	cfg := testAIConfig(t)
	cfg.Routes = map[string]any{
		"translation": map[string]any{
			"operations": map[string]any{
				"translate": "local",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error when local is routed for translate")
	}
}

func TestServiceTranslate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Hello"}}]}`))
	}))
	defer server.Close()

	cfg := testAIConfig(t)
	cfg.Providers["openai"] = ProviderConfig{
		Driver:       "openai",
		BaseURL:      server.URL + "/v1",
		DefaultModel: "gpt-4o-mini",
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	resp, err := svc.Translation().Translate(context.Background(), TranslateRequest{
		Text:           "Olá",
		TargetLanguage: "en",
	})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if resp.Text != "Hello" {
		t.Fatalf("got %q", resp.Text)
	}
	if resp.Provider != "openai" {
		t.Fatalf("provider: got %q", resp.Provider)
	}
}

func TestServiceDetectUsesLocalRoute(t *testing.T) {
	cfg := testAIConfig(t)
	cfg.Providers["openai"] = ProviderConfig{
		Driver:       "openai",
		BaseURL:      "https://example.test/v1",
		DefaultModel: "gpt-4o-mini",
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	resp, err := svc.Translation().Detect(context.Background(), "Hello world")
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if resp.Provider != "local" {
		t.Fatalf("expected local provider, got %q", resp.Provider)
	}
	if resp.Language == "" {
		t.Fatal("expected language code")
	}
}

func TestWireDisabled(t *testing.T) {
	svc, err := Wire(context.Background(), AIConfig{Enabled: false}, WireOptions{})
	if err != nil {
		t.Fatalf("Wire: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when disabled")
	}
}

func TestParseRouteMapLegacyShorthand(t *testing.T) {
	table, err := ParseRouteMap(map[string]any{
		"translation": "openai",
		"chat":        "openai",
	})
	if err != nil {
		t.Fatal(err)
	}
	if providerID, ok := table.resolve(CapabilityTranslation, OpTranslate); !ok || providerID != "openai" {
		t.Fatalf("translate route: got %q ok=%v", providerID, ok)
	}
}

func TestRouteSummary(t *testing.T) {
	cfg := testAIConfig(t)
	cfg.Providers["openai"] = ProviderConfig{
		Driver:       "openai",
		BaseURL:      "https://example.test/v1",
		DefaultModel: "gpt-4o-mini",
	}

	svc, err := NewService(cfg)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	summary := svc.RouteSummary()
	if summary["translation.detect"] != "local" {
		t.Fatalf("unexpected summary: %#v", summary)
	}
	if summary["translation.default"] != "openai" {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}
