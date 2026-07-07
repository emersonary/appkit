# @emersonary/appkit-shell

React shell loading for AppKit sidebar navigation and main content panels (Next.js App Router).

## What's included

- **`ShellLoadingProvider`** — sets loading on route changes; coordinates menu navigation with page readiness
- **`useShellLoading`** — read `contentLoading`, call `startNavigation()` / `setContentLoading()`
- **`useSyncShellLoading(loading)`** — mirror a page's async work into the shell overlay
- **`useShellReady()`** — static pages dismiss loading on mount
- **`ShellContentPanel`** — overlay + hidden content while loading
- **`useNavigationLinkComponent`** — wire menu links to `startNavigation()` when leaving the active route
- **`shell.css`** — base layout classes (theme colors come from the host app)

## Bind to a consumer app

### 1. Add the dependency

```json
{
  "dependencies": {
    "@emersonary/appkit-shell": "file:../../.deps/appkit/blocks/shell/web"
  }
}
```

### 2. Next.js transpile + alias

Add `@emersonary/appkit-shell` to `transpilePackages` and resolve aliases the same way as `@emersonary/appkit-menu`.

### 3. Wrap the app

```tsx
import { ShellLoadingProvider } from "@emersonary/appkit-shell";

export function AppProviders({ children }: { children: React.ReactNode }) {
  return (
    <Suspense fallback={null}>
      <ShellLoadingProvider>{children}</ShellLoadingProvider>
    </Suspense>
  );
}
```

### 4. Main content panel

```tsx
import { ShellContentPanel } from "@emersonary/appkit-shell";

<ShellContentPanel className="relative flex min-h-0 flex-1 flex-col overflow-hidden">
  {children}
</ShellContentPanel>
```

### 5. Menu navigation

```tsx
import { useNavigationLinkComponent } from "@emersonary/appkit-shell";

const linkComponent = useNavigationLinkComponent(
  ({ href, children, className, onClick }) => (
    <Link href={href} className={className} onClick={onClick}>
      {children}
    </Link>
  ),
  pathname,
  search,
);

<Sidebar linkComponent={linkComponent} {...sidebarProps} />
```

### 6. Page hooks

```tsx
// Data-fetching pages
useSyncShellLoading(loading);

// Static pages
useShellReady();
```

Import `@emersonary/appkit-shell/shell.css` and map `.appkit-shell-content-loading` colors to your design tokens if needed.
