import type { Menu, MenuItem } from "./gen/v1/menu_pb";

export function normalizeMenuPath(path: string): string {
  const trimmed = path.trim();
  if (!trimmed || trimmed === "#") {
    return "";
  }
  return trimmed.endsWith("/") && trimmed.length > 1 ? trimmed.slice(0, -1) : trimmed;
}

export function collectMenuRoutes(items: MenuItem[], routes: string[] = []): string[] {
  for (const item of items) {
    const route = item.routeName?.trim();
    if (route) {
      routes.push(route);
    }
    collectMenuRoutes(item.children, routes);
  }
  return routes;
}

export function collectAllMenuRoutes(menus: Menu[]): string[] {
  const routes: string[] = [];
  for (const menu of menus) {
    collectMenuRoutes(menu.items, routes);
  }
  return routes;
}

export function flattenMenuItems(
  items: MenuItem[],
  menuId: string,
  out: Array<{ menuId: string; item: MenuItem }> = [],
): Array<{ menuId: string; item: MenuItem }> {
  for (const item of items) {
    out.push({ menuId, item });
    flattenMenuItems(item.children, menuId, out);
  }
  return out;
}

/** True when pathname belongs to this menu route and no other menu route is a better match. */
export function matchesMenuRoute(
  routeName: string,
  pathname: string,
  allRoutes: string[],
): boolean {
  const route = normalizeMenuPath(routeName);
  const path = normalizeMenuPath(pathname);
  if (!route || !path) {
    return false;
  }
  if (path !== route && !path.startsWith(`${route}/`)) {
    return false;
  }

  for (const other of allRoutes) {
    const candidate = normalizeMenuPath(other);
    if (!candidate || candidate.length <= route.length || candidate === route) {
      continue;
    }
    if (path === candidate || path.startsWith(`${candidate}/`)) {
      return false;
    }
  }

  return true;
}

/** First navigable route in menu item order (depth-first). */
export function firstMenuItemRoute(items: MenuItem[]): string {
  for (const item of items) {
    const route = normalizeMenuPath(item.routeName);
    if (route) {
      return route;
    }
    const nested = firstMenuItemRoute(item.children);
    if (nested) {
      return nested;
    }
  }
  return "";
}

export function resolveActiveMenuItem(
  pathname: string,
  menus: Menu[],
): { menuId: string; permissionId: string } | null {
  const allRoutes = collectAllMenuRoutes(menus);
  const candidates: Array<{ menuId: string; permissionId: string; routeLen: number }> = [];

  for (const menu of menus) {
    for (const { menuId, item } of flattenMenuItems(menu.items, menu.id)) {
      if (!matchesMenuRoute(item.routeName, pathname, allRoutes)) {
        continue;
      }
      candidates.push({
        menuId,
        permissionId: item.permissionId,
        routeLen: normalizeMenuPath(item.routeName).length,
      });
    }
  }

  if (candidates.length === 0) {
    return null;
  }

  candidates.sort((a, b) => b.routeLen - a.routeLen);
  return candidates[0];
}

export function itemContainsSelectedDescendant(
  item: MenuItem,
  selectedPermissionId: string,
): boolean {
  if (!selectedPermissionId) {
    return false;
  }
  return item.children.some(
    (child) =>
      child.permissionId === selectedPermissionId ||
      itemContainsSelectedDescendant(child, selectedPermissionId),
  );
}
