// Smoke-test appkit AI against a live OpenAI / ChatGPT key.
//
// Usage (from appkit/go):
//
//	cp ai/cmd/smoke/secret.example.yaml ai/cmd/smoke/secret.yaml
//	# edit secret.yaml with your key, then:
//	go run ./ai/cmd/smoke/          (from appkit/go)
//	go run .                        (from ai/cmd/smoke)
//
// Or: export OPENAI_API_KEY=sk-...
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/emersonary/appkit/ai"
	"gopkg.in/yaml.v3"
)

type smokeSecret struct {
	OpenAIAPIKey string `yaml:"openai_api_key"`
}

func loadOpenAIAPIKey() (string, error) {
	if key := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); key != "" {
		return key, nil
	}

	path := strings.TrimSpace(os.Getenv("AI_SMOKE_SECRET_FILE"))
	if path == "" {
		for _, candidate := range []string{"secret.yaml", "ai/cmd/smoke/secret.yaml"} {
			if _, err := os.Stat(candidate); err == nil {
				path = candidate
				break
			}
		}
		if path == "" {
			path = "secret.yaml"
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("OPENAI_API_KEY is not set and %s not found (copy secret.example.yaml to secret.yaml)", path)
		}
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	var secret smokeSecret
	if err := yaml.Unmarshal(data, &secret); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}

	key := strings.TrimSpace(secret.OpenAIAPIKey)
	if key == "" || key == "sk-your-key-here" {
		return "", fmt.Errorf("%s: openai_api_key is empty (set OPENAI_API_KEY or fill secret.yaml)", path)
	}

	return key, nil
}

func main() {
	key, err := loadOpenAIAPIKey()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.Setenv("OPENAI_API_KEY", key); err != nil {
		fmt.Fprintf(os.Stderr, "set OPENAI_API_KEY: %v\n", err)
		os.Exit(1)
	}

	svc, err := ai.Wire(context.Background(), smokeConfig(), ai.WireOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "wire ai: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	failed := false
	fail := func(label string, err error) {
		if err != nil {
			fmt.Printf("%s: ERROR %v\n", label, err)
			failed = true
			return
		}
	}

	text := "Olá mundo"
	fmt.Printf("Detect Request: %q\n", text)
	detect, err := svc.Translation().Detect(ctx, text)
	fail("Detect", err)
	if err == nil {
		fmt.Printf(" - Detect: language=%q provider=%q\n", detect.Language, detect.Provider)
	}

	text = "Olá, este é um teste."
	fmt.Printf("\nTranslate Request to English: %q\n", text)
	plain, err := svc.Translation().Translate(ctx, ai.TranslateRequest{
		Text:           text,
		TargetLanguage: "en",
	})
	fail("Translate", err)
	if err == nil {
		fmt.Printf(" - Translate: %q (provider=%q model=%q source=%q)\n",
			plain.Text, plain.Provider, plain.Model, plain.SourceLanguage)
	}

	text = "<p>Olá <strong>mundo</strong></p>"
	fmt.Printf("\nTranslate Request (HTML) to English: %q\n", text)
	html, err := svc.Translation().Translate(ctx, ai.TranslateRequest{
		Text:           text,
		TargetLanguage: "en",
	})
	fail("Translate (HTML input)", err)
	if err == nil {
		fmt.Printf(" - Translate (HTML input): %q\n", html.Text)
	}

	text = "Say hi in one word."
	fmt.Printf("\nChat Request: %q\n", text)
	chat, err := svc.Chat().Send(ctx, ai.ChatRequest{
		Messages: []ai.ChatMessage{{Role: "user", Content: text}},
	})
	fail("Chat", err)
	if err == nil {
		fmt.Printf(" - Chat: %q (provider=%q)\n", chat.Message.Content, chat.Provider)
	}

	fmt.Println("\nRoutes:", svc.RouteSummary())

	if failed {
		os.Exit(1)
	}
}

func smokeConfig() ai.AIConfig {
	return ai.AIConfig{
		Enabled: true,
		Providers: map[string]ai.ProviderConfig{
			"openai": {Driver: "openai"},
			"local":  {Driver: "local"},
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
