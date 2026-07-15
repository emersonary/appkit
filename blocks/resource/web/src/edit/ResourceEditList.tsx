"use client";

import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import {
  isListCheckboxField,
  listCellBoolValue,
  listCellValue,
  listColumnsFromFields,
} from "./list-fields";
import { isResourceListInteractiveTarget, resourceListRowClassName } from "./resource-list";
import { ResourceListBoolCell } from "./ResourceListBoolCell";
import {
  ResourceListMenuIcon,
  ResourceListRowMenu,
  type ResourceListRowMenuItem,
} from "./ResourceListRowMenu";
import type { ResourceEditListHelpers } from "./ResourceListAndEdit";
import type { ResourceViewListHandlers } from "./handlers";
import {
  resolveResourceListAndEditCopy,
  type ResourceListAndEditDescriptions,
} from "./resource-descriptions";
import type { ResourceListItem, ResourceSchema } from "./types";

export type ResourceEditListRowMenuContext = {
  item: ResourceListItem;
  helpers: ResourceEditListHelpers;
  pending: boolean;
  deleteItem: (itemId: string) => void;
};

export type ResourceEditListProps = {
  schema: ResourceSchema;
  items: ResourceListItem[];
  helpers: ResourceEditListHelpers;
  handlers?: ResourceViewListHandlers;
  actions?: { replicate?: boolean; delete?: boolean };
  loading?: boolean;
  error?: string | null;
  /** Resource naming + optional copy overrides (preferred over flat description props). */
  descriptions?: ResourceListAndEditDescriptions;
  description?: ReactNode;
  emptyTitle?: string;
  emptyDescription?: ReactNode;
  selectedId?: string | null;
  onSelectItem?: (itemId: string) => void;
  showEditButton?: boolean;
  showRowMenu?: boolean;
  renderRowMenu?: (context: ResourceEditListRowMenuContext) => ResourceListRowMenuItem[];
  onAudit?: (item: ResourceListItem) => void;
};

function defaultRowMenuItems(
  context: ResourceEditListRowMenuContext,
  options: { replicate?: boolean; delete?: boolean; onAudit?: (item: ResourceListItem) => void },
): ResourceListRowMenuItem[] {
  const { item, helpers, pending, deleteItem } = context;
  const items: ResourceListRowMenuItem[] = [
    {
      id: "edit",
      label: "Edit",
      icon: <ResourceListMenuIcon name="edit" />,
      onSelect: () => helpers.openEdit(item.id),
    },
  ];

  if (options.replicate) {
    items.push({
      id: "replicate",
      label: helpers.replicateLabel ?? "Replicate",
      icon: <ResourceListMenuIcon name="copy" />,
      onSelect: () => helpers.openReplicate(item.id),
    });
  }

  if (options.delete) {
    items.push({
      id: "delete",
      label: helpers.deleteLabel ?? "Delete",
      icon: <ResourceListMenuIcon name="delete" />,
      destructive: true,
      disabled: pending,
      onSelect: () => deleteItem(item.id),
    });
  }

  if (options.onAudit) {
    items.push({
      id: "audit",
      label: "Audit",
      icon: <ResourceListMenuIcon name="audit" />,
      onSelect: () => options.onAudit?.(item),
    });
  }

  return items;
}

