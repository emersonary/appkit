"use client";

import { useCallback, type MouseEvent, type ReactNode } from "react";
import { useShellLoading } from "./shell-loading";

export type NavigationLinkRenderProps = {
  href: string;
  children: ReactNode;
  className?: string;
  onClick?: (event: MouseEvent<HTMLElement>) => void;
};

export function matchesAppRoute(routeName: string, pathname: string, search: string): boolean {
  const trimmed = routeName.trim();
  if (!trimmed || trimmed === "#") {
    return false;
  }

  const [routePath, routeQuery] = trimmed.split("?");
  const path = pathname.endsWith("/") && pathname.length > 1 ? pathname.slice(0, -1) : pathname;

  if (path !== routePath && !path.startsWith(`${routePath}/`)) {
    return false;
  }

  if (!routeQuery) {
    return true;
  }

  const normalizedSearch = search.startsWith("?") ? search.slice(1) : search;
  return normalizedSearch.includes(routeQuery);
}

export function useNavigationLinkComponent(
  renderLink: (props: NavigationLinkRenderProps) => ReactNode,
  pathname: string,
  search: string,
) {
  const { startNavigation } = useShellLoading();

  return useCallback(
    (props: NavigationLinkRenderProps) => {
      const { href, children, className, onClick } = props;
      return renderLink({
        href,
        children,
        className,
        onClick: (event) => {
          if (!matchesAppRoute(href, pathname, search)) {
            startNavigation();
          }
          onClick?.(event);
        },
      });
    },
    [pathname, renderLink, search, startNavigation],
  );
}
