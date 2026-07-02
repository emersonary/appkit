package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIProviderDetect(t *testing.T) {
	var systemPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload openAIChatCompletionRequest
		_ = json.Unmarshal(body, &payload)
		if len(payload.Messages) > 0 {
			systemPrompt = payload.Messages[0].Content
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"pt"}}]}`))
	}))
	defer server.Close()

	p := mustOpenAIProvider(t, server.URL)
	resp, err := p.Detect(context.Background(), "Olá mundo")
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if resp.Language != "pt" {
		t.Fatalf("language: got %q", resp.Language)
	}
	if !strings.Contains(systemPrompt, "ISO 639-1") {
		t.Fatalf("unexpected detect prompt: %q", systemPrompt)
	}
}

func TestOpenAIProviderTranslatePlain(t *testing.T) {
	var systemPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload openAIChatCompletionRequest
		_ = json.Unmarshal(body, &payload)
		if len(payload.Messages) > 0 {
			systemPrompt = payload.Messages[0].Content
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Hello"}}]}`))
	}))
	defer server.Close()

	p := mustOpenAIProvider(t, server.URL)
	resp, err := p.Translate(context.Background(), TranslateRequest{
		Text:           "Olá",
		TargetLanguage: "en",
	})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if resp.Text != "Hello" {
		t.Fatalf("text: got %q", resp.Text)
	}
	if strings.Contains(systemPrompt, "Preserve HTML") {
		t.Fatalf("plain text should not use HTML prompt: %q", systemPrompt)
	}
}

func TestOpenAIProviderTranslateHTML(t *testing.T) {
	var systemPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload openAIChatCompletionRequest
		_ = json.Unmarshal(body, &payload)
		if len(payload.Messages) > 0 {
			systemPrompt = payload.Messages[0].Content
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"<p>Hello</p>"}}]}`))
	}))
	defer server.Close()

	p := mustOpenAIProvider(t, server.URL)
	resp, err := p.Translate(context.Background(), TranslateRequest{
		Text:           "<p>Olá</p>",
		SourceLanguage: "pt",
		TargetLanguage: "en",
	})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if resp.Text != "<p>Hello</p>" {
		t.Fatalf("text: got %q", resp.Text)
	}
	if !strings.Contains(systemPrompt, "Preserve HTML") {
		t.Fatalf("HTML input should use HTML prompt: %q", systemPrompt)
	}
}

func TestOpenAIProviderTranslateFillsSourceLanguage(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		content := "Hello"
		if calls > 1 {
			content = "pt"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"` + content + `"}}]}`))
	}))
	defer server.Close()

	p := mustOpenAIProvider(t, server.URL)
	resp, err := p.Translate(context.Background(), TranslateRequest{
		Text:           "Olá",
		TargetLanguage: "en",
	})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if resp.SourceLanguage != "pt" {
		t.Fatalf("expected detected source pt, got %q", resp.SourceLanguage)
	}
}

func TestCleanModelText(t *testing.T) {
	got := cleanModelText("```\nHello\n```")
	if got != "Hello" {
		t.Fatalf("got %q", got)
	}
}

func mustOpenAIProvider(t *testing.T, serverURL string) *openAIProvider {
	t.Helper()
	p, err := newOpenAIProvider(openAIProviderConfig{
		ProviderID: "openai",
		APIKey:     "test-key",
		BaseURL:    serverURL + "/v1",
		Model:      "gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("newOpenAIProvider: %v", err)
	}
	return p
}
