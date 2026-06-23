import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { MenuService } from "./gen/v1/menu_connect";
import type { GetMenuResponse } from "./gen/v1/menu_pb";
import { Sidebar, type SidebarProps } from "./Sidebar";

type MenuClient = ReturnType<typeof createClient<typeof MenuService>>;

type MenuProviderProps = {
  baseUrl: string;
  getAccessToken: () => string | null | Promise<string | null>;
  children: ReactNode;
  initialMenuId?: string;
};

type MenuContextValue = {
  layout: GetMenuResponse | null;
  loading: boolean;
  error: Error | null;
  activeMenuId: string;
  selectedPermissionId: string;
  collapsed: boolean;
  setActiveMenuId: (menuId: string) => void;
  setSelectedPermissionId: (permissionId: string) => void;
  setCollapsed: (collapsed: boolean) => void;
  refresh: () => Promise<void>;
  sidebarProps: Omit<SidebarProps, "layout">;
};

const MenuContext = createContext<MenuContextValue | null>(null);

export function MenuProvider({
  baseUrl,
  getAccessToken,
  children,
  initialMenuId,
}: MenuProviderProps) {
  const client = useMemo<MenuClient>(() => {
    const transport = createConnectTransport({
      baseUrl,
      interceptors: [
        (next) => async (req) => {
          const token = await getAccessToken();
          if (token) {
            req.header.set("Authorization", `Bearer ${token}`);
          }
          return next(req);
        },
      ],
    });
    return createClient(MenuService, transport);
  }, [baseUrl, getAccessToken]);

  const [layout, setLayout] = useState<GetMenuResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [activeMenuId, setActiveMenuId] = useState(initialMenuId ?? "");
  const [selectedPermissionId, setSelectedPermissionId] = useState("");
  const [collapsed, setCollapsed] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await client.getMenu({});
      setLayout(response);
      setActiveMenuId((current) => current || initialMenuId || response.menus[0]?.id || "");
      setSelectedPermissionId(
        (current) => current || response.defaultSelectedPermissionId || "",
      );
    } catch (err) {
      setError(err instanceof Error ? err : new Error("failed to load menu layout"));
    } finally {
      setLoading(false);
    }
  }, [client, initialMenuId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const value = useMemo<MenuContextValue>(
    () => ({
      layout,
      loading,
      error,
      activeMenuId,
      selectedPermissionId,
      collapsed,
      setActiveMenuId,
      setSelectedPermissionId,
      setCollapsed,
      refresh,
      sidebarProps: {
        activeMenuId,
        selectedPermissionId,
        collapsed,
        onMenuChange: setActiveMenuId,
        onCollapsedChange: setCollapsed,
        onSelectItem: (item) => setSelectedPermissionId(item.permissionId),
      },
    }),
    [
      layout,
      loading,
      error,
      activeMenuId,
      selectedPermissionId,
      collapsed,
      refresh,
    ],
  );

  return <MenuContext.Provider value={value}>{children}</MenuContext.Provider>;
}

export function useMenu() {
  const ctx = useContext(MenuContext);
  if (!ctx) {
    throw new Error("useMenu must be used within MenuProvider");
  }
  return ctx;
}

export function ConnectedSidebar(props: Omit<SidebarProps, "layout">) {
  const { layout, loading, error, sidebarProps } = useMenu();
  if (loading) {
    return <aside className="appkit-menu appkit-menu--loading">Loading menu…</aside>;
  }
  if (error || !layout) {
    return <aside className="appkit-menu appkit-menu--error">Menu unavailable</aside>;
  }
  return <Sidebar layout={layout} {...sidebarProps} {...props} />;
}
