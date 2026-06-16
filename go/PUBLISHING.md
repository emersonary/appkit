# Publishing appkit

Repository: **[github.com/emersonary/appkit](https://github.com/emersonary/appkit)**

The Go module path is **`github.com/emersonary/appkit`**. It must match the Git remote URL.

## 1. Push this module

From the repository root (`appkit/`) or this module directory (`go/`):

```bash
git add .
git commit -m "Initial appkit module"
git branch -M main
git remote add origin git@github.com:emersonary/appkit.git
git push -u origin main
```

HTTPS alternative:

```bash
git remote add origin https://github.com/emersonary/appkit.git
git push -u origin main
```

## 2. Tag a release

Go modules consume **version tags**, not branches:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Use [semantic versioning](https://go.dev/doc/modules/version-numbers): `v0.1.0`, `v0.1.1`, `v0.2.0`, etc.

## 3. Configure your machine

If the repository is private, tell Go not to use the public proxy:

```bash
go env -w GOPRIVATE=github.com/emersonary/appkit
```

For all repos under the org:

```bash
go env -w GOPRIVATE=github.com/emersonary/*
```

Ensure Git can authenticate to GitHub (SSH key or credential helper):

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

## 4. Consume from another project

In the consumer's `go.mod`:

```go
require github.com/emersonary/appkit v0.1.0
```

Remove any local `replace` directive:

```go
// delete after publishing:
// replace github.com/emersonary/appkit => ../appkit/go/go
```

Then:

```bash
go get github.com/emersonary/appkit@v0.1.0
go mod tidy
```

Import packages as usual:

```go
import (
    "github.com/emersonary/appkit/apperror"
    "github.com/emersonary/appkit/currency"
    "github.com/emersonary/appkit/dbhist"
    "github.com/emersonary/appkit/heapedcache"
    "github.com/emersonary/appkit/log"
)
```

## 5. Update existing projects

### via-jeri (monorepo root)

This repository is the **source of truth** for appkit (Go + `@emersonary/appkit-accounts`), the posts service, and shared protos.

While developing locally, consumers in this repo use [`go.work`](../../go.work) and `replace` directives:

```go
require github.com/emersonary/appkit v0.0.0
replace github.com/emersonary/appkit => ../appkit/go
```

Posts API (`services/posts/api/go.mod`):

```go
replace github.com/emersonary/appkit => ../../../appkit/go
```

Accounts block (`blocks/account/go/go.mod`):

```go
replace github.com/emersonary/appkit => ../../../go
```

After publishing appkit tags, remove `replace` lines and `go get` the version in each module.

See [`docs/MONOREPO.md`](../../docs/MONOREPO.md) for layout and publishing services.

See [`../blocks/account/web/README.md`](../blocks/account/web/README.md) for the accounts TypeScript/React package (`@emersonary/appkit-accounts`).

### sahar (`api/go.mod`)

Same pattern — swap the local `replace` for a tagged version once `v0.1.0` is pushed.

## 6. Releasing updates

After changing `proto/auth/v1/auth.proto`, regenerate stubs from the **appkit repo root**:

```bash
make proto
# Windows without GNU make:
# protoc -I proto --go_out=go --go_opt=module=github.com/emersonary/appkit \
#   --go-grpc_out=go --go-grpc_opt=module=github.com/emersonary/appkit \
#   --connect-go_out=go --connect-go_opt=module=github.com/emersonary/appkit \
#   proto/auth/v1/auth.proto
# cd blocks/account/web && npm run generate:auth-proto
```

Commit generated `blocks/account/go/gen/` and `blocks/account/web/src/gen/` with the proto change, then tag.

```bash
git add .
git commit -m "Describe the change"
git tag v0.1.1
git push origin main
git push origin v0.1.1
```

In each consumer:

```bash
go get github.com/emersonary/appkit@v0.1.1
go mod tidy
```

## 7. CI example (GitHub Actions)

In a consumer repo under the same org:

```yaml
env:
  GOPRIVATE: github.com/emersonary/appkit

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: git config --global url."https://${{ secrets.GITHUB_TOKEN }}@github.com/".insteadOf "https://github.com/"
      - run: cd go && go test ./...
```

## Checklist

- [x] Repo [github.com/emersonary/appkit](https://github.com/emersonary/appkit) exists
- [ ] Code pushed to `main`
- [ ] Tag `v0.1.0` pushed
- [ ] `GOPRIVATE` set on dev machines and CI
- [ ] Consumer projects: remove `replace`, `go get` tagged version
