# Publishing appkit

Repository: **[github.com/emersonary/appkit](https://github.com/emersonary/appkit)**

The Go module path is **`github.com/emersonary/appkit`**. It must match the Git remote URL.

## 1. Push this module

From the `appkit` directory:

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
// replace github.com/emersonary/appkit => ../appkit
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

### via-jeri (`backend/go.mod`)

While developing locally:

```go
require github.com/emersonary/appkit v0.0.0
replace github.com/emersonary/appkit => ../appkit
```

After publishing:

```go
require github.com/emersonary/appkit v0.1.0
```

Remove the `replace` line.

### sahar (`api/go.mod`)

Same pattern — swap the local `replace` for a tagged version once `v0.1.0` is pushed.

## 6. Releasing updates

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
      - run: go test ./...
```

## Checklist

- [x] Repo [github.com/emersonary/appkit](https://github.com/emersonary/appkit) exists
- [ ] Code pushed to `main`
- [ ] Tag `v0.1.0` pushed
- [ ] `GOPRIVATE` set on dev machines and CI
- [ ] Consumer projects: remove `replace`, `go get` tagged version
