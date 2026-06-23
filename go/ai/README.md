# AI (`appkit/go/ai`)

YAML-driven routing from AI **service types** to **providers** (OpenAI first).

## Service types

| Type | Purpose |
|------|---------|
| `translation` | Plain text translation |
| `chat` | Conversational agent (phase 2) |
| `chat_translation` | Understand + act + reply in customer language (Solidia) |

## Enable

```yaml
ai:
  enabled: true
  providers:
    openai:
      driver: openai
      api_key_env: OPENAI_API_KEY
      base_url: https://api.openai.com/v1
      default_model: gpt-4o-mini
  routes:
    translation: openai
    chat: openai
    chat_translation: openai
```

See `ai.example.yaml` for a full sample.

## Usage

```go
resp, err := aiService.Translate(ctx, ai.TranslateRequest{
    Text:           "Olá",
    TargetLanguage: "en",
})
```

## API key

Set `OPENAI_API_KEY` in the environment, or `providers.openai.api_key` in config (not recommended for production).

## Phase 1 scope

- `Translate` via OpenAI chat completions
- Config validation at `Wire` / `NewService`
- `chat` and `chat_translation` are routed in YAML; methods arrive in a later phase
