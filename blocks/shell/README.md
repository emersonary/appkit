# Shell block (web)

Coordinates **main-panel loading** when navigating via AppKit menu links and while page content fetches data.

## Package

`@emersonary/appkit-shell` — `blocks/shell/web`

## Consumer wiring (Next.js)

1. Add dependency: `"@emersonary/appkit-shell": "file:../../.deps/appkit/blocks/shell/web"`
2. Import `shell.css` or map `.appkit-shell-*` classes in your theme.
3. Wrap the app (inside `Suspense` if using `useSearchParams`):

```tsx
import { ShellLoadingProvider } from "@emersonary/appkit-shell";

<ShellLoadingProvider>{children}</ShellLoadingProvider>
```

4. Main content area:

```tsx
import { ShellContentPanel } from "@emersonary/appkit-shell";

<ShellContentPanel>{children}</ShellContentPanel>
```

5. Menu `linkComponent` (with `@emersonary/appkit-menu`):

```tsx
import { useNavigationLinkComponent } from "@emersonary/appkit-shell";

const linkComponent = useNavigationLinkComponent(
  ({ href, children, className, onClick }) => (
    <Link href={href} className={className} onClick={onClick}>{children}</Link>
  ),
  pathname,
  search,
);
```

6. Pages: `useSyncShellLoading(loading)` or `useShellReady()`.
