# appkit protos

Language-neutral API contracts shared by `go/` and `web/`.

| File | Consumers |
|------|-----------|
| `auth/v1/auth.proto` | `go/gen/auth/v1/`, `web/accounts/src/gen/auth/v1/` |

## Workflow

1. Edit `.proto` here (source of truth).
2. Regenerate stubs from appkit root:

   ```bash
   make proto
   ```

3. Sahar copies `auth.proto` with its own `go_package` via `via-jeri/scripts/sync-auth-proto-to-sahar.mjs`.

Do not edit generated `gen/` trees by hand.
