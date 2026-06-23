# Social block

Publishing integrations and tenant-scoped OAuth for social platforms.

## Layout

- `go/` — module `github.com/emersonary/appkit/social` (publish clients, templates, OAuth)
- `web/` — `@emersonary/appkit-social` React components

## Tenant secrets

OAuth app credentials are declared per tenant in feed metadata:

```yaml
social:
  platforms:
    li:
      oauth:
        client_id_env: EMERSONARYDEV_LINKEDIN_CLIENT_ID
        client_secret_env: EMERSONARYDEV_LINKEDIN_CLIENT_SECRET
```

Each tenant uses its own `client_id_env` and `client_secret_env` (and optional `access_token_env` for dev fallback).

## OAuth routes

- `GET /auth/social/{platform}/start?token=…`
- `GET /auth/social/{platform}/callback`
- `GET /auth/social/{platform}/status`
- `DELETE /auth/social/{platform}`

## Frontend

```tsx
import { SocialConnectionCard } from '@emersonary/appkit-social';

<SocialConnectionCard
  platform="li"
  apiBaseUrl={import.meta.env.VITE_POSTS_API_URL ?? 'http://localhost:8090'}
  getAccessToken={() => readLaunchSession()?.token}
  labels={{ title: 'LinkedIn', connectButton: 'Conectar LinkedIn' }}
/>
```
