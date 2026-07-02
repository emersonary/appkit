# AI (`appkit/go/ai`)

YAML-driven routing from AI **capability operations** to **providers**.

## Capabilities

| Capability | Operations |
|------------|------------|
| `translation` | `detect`, `translate` |
| `chat` | `send` |

Each operation can route to a different provider.

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
    local:
      driver: local
  routes:
    translation:
      default: openai
      operations:
        detect: local
        translate: openai
    chat:
      default: openai
```

Legacy shorthand still works:

```yaml
routes:
  translation: openai
  chat: openai
```

See `ai.example.yaml` for a full sample.

## Usage

```go
detected, err := aiService.Translation().Detect(ctx, "Olá mundo")
translated, err := aiService.Translation().Translate(ctx, ai.TranslateRequest{
    Text:           "Olá",
    TargetLanguage: "en",
})
reply, err := aiService.Chat().Send(ctx, ai.ChatRequest{
    Messages: []ai.ChatMessage{{Role: "user", Content: "Hello"}},
})
```

`Service.Translate(...)` remains as a convenience wrapper for `Translation().Translate(...)`.

### Live smoke test

From `appkit/go`, copy the gitignored secret file and add your key:

```bash
cp ai/cmd/smoke/secret.example.yaml ai/cmd/smoke/secret.yaml
# edit ai/cmd/smoke/secret.yaml — set openai_api_key
go run ./ai/cmd/smoke/
```

Or set `OPENAI_API_KEY` in the environment (takes precedence over `secret.yaml`).

## OpenAI / ChatGPT setup

The `openai` driver (alias: `chatgpt`) uses the [OpenAI Chat Completions API](https://platform.openai.com/docs/api-reference/chat) for translation and chat.

### 1. Create an OpenAI account

1. Go to [https://platform.openai.com](https://platform.openai.com) and sign up or log in.
2. Open **Settings → Billing** and add a payment method. API usage is pay-as-you-go (translation with `gpt-4o-mini` is inexpensive, but billing must be enabled).

### 2. Create an API key

1. Open [API keys](https://platform.openai.com/api-keys).
2. Click **Create new secret key** (name it e.g. `emersonary-dev`).
3. Copy the key immediately — it is shown only once.
4. Store it as an environment variable (never commit it to git):

```bash
export OPENAI_API_KEY="sk-..."
```

On Windows (PowerShell):

```powershell
$env:OPENAI_API_KEY = "sk-..."
```

On the posts/emersonary server, add the same variable to your service env file.

### 3. Optional: usage limits

In the OpenAI dashboard, set **Usage limits** / monthly budget so runaway jobs cannot overspend.

### 4. Enable appkit AI block

Add to your main config (or `config_path` YAML):

```yaml
ai:
  enabled: true
  providers:
    openai:
      driver: openai          # or chatgpt (same driver)
      api_key_env: OPENAI_API_KEY
      base_url: https://api.openai.com/v1
      default_model: gpt-4o-mini
      timeout: 30s
    local:
      driver: local
  routes:
    translation:
      default: openai
      operations:
        detect: local
        translate: openai
    chat:
      default: openai
```

Restart the API so `ai.Wire` loads the block (`runtime` logs `ai block enabled` with route summary).

### 5. Verify from Go

```go
svc, err := ai.Wire(ctx, cfg.AI, ai.WireOptions{})
detected, err := svc.Translation().Detect(ctx, "Olá mundo")
translated, err := svc.Translation().Translate(ctx, ai.TranslateRequest{
    Text:           "Olá",
    TargetLanguage: "en",
})
```

### Translation methods (OpenAI provider)

| Method | HTTP (vendor) | Behavior |
|--------|---------------|----------|
| `Detect` | `POST /v1/chat/completions` | Returns ISO 639-1 language code |
| `Translate` | `POST /v1/chat/completions` | Plain or HTML input (tags auto-detected); fills `SourceLanguage` via detect when omitted |

Both use system prompts; there is no separate OpenAI “Translate API” in this library.

## Providers

| Driver | Supported operations |
|--------|----------------------|
| `openai` / `chatgpt` | translation detect/translate (HTML auto-detected), chat send |
| `local` | translation detect only (offline heuristic) |

## API key

Set `OPENAI_API_KEY` in the environment, or `providers.openai.api_key` in config (not recommended for production).
