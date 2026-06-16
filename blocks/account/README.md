# account block

Cross-cutting **accounts** feature: Go service library and React UI shipped together.

```
blocks/account/
├── go/     # module github.com/emersonary/appkit/accounts
└── web/    # npm @emersonary/appkit-accounts
```

Proto contract: [`../../proto/auth/v1/`](../../proto/auth/v1/) (codegen into `go/gen/` and `web/src/gen/`).

## Codegen

From appkit root:

```bash
make proto
```

## Consumers

**Go** — import paths unchanged:

```go
import "github.com/emersonary/appkit/accounts"
```

Local replace (Sahar / monorepo):

```go
replace github.com/emersonary/appkit/accounts => ../../via-jeri/appkit/blocks/account/go
```

**npm**:

```json
"@emersonary/appkit-accounts": "file:../../via-jeri/appkit/blocks/account/web"
```

See [`web/README.md`](web/README.md) and [`web/BINDING.md`](web/BINDING.md).
