# appkit

Shared platform libraries — **Go** (`go/`), **TypeScript/React** (`web/`), and **protobuf** (`proto/`).

| Track | Path | Import |
|-------|------|--------|
| Go module | [`go/`](go/) | `github.com/emersonary/appkit/...` |
| Accounts UI | [`web/accounts/`](web/accounts/) | `@emersonary/appkit-accounts` |
| Contracts | [`proto/`](proto/) | source of truth for codegen |

## Layout

```
appkit/
├── Makefile              # make proto, make test
├── proto/auth/v1/        # hand-written .proto
├── go/
│   ├── go.mod
│   ├── accounts/         # Go auth service
│   └── gen/auth/v1/      # generated Go stubs
└── web/
    └── accounts/         # @emersonary/appkit-accounts
        └── src/gen/      # generated TypeScript stubs
```

## Codegen

```bash
make proto-install   # once per machine
make proto           # proto-go + proto-ts
make test            # go test + tsc
```

## Local development

**Go** (via-jeri / Sahar):

```go
replace github.com/emersonary/appkit => ../appkit/go
```

**npm** (Sahar web):

```json
"@emersonary/appkit-accounts": "file:../../via-jeri/appkit/web/accounts"
```

## Docs

- Go packages: [`go/README.md`](go/README.md)
- Go publishing: [`go/PUBLISHING.md`](go/PUBLISHING.md)
- Accounts npm package: [`web/accounts/README.md`](web/accounts/README.md)
- Bind to Sahar: [`web/accounts/BINDING.md`](web/accounts/BINDING.md)
- Proto workflow: [`proto/README.md`](proto/README.md)

This directory is one git repository (`github.com/emersonary/appkit`).
