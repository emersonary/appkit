# Binding `@emersonary/appkit-accounts` to Sahar

Checklist for linking this package to `sahar/web`.

## Layout

```
via-jeri/appkit/
  blocks/account/proto/v1/account.proto   ← source of truth
  blocks/account/go/                      ← Go backend service
  blocks/account/web/                     ← this npm package
sahar/
  proto/account/v1/account.proto        ← synced copy (go_package differs)
  web/src/accounts/                     ← Sahar-only glue (AccountWorkflow, membership)
```

## Steps

### 1. Install package

`via-jeri/package.json` workspace entry:

```json
"appkit/blocks/account/web"
```

`sahar/web/package.json`:

```json
"@emersonary/appkit-accounts": "file:../../via-jeri/appkit/blocks/account/web"
```

Run `npm install` in both roots.

### 2. Vite config (`sahar/web/vite.config.ts`)

- Alias `@emersonary/appkit-accounts` → `.../appkit/blocks/account/web/src/index.ts`
- Add `appkit/blocks/account/web` to `server.fs.allow`
- Keep `dedupe: ['react', 'react-dom']`

### 3. Sahar glue files (stay in Sahar, not appkit)

| File | Role |
|------|------|
| `accounts/connect/saharAccountClient.ts` | `createConnectAccountClient` + `sahar_account_session` storage |
| `accounts/context/AccountContext.tsx` | `AccountsProvider` + membership + i18n errors |
| `accounts/pages/AccountWorkflow.tsx` | Hero, i18n, routing → `ConnectedLoginWorkflow` |
| `accounts/components/UserAccountMenu.tsx` | `AccountMenu` + Sahar theme CSS |

### 4. Imports — use the new package name

Replace `@emersonary/appkit-ui` with `@emersonary/appkit-accounts` in:

- `AccountContext.tsx`
- `saharAccountClient.ts`
- `AccountWorkflow.tsx`
- `UserAccountMenu.tsx`

### 5. Proto workflow

```bash
# Sync account.proto appkit → Sahar
node via-jeri/scripts/sync-account-proto-to-sahar.mjs

# Regenerate stubs
cd via-jeri/appkit/blocks/account/web && npm run generate:account-proto
cd sahar && make proto-go
cd sahar/web && npm run generate:proto
```

### 6. Verify

```bash
cd sahar/web && npx tsc -b && npm run build
cd sahar/api && go build ./...
```

## Peer dependencies

Consumer must install (Sahar already has these):

- `react`, `react-dom`
- `@connectrpc/connect`, `@connectrpc/connect-web`
- `@bufbuild/protobuf`
