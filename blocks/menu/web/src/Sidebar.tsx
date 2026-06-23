import type { ReactNode } from "react";
import type { GetMenuResponse, Menu, MenuItem } from "./gen/v1/menu_pb";

export type SidebarProps = {
  layout: GetMenuResponse;
  activeMenuId?: string;
  selectedPermissionId?: string;
  collapsed?: boolean;
  onSelectItem?: (item: MenuItem, menu: Menu) => void;
  onMenuChange?: (menuId: string) => void;
  onCollapsedChange?: (collapsed: boolean) => void;
  className?: string;
  renderIcon?: (item: MenuItem) => ReactNode;
  linkComponent?: (props: {
    href: string;
    children: ReactNode;
    className?: string;
    onClick?: () => void;
  }) => ReactNode;
};

function defaultLink(props: {
  href: string;
  children: ReactNode;
  className?: string;
  onClick?: () => void;
}) {
  return (
    <a href={props.href} className={props.className} onClick={props.onClick}>
      {props.children}
    </a>
  );
}

function MenuTree({
  items,
  selectedPermissionId,
  onSelect,
  renderIcon,
  linkComponent = defaultLink,
}: {
  items: MenuItem[];
  selectedPermissionId?: string;
  onSelect: (item: MenuItem) => void;
  renderIcon?: (item: MenuItem) => ReactNode;
  linkComponent?: SidebarProps["linkComponent"];
}) {
  return (
    <ul className="appkit-menu__tree">
      {items.map((item) => {
        const selected = item.permissionId === selectedPermissionId;
        const href = item.routeName || "#";
        return (
          <li key={item.fullId} className="appkit-menu__tree-item">
            {linkComponent({
              href,
              className: selected ? "appkit-menu__link is-selected" : "appkit-menu__link",
              onClick: () => onSelect(item),
              children: (
                <>
                  {renderIcon?.(item)}
                  <span>{item.name}</span>
                </>
              ),
            })}
            {item.children.length > 0 ? (
              <MenuTree
                items={item.children}
                selectedPermissionId={selectedPermissionId}
                onSelect={onSelect}
                renderIcon={renderIcon}
                linkComponent={linkComponent}
              />
            ) : null}
          </li>
        );
      })}
    </ul>
  );
}

export function Sidebar({
  layout,
  activeMenuId,
  selectedPermissionId,
  collapsed,
  onSelectItem,
  onMenuChange,
  onCollapsedChange,
  className,
  renderIcon,
  linkComponent,
}: SidebarProps) {
  const sidebar = layout.sidebar;
  const menus = layout.menus;
  const currentMenuId = activeMenuId ?? menus[0]?.id ?? "";
  const currentMenu = menus.find((menu) => menu.id === currentMenuId) ?? menus[0];
  const selected =
    selectedPermissionId ?? layout.defaultSelectedPermissionId ?? "";

  const rootClass = [
    "appkit-menu",
    className,
    sidebar?.floating ? "appkit-menu--floating" : "",
    sidebar?.locked ? "appkit-menu--locked" : "",
    collapsed ? "appkit-menu--collapsed" : "",
  ]
    .filter(Boolean)
    .join(" ");

  const handleSelect = (item: MenuItem) => {
    if (currentMenu) {
      onSelectItem?.(item, currentMenu);
    }
    if (sidebar?.hideWhenSelected && !sidebar.locked) {
      onCollapsedChange?.(true);
    }
  };

  return (
    <aside className={rootClass} aria-label="Application navigation">
      <div className="appkit-menu__panel">
        {currentMenu ? (
          <div className="appkit-menu__content">
            <div className="appkit-menu__title">{currentMenu.name}</div>
            <MenuTree
              items={currentMenu.items}
              selectedPermissionId={selected}
              onSelect={handleSelect}
              renderIcon={renderIcon}
              linkComponent={linkComponent}
            />
          </div>
        ) : null}

        {menus.length > 1 ? (
          <nav className="appkit-menu__tabs" aria-label="Menu sections">
            {menus.map((menu) => (
              <button
                key={menu.id}
                type="button"
                className={
                  menu.id === currentMenuId
                    ? "appkit-menu__tab is-active"
                    : "appkit-menu__tab"
                }
                onClick={() => onMenuChange?.(menu.id)}
              >
                {menu.name}
              </button>
            ))}
          </nav>
        ) : null}
      </div>
    </aside>
  );
}
