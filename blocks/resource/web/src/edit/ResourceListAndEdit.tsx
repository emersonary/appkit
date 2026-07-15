"use client";

import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { ResourceEdit, type ResourceEditProps } from "./ResourceEdit";
import { ResourceEditList, type ResourceEditListProps } from "./ResourceEditList";
import type { ResourceViewEditHandlers, ResourceEditMode, ResourceViewListHandlers } from "./handlers";
import {
  resolveResourceListAndEditCopy,
  type ResourceListAndEditDescriptions,
} from "./resource-descriptions";
import type { ResourceEditState } from "./types";

export type { ResourceViewEditHandlers, ResourceEditLoadRequest, ResourceEditMode, ResourceViewListHandlers } from "./handlers";
export type {
  ResourceListAndEditDescriptions,
  ResourceListAndEditDescriptionOverrides,
  ResolvedResourceListAndEditCopy,
} from "./resource-descriptions";

export type ResourceListAndEditActions = {
  create?: boolean;
  edit?: boolean;
  replicate?: boolean;
  delete?: boolean;
};

/** Form presentation options (labels, layout, related links). Handlers use `editHandlers` / `listHandlers`. */
export type ResourceEditOptions = Omit<
  ResourceEditProps,
  "state" | "handlers" | "saving" | "onBack" | "description"
> & {
  createLabel?: string;
  replicateLabel?: string;
  deleteLabel?: string;
  /**
   * Intro copy for standalone edit screens (`ResourceEditScreen`).
   * For `ResourceListAndEdit`, prefer the top-level `descriptions` prop.
   */
  description?: ReactNode | ((mode: ResourceEditMode) => ReactNode);
};

export type ResourceEditListHelpers = {
  formLoading: boolean;
  loadError: string | null;
  openCreate: () => void;
  openEdit: (itemId: string) => void;
  openReplicate: (itemId: string) => void;
  createLabel?: string;
  replicateLabel?: string;
  deleteLabel?: string;
};

export type ResourceEditListConfig = Omit<
  ResourceEditListProps,
  "helpers" | "handlers" | "actions" | "descriptions" | "description" | "emptyTitle" | "emptyDescription"
>;

export type ResourceListAndEditProps = {
  editHandlers: ResourceViewEditHandlers;
  listHandlers?: ResourceViewListHandlers;
  form?: ResourceEditOptions;
  actions?: ResourceListAndEditActions;
  list?: ResourceEditListConfig;
  /** Resource naming + optional copy overrides for list/form surfaces. */
  descriptions?: ResourceListAndEditDescriptions;
  onEditingChange?: (
    editing: boolean,
    state: ResourceEditState | null,
    mode: ResourceEditMode | null,
  ) => void;
  /** Override the built-in schema-driven list. Ignored when `list` is set unless provided explicitly. */
  renderList?: (helpers: ResourceEditListHelpers) => ReactNode;
  renderForm?: (
    props: ResourceEditProps & {
      mode: ResourceEditMode;
      onBack: () => void;
      loadError: string | null;
    },
  ) => ReactNode;
};

const defaultActions: Required<ResourceListAndEditActions> = {
  create: true,
  edit: true,
  replicate: true,
  delete: false,
};

export function ResourceListAndEdit({
  editHandlers,
  listHandlers,
  form: formOptions = {},
  actions: actionsProp,
  list,
  descriptions,
  onEditingChange,
  renderList,
  renderForm,
}: ResourceListAndEditProps) {
  const { onLoad, onSubmit, onUploadImage } = editHandlers;
  const onLoadRef = useRef(onLoad);
  const onSubmitRef = useRef(onSubmit);
  const onSavedRef = useRef(listHandlers?.onSaved);
  const onEditingChangeRef = useRef(onEditingChange);

  useEffect(() => {
    onLoadRef.current = onLoad;
    onSubmitRef.current = onSubmit;
    onSavedRef.current = listHandlers?.onSaved;
    onEditingChangeRef.current = onEditingChange;
  });

  const actions = { ...defaultActions, ...actionsProp };
  const copy = useMemo(() => resolveResourceListAndEditCopy(descriptions), [descriptions]);
  const [mode, setMode] = useState<ResourceEditMode | null>(null);
  const [editState, setEditState] = useState<ResourceEditState | null>(null);
  const [formLoading, setFormLoading] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  const closeForm = useCallback(() => {
    setMode(null);
    setEditState(null);
    setLoadError(null);
    onEditingChangeRef.current?.(false, null, null);
  }, []);

  const openForm = useCallback(async (nextMode: ResourceEditMode, itemId?: string) => {
    if (itemId && (nextMode === "edit" || nextMode === "replicate")) {
      list?.onSelectItem?.(itemId);
    }
    setLoadError(null);
    setFormLoading(true);
    try {
      const state = await onLoadRef.current({ mode: nextMode, itemId });
      setMode(nextMode);
      setEditState(state);
      onEditingChangeRef.current?.(true, state, nextMode);
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "Failed to load item");
      onEditingChangeRef.current?.(false, null, null);
    } finally {
      setFormLoading(false);
    }
  }, [list]);

  const handlePersist = useCallback(
    async (values: Record<string, string>) => {
      if (!mode || !editState) {
        return;
      }

      const nextState = await onSubmitRef.current(values, { mode, state: editState });
      await onSavedRef.current?.();
      closeForm();
      return nextState;
    },
    [closeForm, editState, mode],
  );

  const listHelpers: ResourceEditListHelpers = {
    formLoading,
    loadError,
    openCreate: actions.create
      ? () => {
          void openForm("create");
        }
      : () => {},
    openEdit: actions.edit
      ? (itemId) => {
          void openForm("edit", itemId);
        }
      : () => {},
    openReplicate: actions.replicate
      ? (itemId) => {
          void openForm("replicate", itemId);
        }
      : () => {},
    createLabel: copy?.createLabel ?? formOptions.createLabel,
    replicateLabel: copy?.replicateLabel ?? formOptions.replicateLabel,
    deleteLabel: copy?.deleteLabel ?? formOptions.deleteLabel,
  };

  if (mode && editState) {
    const formProps: ResourceEditProps = {
      state: editState,
      twoColumnLayout: formOptions.twoColumnLayout,
      descriptions,
      mode,
      description:
        typeof formOptions.description === "function"
          ? formOptions.description(mode)
          : formOptions.description,
      handlers: {
        onSubmit: handlePersist,
        onCreate: handlePersist,
        onUploadImage,
      },
      LinkComponent: formOptions.LinkComponent,
      renderRelatedLinkIcon: formOptions.renderRelatedLinkIcon,
      renderRelatedLinkChevron: formOptions.renderRelatedLinkChevron,
      relatedLinksTitle: formOptions.relatedLinksTitle,
      onBack: closeForm,
      backLabel: formOptions.backLabel,
    };

    const formShell = renderForm ? (
      renderForm({ ...formProps, mode, onBack: closeForm, loadError })
    ) : (
      <>
        {loadError ? <div className="appkit-resource-edit-view__error">{loadError}</div> : null}
        <ResourceEdit {...formProps} />
      </>
    );

    return <div className="appkit-resource-edit-view">{formShell}</div>;
  }

  return (
    <div className="appkit-resource-edit-view">
      {list ? (
        <ResourceEditList
          {...list}
          descriptions={descriptions}
          helpers={listHelpers}
          handlers={listHandlers}
          actions={{ replicate: actions.replicate, delete: actions.delete }}
        />
      ) : renderList ? (
        renderList(listHelpers)
      ) : null}
    </div>
  );
}
