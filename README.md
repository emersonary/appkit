# appkit

Shared platform libraries — **Go** (`go/`), **blocks** (`blocks/`), **TypeScript/React** (`web/` legacy path removed for accounts), and **protobuf** (`proto/`).

| Track | Path | Import |
|-------|------|--------|
| Go module | [`go/`](go/) | `github.com/emersonary/appkit/...` |
| Account block (Go) | [`blocks/account/go/`](blocks/account/go/) | `github.com/emersonary/appkit/accounts/...` |
| Account block (UI) | [`blocks/account/web/`](blocks/account/web/) | `@emersonary/appkit-accounts` |
| Contracts | [`proto/`](proto/) | source of truth for codegen |

## Layout

```
appkit/
├── Makefile
├── proto/auth/v1/
├── go/                     # core Go module
│   ├── go.mod
│   ├── apperror/
│   ├── currency/
│   └── ...
└── blocks/
    └── account/            # feature block (Go + web together)
        ├── go/             # github.com/emersonary/appkit/accounts
        └── web/            # @emersonary/appkit-accounts
```

## Codegen

```bash
make proto-install   # once per machine
make proto           # auth proto → blocks/account/go/gen + web/src/gen
make test
```

## Local development

**Go core**:

```go
replace github.com/emersonary/appkit => ../appkit/go
```

**Accounts block**:

```go
replace github.com/emersonary/appkit/accounts => ../appkit/blocks/account/go
```

**npm**:

```json
"@emersonary/appkit-accounts": "file:../../via-jeri/appkit/blocks/account/web"
```

## Docs

- [`blocks/account/README.md`](blocks/account/README.md)
- Go packages: [`go/README.md`](go/README.md)
- Go publishing: [`go/PUBLISHING.md`](go/PUBLISHING.md)
- Accounts UI: [`blocks/account/web/README.md`](blocks/account/web/README.md)
- Proto: [`proto/README.md`](proto/README.md)

This directory is one git repository (`github.com/emersonary/appkit`).
