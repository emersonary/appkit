"use client";

import { useCallback, useMemo } from "react";
import type { ResourceViewListHandlers } from "./handlers";
import type { ResourceListItem } from "./types";

export type ListConfirmDeleteOptions = {
  /** Values key used for the row label in the confirm dialog. Default: `name`. */
  nameKey?: string;
  /** Label when the name field is empty. Default: `this item`. */
  fallbackLabel?: string;
};

export type RepairListSelectionOptions = {
  items: ResourceListItem[];
  selectedId?: string | null;
  onSelectItem?: (itemId: string) => void;
  onClearSelection?: () => void;
};

export type ResourceViewListHandlersOptions = {
  items: ResourceListItem[];
  reload: () => Promise<ResourceListItem[]>;
  selectedId?: string | null;
  onSelectItem?: (itemId: string) => void;
  onClearSelection?: () => void;
  confirmDelete?: ListConfirmDeleteOptions;
};

/** Build a delete-confirm handler from the current list rows. */
export function confirmDeleteFromListItems(
  items: ResourceListItem[],
  itemId: string,
  options: ListConfirmDeleteOptions = {},
): boolean {
  const nameKey = options.nameKey ?? "name";
  const fallbackLabel = options.fallbackLabel ?? "this item";
  const item = items.find((entry) => entry.id === itemId);
  const name = item?.values[nameKey]?.trim() || fallbackLabel;
  return window.confirm(`Delete "${name}"? This cannot be undone.`);
}

/** After delete, keep URL/list selection valid when the selected row is gone. */
export function repairListSelectionAfterDelete({
  items,
  selectedId,
  onSelectItem,
  onClearSelection,
}: RepairListSelectionOptions): void {
  if (!selectedId || items.some((item) => item.id === selectedId)) {
    return;
  }
  const nextId = items[0]?.id;
  if (nextId && onSelectItem) {
    onSelectItem(nextId);
    return;
  }
  onClearSelection?.();
}

/** Reload list after a successful save. */
export function createReloadOnSaved(reload: () => void | Promise<void>): () => void {
  return () => {
    void reload();
  };
}

/** Reload list and repair selection after a successful delete. */
export function createReloadAndRepairOnDeleted(
  reload: () => Promise<ResourceListItem[]>,
  selection: Omit<RepairListSelectionOptions, "items">,
): () => Promise<void> {
  return async () => {
    const nextItems = await reload();
    repairListSelectionAfterDelete({ items: nextItems, ...selection });
  };
}

/**
 * Standard list lifecycle handlers for `listHandlers`: confirm delete, reload on save,
 * reload + selection repair on delete.
 */
export function useResourceViewListHandlers({
  items,
  reload,
  selectedId,
  onSelectItem,
  onClearSelection,
  confirmDelete: confirmDeleteOptions,
}: ResourceViewListHandlersOptions): ResourceViewListHandlers {
  const nameKey = confirmDeleteOptions?.nameKey ?? "name";
  const fallbackLabel = confirmDeleteOptions?.fallbackLabel ?? "this item";

  const confirmDelete = useCallback(
    (itemId: string) => confirmDeleteFromListItems(items, itemId, { nameKey, fallbackLabel }),
    [fallbackLabel, items, nameKey],
  );

  const onSaved = useCallback(() => {
    void reload();
  }, [reload]);

  const onDeleted = useCallback(async () => {
    const nextItems = await reload();
    repairListSelectionAfterDelete({
      items: nextItems,
      selectedId,
      onSelectItem,
      onClearSelection,
    });
  }, [onClearSelection, onSelectItem, reload, selectedId]);

  return useMemo(
    () => ({
      onSaved,
      confirmDelete,
      onDeleted,
    }),
    [confirmDelete, onDeleted, onSaved],
  );
}
