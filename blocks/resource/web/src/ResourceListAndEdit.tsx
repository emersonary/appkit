import { useCallback, useState, type ReactNode } from "react";
import { ResourceEdit, type ResourceEditProps } from "./ResourceEdit";
import { ResourceList, type ResourceCellInterceptProps, type ResourceListProps, type ResourceRowAction } from "./ResourceList";
import { itemID, itemName } from "./schema";
import type { ResourceItem, ResourceEditMode, ResourceSchema } from "./types";

export type ResourceEditConfig = Omit<
  ResourceEditProps,
  "schema" | "item" | "onCancel" | "onSubmit" | "saving" | "error" | "mode"
> & {
  onSubmit: (values: ResourceItem, context: { mode: ResourceEditMode }) => void | Promise<void>;
  onEditRequest?: (
    item: ResourceItem,
    mode: Exclude<ResourceEditMode, "create">,
  ) => ResourceItem | Promise<ResourceItem>;
  onDelete?: (item: ResourceItem) => void | Promise<void>;
  onDeleted?: () => void | Promise<void>;
  confirmDelete?: (item: ResourceItem) => boolean | Promise<boolean>;
  createLabel?: string;
  replicateLabel?: string;
  deleteLabel?: string;
};

export type ResourceListAndEditActions = {
  create?: boolean;
  edit?: boolean;
  replicate?: boolean;
  delete?: boolean;
};

export type ResourceListAndEditProps = {
  schema: ResourceSchema;
  items: ResourceItem[];
  total?: number;
  page?: number;
  pageSize?: number;
  loading?: boolean;
  renderCell?: (props: ResourceCellInterceptProps) => ReactNode;
  interceptCell?: (props: ResourceCellInterceptProps) => ReactNode | null | undefined;
  onPageChange?: (page: number) => void;
  onToggle?: ResourceListProps["onToggle"];
  editForm: ResourceEditConfig;
  actions?: ResourceListAndEditActions;
};

const defaultActions: Required<ResourceListAndEditActions> = {
  create: true,
  edit: true,
  replicate: true,
  delete: false,
};

export function ResourceListAndEdit({
  schema,
  items,
  total,
  page,
  pageSize,
  loading,
  renderCell,
  interceptCell,
  onPageChange,
  onToggle,
  editForm,
  actions: actionsProp,
}: ResourceListAndEditProps) {
  const actions = { ...defaultActions, ...actionsProp };
  const [mode, setMode] = useState<ResourceEditMode | null>(null);
  const [formItem, setFormItem] = useState<ResourceItem | undefined>();
  const [formLoading, setFormLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const closeForm = useCallback(() => {
    setMode(null);
    setFormItem(undefined);
    setError(null);
  }, []);

  const openForm = useCallback(
    async (nextMode: ResourceEditMode, listItem?: ResourceItem) => {
      setLoadError(null);
      setError(null);

      if (nextMode === "create") {
        setMode("create");
        setFormItem({});
        return;
      }

      if (!listItem) {
        return;
      }

      setFormLoading(true);
      try {
        const item = editForm.onEditRequest
          ? await editForm.onEditRequest(listItem, nextMode)
          : { ...listItem };
        setMode(nextMode);
        setFormItem(item);
      } catch (err) {
        setLoadError(err instanceof Error ? err.message : "Failed to load item");
      } finally {
        setFormLoading(false);
      }
    },
    [editForm],
  );

  const handleSubmit = useCallback(
    async (values: ResourceItem) => {
      if (!mode) {
        return;
      }

      setSaving(true);
      setError(null);
      try {
        await editForm.onSubmit(values, { mode });
        closeForm();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Save failed");
      } finally {
        setSaving(false);
      }
    },
    [closeForm, editForm, mode],
  );

  const handleDelete = useCallback(
    async (item: ResourceItem) => {
      if (!editForm.onDelete) {
        return;
      }

      const confirmed = editForm.confirmDelete
        ? await editForm.confirmDelete(item)
        : window.confirm(
            `Delete "${itemName(schema, item) || "this item"}"? This cannot be undone.`,
          );
      if (!confirmed) {
        return;
      }

      const itemKey = itemID(schema, item);
      setDeleteError(null);
      setDeletingId(itemKey || null);
      try {
        await editForm.onDelete(item);
        await editForm.onDeleted?.();
      } catch (err) {
        setDeleteError(err instanceof Error ? err.message : "Delete failed");
      } finally {
        setDeletingId(null);
      }
    },
    [editForm, schema],
  );

  const rowActions: ResourceRowAction[] = [];
  if (actions.edit) {
    rowActions.push({
      id: "edit",
      label: "Edit",
      onAction: (item) => {
        void openForm("edit", item);
      },
    });
  }
  if (actions.replicate) {
    rowActions.push({
      id: "replicate",
      label: editForm.replicateLabel ?? "Replicate",
      onAction: (item) => {
        void openForm("replicate", item);
      },
    });
  }
  if (actions.delete && editForm.onDelete) {
    rowActions.push({
      id: "delete",
      label: editForm.deleteLabel ?? "Delete",
      destructive: true,
      disabled: (item) => deletingId === itemID(schema, item),
      onAction: (item) => {
        void handleDelete(item);
      },
    });
  }

  if (mode) {
    return (
      <div className="appkit-resource-view">
        <ResourceEdit
          schema={schema}
          item={formItem}
          mode={mode}
          error={error}
          saving={saving}
          relationOptions={editForm.relationOptions}
          readOnly={editForm.readOnly}
          submitLabel={editForm.submitLabel}
          renderField={editForm.renderField}
          interceptField={editForm.interceptField}
          onSubmit={handleSubmit}
          onCancel={closeForm}
          cancelLabel="Back"
        />
      </div>
    );
  }

  return (
    <div className="appkit-resource-view">
      {loadError ? <p className="appkit-resource-view__error">{loadError}</p> : null}
      {deleteError ? <p className="appkit-resource-view__error">{deleteError}</p> : null}
      {formLoading ? <p className="appkit-resource-view__loading">Loading...</p> : null}
      {actions.create ? (
        <div className="appkit-resource-view__toolbar">
          <button
            type="button"
            className="appkit-resource-button"
            disabled={formLoading}
            onClick={() => {
              void openForm("create");
            }}
          >
            {editForm.createLabel ?? `New ${schema.name}`}
          </button>
        </div>
      ) : null}
      <ResourceList
        schema={schema}
        items={items}
        total={total}
        page={page}
        pageSize={pageSize}
        loading={loading || formLoading}
        renderCell={renderCell}
        interceptCell={interceptCell}
        rowActions={rowActions}
        onPageChange={onPageChange}
        onToggle={onToggle}
      />
    </div>
  );
}
