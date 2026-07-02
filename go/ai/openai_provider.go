package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type openAIProviderConfig struct {
	ProviderID string
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type openAIProvider struct {
	providerID string
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func newOpenAIProvider(cfg openAIProviderConfig) (*openAIProvider, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, ErrInvalidConfig.With("api_key", "required")
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, ErrInvalidConfig.With("base_url", "required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, ErrInvalidConfig.With("model", "required")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}

	return &openAIProvider{
		providerID: strings.TrimSpace(cfg.ProviderID),
		apiKey:     cfg.APIKey,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		model:      cfg.Model,
		httpClient: httpClient,
	}, nil
}

func (p *openAIProvider) ID() string   { return p.providerID }
func (p *openAIProvider) Driver() string { return defaultOpenAIDriver }

func (p *openAIProvider) Translation() TranslationBackend { return p }
func (p *openAIProvider) Chat() ChatBackend               { return p }

func (p *openAIProvider) Detect(ctx context.Context, text string) (DetectResponse, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return DetectResponse{}, ErrInvalidRequest.With("text", "required")
	}

	content, err := p.complete(ctx, openAIDetectSystemPrompt(), text, 0)
	if err != nil {
		return DetectResponse{}, err
	}

	language := normalizeLanguageCode(cleanModelText(content))
	if language == "" {
		return DetectResponse{}, ErrDetectFailed.With("response", "empty language code")
	}

	return DetectResponse{
		Language: language,
		Provider: p.providerID,
		Model:    p.model,
	}, nil
}

func (p *openAIProvider) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return TranslateResponse{}, ErrInvalidRequest.With("text", "required")
	}
	target := strings.TrimSpace(req.TargetLanguage)
	if target == "" {
		return TranslateResponse{}, ErrInvalidRequest.With("target_language", "required")
	}

	source := strings.TrimSpace(req.SourceLanguage)
	prompt := openAITranslateSystemPrompt(source, target, translationLooksLikeHTML(text))
	content, err := p.complete(ctx, prompt, text, 0)
	if err != nil {
		return TranslateResponse{}, err
	}
	content = cleanModelText(content)

	if source == "" {
		if detected, detectErr := p.Detect(ctx, text); detectErr == nil {
			source = detected.Language
		}
	}

	return TranslateResponse{
		Text:           content,
		SourceLanguage: source,
		TargetLanguage: target,
		Provider:       p.providerID,
		Model:          p.model,
	}, nil
}

func (p *openAIProvider) Send(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	messages := make([]openAIChatMessage, 0, len(req.Messages)+1)
	if system := strings.TrimSpace(req.SystemPrompt); system != "" {
		messages = append(messages, openAIChatMessage{Role: "system", Content: system})
	}
	for _, msg := range req.Messages {
		role := strings.TrimSpace(msg.Role)
		content := strings.TrimSpace(msg.Content)
		if role == "" || content == "" {
			continue
		}
		messages = append(messages, openAIChatMessage{Role: role, Content: content})
	}
	if len(messages) == 0 {
		return ChatResponse{}, ErrInvalidRequest.With("messages", "required")
	}

	payload := openAIChatCompletionRequest{
		Model:       p.model,
		Messages:    messages,
		Temperature: 0.2,
	}

	content, err := p.completePayload(ctx, payload)
	if err != nil {
		return ChatResponse{}, err
	}

	return ChatResponse{
		Message: ChatMessage{
			Role:    "assistant",
			Content: content,
		},
		Provider: p.providerID,
		Model:    p.model,
	}, nil
}

func (p *openAIProvider) complete(ctx context.Context, systemPrompt, userContent string, temperature float64) (string, error) {
	return p.completePayload(ctx, openAIChatCompletionRequest{
		Model: p.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		Temperature: temperature,
	})
}

func (p *openAIProvider) completePayload(ctx context.Context, payload openAIChatCompletionRequest) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", wrapErr(ErrCompletionFailed, "marshal", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", wrapErr(ErrCompletionFailed, "request", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", wrapErr(ErrCompletionFailed, "http", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", wrapErr(ErrCompletionFailed, "read", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", ErrCompletionFailed.With(
			"status", fmt.Sprintf("%d", resp.StatusCode),
			"body", truncateString(string(respBody), 512),
		)
	}

	var completion openAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return "", wrapErr(ErrCompletionFailed, "decode", err)
	}

	content := strings.TrimSpace(cleanModelText(completion.firstMessage()))
	if content == "" {
		return "", ErrCompletionFailed.With("response", "empty completion")
	}
	return content, nil
}

func openAIDetectSystemPrompt() string {
	return "Detect the language of the user message. Return only the ISO 639-1 language code in lowercase with no commentary."
}

func openAITranslateSystemPrompt(sourceLanguage, targetLanguage string, preserveHTML bool) string {
	htmlNote := ""
	if preserveHTML {
		htmlNote = " Preserve HTML tags and structure."
	}
	if sourceLanguage == "" {
		return fmt.Sprintf(
			"You are a professional translator. Detect the source language and translate the user message into %s.%s Return only the translated text with no commentary.",
			targetLanguage,
			htmlNote,
		)
	}
	return fmt.Sprintf(
		"You are a professional translator. Translate the user message from %s to %s.%s Return only the translated text with no commentary.",
		sourceLanguage,
		targetLanguage,
		htmlNote,
	)
}

func normalizeLanguageCode(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}
	if idx := strings.IndexAny(raw, " \t\n\r(,{"); idx >= 0 {
		raw = raw[:idx]
	}
	return raw
}

type openAIChatCompletionRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatCompletionResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
}

func (r openAIChatCompletionResponse) firstMessage() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
