import { useEffect, useMemo, useState, type ReactNode } from "react";
import type { GetMenuResponse, Menu, MenuItem } from "./gen/v1/menu_pb";
import { itemContainsSelectedDescendant } from "./routes";

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
  renderMenuIcon?: (menu: Menu) => ReactNode;
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

function itemHasActiveDescendant(item: MenuItem, selectedPermissionId: string): boolean {
  return itemContainsSelectedDescendant(item, selectedPermissionId);
}

type MenuSection = {
  groupName: string;
  items: MenuItem[];
};

function readGroupName(item: MenuItem): string {
  const direct = item.groupName?.trim();
  if (direct) {
    return direct;
  }
  const legacy = (item as MenuItem & { group_name?: string }).group_name?.trim();
  return legacy ?? "";
}

function groupMenuItems(items: MenuItem[]): MenuSection[] {
  const sections: MenuSection[] = [];
  for (const item of items) {
    const groupName = readGroupName(item);
    const last = sections[sections.length - 1];
    if (last && last.groupName === groupName) {
      last.items.push(item);
    } else {
      sections.push({ groupName, items: [item] });
    }
  }
  return sections;
}

function MenuTree({
  items,
  selectedPermissionId,
  expandedIds,
  onToggle,
  onSelect,
  renderIcon,
  linkComponent = defaultLink,
  depth = 0,
}: {
  items: MenuItem[];
  selectedPermissionId: string;
  expandedIds: Set<string>;
  onToggle: (fullId: string) => void;
  onSelect: (item: MenuItem) => void;
  renderIcon?: (item: MenuItem) => ReactNode;
  linkComponent?: SidebarProps["linkComponent"];
  depth?: number;
}) {
  return (
    <ul className={`appkit-menu__tree${depth > 0 ? " appkit-menu__tree--nested" : ""}`}>
      {items.map((item) => {
        const hasChildren = item.children.length > 0;
        const selected =
          Boolean(selectedPermissionId) && item.permissionId === selectedPermissionId;
        const expanded =
          expandedIds.has(item.fullId) || itemHasActiveDescendant(item, selectedPermissionId);
        const href = item.routeName || "#";

        return (
          <li key={item.fullId} className="appkit-menu__tree-item">
            <div className="appkit-menu__row">
              {linkComponent({
                href,
                className: selected ? "appkit-menu__link is-selected" : "appkit-menu__link",
                onClick: () => onSelect(item),
                children: (
                  <>
                    <span className="appkit-menu__icon" aria-hidden="true">
                      {renderIcon?.(item)}
                    </span>
                    <span className="appkit-menu__label">{item.name}</span>
                  </>
                ),
              })}
              {hasChildren ? (
                <button
                  type="button"
                  className={`appkit-menu__chevron${expanded ? " is-expanded" : ""}`}
                  aria-expanded={expanded}
                  aria-label={expanded ? `Collapse ${item.name}` : `Expand ${item.name}`}
                  onClick={() => onToggle(item.fullId)}
                />
              ) : null}
            </div>
            {hasChildren && expanded ? (
              <MenuTree
                items={item.children}
                selectedPermissionId={selectedPermissionId}
                expandedIds={expandedIds}
                onToggle={onToggle}
                onSelect={onSelect}
                renderIcon={renderIcon}
                linkComponent={linkComponent}
                depth={depth + 1}
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
  renderMenuIcon,
  linkComponent,
}: SidebarProps) {
  const sidebar = layout.sidebar;
  const menus = layout.menus;
  const currentMenuId = activeMenuId ?? menus[0]?.id ?? "";
  const currentMenu = menus.find((menu) => menu.id === currentMenuId) ?? menus[0];
  const selected = selectedPermissionId ?? layout.defaultSelectedPermissionId ?? "";
  const [expandedIds, setExpandedIds] = useState<Set<string>>(() => new Set());

  const sections = useMemo(
    () => groupMenuItems(currentMenu?.items ?? []),
    [currentMenu?.items],
  );

  useEffect(() => {
    if (!currentMenu || !selected) {
      return;
    }
    const next = new Set<string>();
    const visit = (items: MenuItem[]) => {
      for (const item of items) {
        if (item.children.length > 0 && itemHasActiveDescendant(item, selected)) {
          next.add(item.fullId);
        }
        visit(item.children);
      }
    };
    visit(currentMenu.items);
    setExpandedIds(next);
  }, [currentMenu, selected]);

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

  const toggleExpanded = (fullId: string) => {
    setExpandedIds((current) => {
      const next = new Set(current);
      if (next.has(fullId)) {
        next.delete(fullId);
      } else {
        next.add(fullId);
      }
      return next;
    });
  };

  return (
    <aside className={rootClass} aria-label="Application navigation">
      <div className="appkit-menu__panel">
        {currentMenu ? (
          <div className="appkit-menu__content">
            {menus.length <= 1 ? (
              <div className="appkit-menu__title">{currentMenu.name}</div>
            ) : null}
            {sections.map((section, index) => (
              <section
                key={`${section.groupName || "ungrouped"}-${index}`}
                className={
                  index > 0 ? "appkit-menu__section appkit-menu__section--bordered" : "appkit-menu__section"
                }
              >
                {section.groupName ? (
                  <h3 className="appkit-menu__section-title">{section.groupName}</h3>
                ) : null}
                <MenuTree
                  items={section.items}
                  selectedPermissionId={selected}
                  expandedIds={expandedIds}
                  onToggle={toggleExpanded}
                  onSelect={handleSelect}
                  renderIcon={renderIcon}
                  linkComponent={linkComponent}
                />
              </section>
            ))}
          </div>
        ) : null}

        {menus.length > 1 ? (
          <nav className="appkit-menu__dock" aria-label="Main sections">
            {menus.map((menu) => {
              const active = menu.id === currentMenuId;
              return (
                <button
                  key={menu.id}
                  type="button"
                  className={active ? "appkit-menu__dock-btn is-active" : "appkit-menu__dock-btn"}
                  title={menu.name}
                  aria-label={menu.name}
                  aria-current={active ? "page" : undefined}
                  onClick={() => onMenuChange?.(menu.id)}
                >
                  <span className="appkit-menu__dock-icon" aria-hidden="true">
                    {renderMenuIcon?.(menu)}
                  </span>
                </button>
              );
            })}
          </nav>
        ) : null}
      </div>
    </aside>
  );
}
