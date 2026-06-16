# account proto

Language-neutral auth/account API contract for the account block.

| File | Consumers |
|------|-----------|
| `v1/account.proto` | `go/gen/account/v1/`, `web/src/gen/v1/` |

## Workflow

1. Edit `.proto` here (source of truth).
2. Regenerate stubs from appkit root:

   ```bash
   make proto
   ```

3. Sahar copies `account.proto` with its own `go_package` via `via-jeri/scripts/sync-auth-proto-to-sahar.mjs`.

Do not edit generated `gen/` trees by hand.
