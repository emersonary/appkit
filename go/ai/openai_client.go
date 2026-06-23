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

type openAIClientConfig struct {
	ProviderID string
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    time.Duration
	HTTPClient *http.Client
}

type openAIClient struct {
	providerID string
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

func newOpenAIClient(cfg openAIClientConfig) (*openAIClient, error) {
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

	return &openAIClient{
		providerID: strings.TrimSpace(cfg.ProviderID),
		apiKey:     cfg.APIKey,
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		model:      cfg.Model,
		httpClient: httpClient,
	}, nil
}

func (c *openAIClient) Translate(ctx context.Context, req TranslateRequest) (TranslateResponse, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return TranslateResponse{}, ErrInvalidRequest.With("text", "required")
	}
	target := strings.TrimSpace(req.TargetLanguage)
	if target == "" {
		return TranslateResponse{}, ErrInvalidRequest.With("target_language", "required")
	}

	source := strings.TrimSpace(req.SourceLanguage)
	payload := openAIChatCompletionRequest{
		Model: c.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: openAITranslateSystemPrompt(source, target)},
			{Role: "user", Content: text},
		},
		Temperature: 0,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return TranslateResponse{}, wrapErr(ErrTranslateFailed, "marshal", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return TranslateResponse{}, wrapErr(ErrTranslateFailed, "request", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return TranslateResponse{}, wrapErr(ErrTranslateFailed, "http", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TranslateResponse{}, wrapErr(ErrTranslateFailed, "read", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return TranslateResponse{}, ErrTranslateFailed.With(
			"status", fmt.Sprintf("%d", resp.StatusCode),
			"body", truncateString(string(respBody), 512),
		)
	}

	var completion openAIChatCompletionResponse
	if err := json.Unmarshal(respBody, &completion); err != nil {
		return TranslateResponse{}, wrapErr(ErrTranslateFailed, "decode", err)
	}

	translated := strings.TrimSpace(completion.firstMessage())
	if translated == "" {
		return TranslateResponse{}, ErrTranslateFailed.With("response", "empty translation")
	}

	return TranslateResponse{
		Text:           translated,
		SourceLanguage: source,
		TargetLanguage: target,
		Provider:       c.providerID,
		Model:          c.model,
	}, nil
}

func openAITranslateSystemPrompt(sourceLanguage, targetLanguage string) string {
	if sourceLanguage == "" {
		return fmt.Sprintf(
			"You are a professional translator. Detect the source language and translate the user message into %s. Return only the translated text with no commentary.",
			targetLanguage,
		)
	}
	return fmt.Sprintf(
		"You are a professional translator. Translate the user message from %s to %s. Return only the translated text with no commentary.",
		sourceLanguage,
		targetLanguage,
	)
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

type openAIChatCompletionRequest struct {
	Model       string             `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float64            `json:"temperature"`
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
