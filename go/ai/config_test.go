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
		},
		Routes: map[string]string{
			"translation":      "openai",
			"chat":             "openai",
			"chat_translation": "openai",
		},
	}
}

func TestAIConfigRequiresTranslationRoute(t *testing.T) {
	cfg := testAIConfig(t)
	delete(cfg.Routes, "translation")

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing translation route")
	}
}

func TestAIConfigRequiresAPIKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")
	cfg := testAIConfig(t)

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for missing api key")
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

	resp, err := svc.Translate(context.Background(), TranslateRequest{
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

func TestWireDisabled(t *testing.T) {
	svc, err := Wire(context.Background(), AIConfig{Enabled: false}, WireOptions{})
	if err != nil {
		t.Fatalf("Wire: %v", err)
	}
	if svc != nil {
		t.Fatal("expected nil service when disabled")
	}
}

func TestParseServiceType(t *testing.T) {
	if _, err := ParseServiceType("translation"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseServiceType("nope"); err == nil {
		t.Fatal("expected error")
	}
}
