# @emersonary/appkit-accounts

React accounts library: login/register workflow, header menu, Connect account client, shared `account.proto` stubs.

## What's included

- **UI:** `ConnectedLoginWorkflow`, `LoginWorkflow`, `AccountMenu`, `AccountsProvider`, …
- **API:** `createConnectAccountClient`, `createLocalStorageSessionStorage`, `AccountsError`
- **Proto:** generated TS from `appkit/blocks/account/proto/v1/account.proto` (`npm run generate:account-proto`)

Styling is headless — pass `classNames` and `labels` from the host app.

## Bind to a consumer app (e.g. Sahar)

### 1. Add the dependency

**via-jeri workspace** (monorepo root `package.json`):

```json
"workspaces": [
  "appkit/blocks/account/web"
]
```

**Sahar** (`web/package.json`):

```json
{
  "dependencies": {
    "@emersonary/appkit-accounts": "file:../../via-jeri/appkit/blocks/account/web"
  }
}
```

From repo root:

```bash
cd via-jeri && npm install
cd ../sahar/web && npm install
```

### 2. Vite alias (dev — compile from source)

```ts
// vite.config.ts
import path from 'node:path';

const accountsPkg = path.resolve(__dirname, '../via-jeri/appkit/blocks/account/web/src/index.ts');

export default defineConfig({
  resolve: {
    alias: {
      '@emersonary/appkit-accounts': accountsPkg,
    },
    dedupe: ['react', 'react-dom'],
  },
  server: {
    fs: {
      allow: [repoRoot, path.resolve(repoRoot, '../via-jeri/appkit/blocks/account/web')],
    },
  },
});
```

### 3. Import in the app

```tsx
import {
  AccountsProvider,
  ConnectedLoginWorkflow,
  AccountMenu,
  createConnectAccountClient,
  createLocalStorageSessionStorage,
} from '@emersonary/appkit-accounts';
```

### 4. Wire account client + provider

```tsx
const accountClient = createConnectAccountClient({
  baseUrl: '', // or VITE_API_URL
  storage: createLocalStorageSessionStorage('my_app_account_session'),
});

<AccountsProvider tenancy={{ enabled: false }} accountClient={accountClient}>
  {children}
</AccountsProvider>
```

### 5. Regenerate proto stubs

```bash
cd via-jeri/appkit/blocks/account/web
npm run generate:account-proto
```

Sahar API proto stays synced separately:

```bash
cd sahar && make sync-appkit-proto proto-go
cd sahar/web && npm run generate:proto
```

## Publishing (future)

Tag the **appkit** git repo and publish `@emersonary/appkit-accounts` to npm (private scope or GitHub Packages). Until then, use the `file:` path above.

See also [`BINDING.md`](BINDING.md) for the full Sahar integration checklist.
