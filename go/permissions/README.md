# Permissions (`appkit/go/permissions`)

Setup-driven RBAC: profiles, permission groups (FE sessions), categories (sidebar sections), and a permission tree with `be_action` bitmasks.

## Enable

```yaml
permissions:
  enabled: true
  schema: account
  default_profile: member
  groups: [...]
  categories: [...]
  permissions: [...]
  profiles: [...]
  profile_permissions: [...]
```

See `permissions.example.yaml` for a full seed. Load standalone files with `LoadSetup`; inline YAML uses `SetupInput.Resolve()` or `ResolveSetup`.

## Action bits

| Bit | Value |
|-----|-------|
| List | 1 |
| Create | 2 |
| Update | 4 |
| Delete | 8 |

## Service (no HTTP routes in this package)

- `HasPermission(ctx, accountID, idPermission, actionBit)`
- `ListFlatForAccount(ctx, accountID)` — for consumers that build FE trees elsewhere
- `ListCatalog(ctx)` — full catalog (groups, categories, permissions)
- `AssignAccountProfile`, `AssignNewAccountProfile`

When wired with `accounts`, new registrations receive `id_profile` (`admin` when `register_as_admin`).

## Tables

- `permission_groups` — top FE session (Manage, Settings, …)
- `permission_categories` — section inside a group
- `permissions` — tree (`id_parent`), `be_action`, `route_name`, `icon` (`id_permission` must not contain `.`; tree paths are derived from parent links)
- `profiles`, `profile_permissions`
- `accounts.id_profile` (FK added when permissions schema applies)