export function ResourceEditList({
  schema,
  items,
  helpers,
  handlers,
  actions,
  loading = false,
  error = null,
  descriptions,
  description: descriptionProp,
  emptyTitle: emptyTitleProp,
  emptyDescription: emptyDescriptionProp,
  selectedId = null,
  onSelectItem,
  showEditButton = true,
  showRowMenu = true,
  renderRowMenu,
  onAudit,
}: ResourceEditListProps) {
  const copy = useMemo(() => resolveResourceListAndEditCopy(descriptions), [descriptions]);
  const description = descriptionProp ?? copy?.list;
  const emptyTitle = emptyTitleProp ?? copy?.emptyTitle ?? "No items yet";
  const emptyDescription = emptyDescriptionProp ?? copy?.empty;
  const columns = listColumnsFromFields(schema.fields);
  const selectedRowRef = useRef<HTMLTableRowElement>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const {
    formLoading,
    loadError,
    openCreate,
    openEdit,
    createLabel: createLabelFromHelpers,
  } = helpers;
  const createLabel = createLabelFromHelpers ?? copy?.createLabel;

  const deleteItem = useCallback(
    async (itemId: string) => {
      const { onDelete, onDeleted, confirmDelete } = handlers ?? {};
      if (!onDelete) {
        return;
      }

      const confirmed = confirmDelete
        ? await confirmDelete(itemId)
        : window.confirm("Delete this item? This cannot be undone.");
      if (!confirmed) {
        return;
      }

      setDeleteError(null);
      setDeletingId(itemId);
      try {
        await onDelete(itemId);
        await onDeleted?.();
      } catch (err) {
        setDeleteError(err instanceof Error ? err.message : "Delete failed");
      } finally {
        setDeletingId(null);
      }
    },
    [handlers],
  );

  useEffect(() => {
    if (!selectedId || loading || items.length === 0) {
      return;
    }
    selectedRowRef.current?.scrollIntoView({ block: "nearest", behavior: "auto" });
  }, [loading, items.length, selectedId]);

  const hasRowActions = showEditButton || showRowMenu;
  const canDelete = Boolean(actions?.delete && handlers?.onDelete);

  return (
    <div className="appkit-resource-edit-list">
      <div className="appkit-resource-edit-list__header">
        {description ? <div className="appkit-resource-edit-list__description">{description}</div> : null}
        <button
          type="button"
          className="appkit-resource-edit-button appkit-resource-edit-button--primary appkit-resource-edit-list__create"
          disabled={formLoading}
          onClick={openCreate}
        >
          {createLabel ?? "Add"}
        </button>
      </div>

      {loadError ? <div className="appkit-resource-edit-view__error">{loadError}</div> : null}
      {formLoading ? <p className="appkit-resource-edit-list__status">Loading...</p> : null}
      {deleteError ? <div className="appkit-resource-edit-view__error">{deleteError}</div> : null}
      {error ? <div className="appkit-resource-edit-view__error">{error}</div> : null}

      {items.length === 0 && !loading && !error ? (
        <div className="appkit-resource-edit-list__empty">
          <p className="appkit-resource-edit-list__empty-title">{emptyTitle}</p>
          {emptyDescription ? (
            <div className="appkit-resource-edit-list__empty-description">{emptyDescription}</div>
          ) : null}
          <button
            type="button"
            className="appkit-resource-edit-button appkit-resource-edit-button--primary appkit-resource-edit-list__empty-action"
            disabled={formLoading}
            onClick={openCreate}
          >
            {createLabel ?? "Add"}
          </button>
        </div>
      ) : (
        <div className="appkit-resource-edit-list__table-wrap">
          <table className="appkit-resource-edit-list__table">
            <thead className="appkit-resource-edit-list__head">
              <tr>
                {columns.map((column) => (
                  <th
                    key={column.key}
                    className={
                      isListCheckboxField(column)
                        ? "appkit-resource-edit-list__th appkit-resource-edit-list__th--bool"
                        : "appkit-resource-edit-list__th"
                    }
                  >
                    {column.label}
                  </th>
                ))}
                {hasRowActions ? (
                  <th className="appkit-resource-edit-list__th appkit-resource-edit-list__th--actions">Actions</th>
                ) : null}
              </tr>
            </thead>
            <tbody>
              {items.map((item, index) => {
                const isSelected = item.id === selectedId;
                const isPending = deletingId === item.id;
                const isLast = index === items.length - 1;
                const rowLabel = item.values.name ?? item.id;
                const menuContext: ResourceEditListRowMenuContext = {
                  item,
                  helpers,
                  pending: isPending,
                  deleteItem: (itemId) => {
                    void deleteItem(itemId);
                  },
                };
                const selectRow = () => {
                  onSelectItem?.(item.id);
                };
                const menuItems = (
                  renderRowMenu?.(menuContext) ??
                  defaultRowMenuItems(menuContext, {
                    replicate: actions?.replicate,
                    delete: canDelete,
                    onAudit,
                  })
                ).map((menuItem) => ({
                  ...menuItem,
                  onSelect: () => {
                    selectRow();
                    menuItem.onSelect();
                  },
                }));

                return (
                  <tr
                    key={item.id}
                    ref={isSelected ? selectedRowRef : undefined}
                    aria-selected={isSelected}
                    className={`${resourceListRowClassName(isSelected)}${isLast ? "" : " appkit-resource-edit-list__row--bordered"}`}
                    onClick={(event) => {
                      if (!onSelectItem || isResourceListInteractiveTarget(event.target)) {
                        return;
                      }
                      onSelectItem(item.id);
                    }}
                  >
                    {columns.map((column, columnIndex) => {
                      const isBool = isListCheckboxField(column);
                      const className = [
                        "appkit-resource-edit-list__td",
                        columnIndex === 0 ? "appkit-resource-edit-list__td--primary" : "",
                        isBool ? "appkit-resource-edit-list__td--bool" : "",
                      ]
                        .filter(Boolean)
                        .join(" ");
                      return (
                        <td key={column.key} className={className}>
                          {isBool ? (
                            <ResourceListBoolCell
                              checked={listCellBoolValue(item.values, column.key)}
                              label={`${rowLabel} ${column.label.toLowerCase()}`}
                            />
                          ) : (
                            listCellValue(item.values, column.key)
                          )}
                        </td>
                      );
                    })}
                    {hasRowActions ? (
                      <td className="appkit-resource-edit-list__td appkit-resource-edit-list__td--actions">
                        <div className="appkit-resource-edit-list__actions">
                          {showEditButton ? (
                            <button
                              type="button"
                              className="appkit-resource-edit-button appkit-resource-edit-button--primary appkit-resource-edit-button--sm"
                              disabled={isPending || formLoading}
                              onClick={() => {
                                selectRow();
                                openEdit(item.id);
                              }}
                            >
                              Edit
                            </button>
                          ) : null}
                          {showRowMenu && menuItems.length > 0 ? (
                            <ResourceListRowMenu
                              ariaLabel={`Actions for ${rowLabel}`}
                              items={menuItems}
                              onOpenChange={(open) => {
                                if (open) {
                                  selectRow();
                                }
                              }}
                            />
                          ) : null}
                        </div>
                      </td>
                    ) : null}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
