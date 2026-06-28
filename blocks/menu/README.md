# Menu block

Backend-defined navigation menus filtered by account permissions, with a React sidebar and Connect RPC.

## Backend

- `menu.Setup` — multiple named menus, each listing permission **id** roots
- `menu.Service.GetMenu` — resolves trees, expands descendants, filters by grants
- `MenuService.GetMenu` — Connect + gRPC transport

Menus require the permissions block to be enabled.

## Config

```yaml
permissions:
  enabled: true
  config_path: config/permissions.yaml

menu:
  enabled: true
  config_path: config/menu.yaml
```

See `menu.example.yaml`.

## UI

```tsx
import { MenuProvider, ConnectedSidebar } from "@emersonary/appkit-menu";
import "@emersonary/appkit-menu/sidebar.css";

<MenuProvider baseUrl="/api" getAccessToken={() => token}>
  <ConnectedSidebar />
</MenuProvider>
```

Sidebar props come from backend `sidebar` config:

- `floating`
- `hide_when_selected`
- `locked`
- `default_menu` — permission **id** for initial selection

Menu tabs render at the **bottom** of the sidebar when multiple menus are configured.

## Codegen

```bash
# Go
protoc -I proto \
  --go_out=go --go_opt=module=github.com/emersonary/appkit/menu \
  --go-grpc_out=go --go-grpc_opt=module=github.com/emersonary/appkit/menu \
  --connect-go_out=go --connect-go_opt=module=github.com/emersonary/appkit/menu \
  proto/v1/menu.proto

# TypeScript
cd web && npm run generate:menu-proto
```

## Consumer transport

```go
menutransport.New(rt.Menu, &menutransport.Mount{
  HTTPMux: mux,
  GRPCServer: grpcSrv,
  ResolveAccountID: func(ctx context.Context, token string) (string, error) {
    session, err := rt.Accounts.SessionFromToken(ctx, token)
    if err != nil {
      return "", err
    }
    return session.Account.ID, nil
  },
})
```
