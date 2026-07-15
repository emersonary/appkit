"use client";

import { useEffect, useId, useLayoutEffect, useRef, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

export type ResourceListRowMenuItem = {
  id: string;
  label: string;
  onSelect: () => void;
  destructive?: boolean;
  disabled?: boolean;
  icon?: ReactNode;
};

type ResourceListRowMenuProps = {
  items: ResourceListRowMenuItem[];
  ariaLabel?: string;
  /** Called when the menu opens or closes (e.g. to select the row on open). */
  onOpenChange?: (open: boolean) => void;
};

type MenuPosition = {
  top: number;
  left: number;
};

const MENU_WIDTH = 176;

export function ResourceListRowMenu({
  items,
  ariaLabel = "Row actions",
  onOpenChange,
}: ResourceListRowMenuProps) {
  const menuId = useId();
  const rootRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);
  const [position, setPosition] = useState<MenuPosition | null>(null);

  function setMenuOpen(next: boolean) {
    setOpen(next);
    onOpenChange?.(next);
  }

  useLayoutEffect(() => {
    if (!open || !buttonRef.current) {
      return;
    }

    function updatePosition() {
      const button = buttonRef.current;
      if (!button) {
        return;
      }
      const rect = button.getBoundingClientRect();
      const menuHeight = menuRef.current?.offsetHeight ?? items.length * 40 + 8;
      const viewportPadding = 8;
      let top = rect.bottom + 4;
      let left = rect.right - MENU_WIDTH;

      if (left < viewportPadding) {
        left = viewportPadding;
      }
      if (left + MENU_WIDTH > window.innerWidth - viewportPadding) {
        left = window.innerWidth - MENU_WIDTH - viewportPadding;
      }
      if (top + menuHeight > window.innerHeight - viewportPadding) {
        top = rect.top - menuHeight - 4;
      }

      setPosition({ top, left });
    }

    updatePosition();
    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);
    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [open, items.length]);

  useEffect(() => {
    if (!open) {
      return;
    }
    function handlePointerDown(event: MouseEvent) {
      const target = event.target as Node;
      if (rootRef.current?.contains(target) || menuRef.current?.contains(target)) {
        return;
      }
      setMenuOpen(false);
    }
    function handleEscape(event: KeyboardEvent) {
      if (event.key === "Escape") {
        setMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [open]);

  return (
    <div ref={rootRef} className="appkit-resource-edit-list__menu-root">
      <button
        ref={buttonRef}
        type="button"
        aria-label={ariaLabel}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        className="appkit-resource-edit-list__menu-trigger"
        onClick={() => setMenuOpen(!open)}
      >
        <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden className="appkit-resource-edit-list__menu-icon">
          <circle cx="5" cy="12" r="1.75" />
          <circle cx="12" cy="12" r="1.75" />
          <circle cx="19" cy="12" r="1.75" />
        </svg>
      </button>
      {open && position
        ? createPortal(
            <div
              ref={menuRef}
              id={menuId}
              role="menu"
              style={{ top: position.top, left: position.left, width: MENU_WIDTH }}
              className="appkit-resource-edit-list__menu"
            >
              {items.map((item) => (
                <button
                  key={item.id}
                  type="button"
                  role="menuitem"
                  disabled={item.disabled}
                  className={
                    item.destructive
                      ? "appkit-resource-edit-list__menu-item appkit-resource-edit-list__menu-item--danger"
                      : "appkit-resource-edit-list__menu-item"
                  }
                  onClick={() => {
                    setMenuOpen(false);
                    item.onSelect();
                  }}
                >
                  {item.icon ? (
                    <span className="appkit-resource-edit-list__menu-item-icon" aria-hidden>
                      {item.icon}
                    </span>
                  ) : null}
                  <span>{item.label}</span>
                </button>
              ))}
            </div>,
            document.body,
          )
        : null}
    </div>
  );
}

export function ResourceListMenuIcon({ name }: { name: "edit" | "copy" | "delete" | "audit" }) {
  switch (name) {
    case "edit":
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <path d="M12 20h9" />
          <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
        </svg>
      );
    case "copy":
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <rect x="9" y="9" width="13" height="13" rx="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      );
    case "delete":
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <path d="M3 6h18" />
          <path d="M8 6V4h8v2" />
          <path d="M19 6l-1 14H6L5 6" />
          <path d="M10 11v6" />
          <path d="M14 11v6" />
        </svg>
      );
    case "audit":
      return (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden>
          <path d="M8 6h13" />
          <path d="M8 12h13" />
          <path d="M8 18h13" />
          <path d="M3 6h.01" />
          <path d="M3 12h.01" />
          <path d="M3 18h.01" />
        </svg>
      );
  }
}
