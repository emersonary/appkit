# Blog block

Shared blog post rendering and Connect RPC contract for the posts platform.

## Layout

| Path | Role |
|------|------|
| `proto/v1/blog.proto` | `BlogService` contract (public + admin RPCs) |
| `go/gen/` | Generated Go types and Connect handlers |
| `go/transport/` | Reusable Connect/gRPC mount helpers |
| `web/` | `BlogPostView`, client helpers, hooks |

**posts-api** implements `BlogService` (database, translations, media, social). Consumer sites and posts-admin import this block for UI and types.

## Backend (posts-api)

```go
import (
    blogtransport "github.com/emersonary/appkit/blog/transport"
    blogv1connect "github.com/emersonary/appkit/blog/gen/blog/v1/blogv1connect"
)

blogtransport.Register(
    myBlogHandler, // implements blogv1connect.BlogServiceHandler
    &blogtransport.Mount{
        HTTPMux: mux,
        ConnectOptions: []connect.HandlerOption{...},
    },
)
```

## Web

```tsx
import { BlogPostView, usePublishedPost } from "@emersonary/appkit-blog";
import "@emersonary/appkit-blog/blog.css";

const { post, loading, error } = usePublishedPost({
  baseUrl: "",
  projectId: "sahar",
  slug: "my-post",
  language: "pt",
});
```

Set `data-theme` on `<html>` in each consumer site for brand colors.

## Codegen

```bash
# Go
cd blocks/blog
protoc -I proto \
  --go_out=go --go_opt=module=github.com/emersonary/appkit/blog \
  --go-grpc_out=go --go-grpc_opt=module=github.com/emersonary/appkit/blog \
  --connect-go_out=go --connect-go_opt=module=github.com/emersonary/appkit/blog \
  proto/v1/blog.proto

# TypeScript
cd web && npm install && npm run generate:blog-proto
```
