"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ResourceListAndEdit,
  type ResourceEditListConfig,
  type ResourceListAndEditProps,
} from "./ResourceListAndEdit";
import type { ResourceViewEditHandlers, ResourceViewListHandlers } from "./handlers";
import {
  buildReplicaEditState,
  cloneResourceEditState,
  withFetchSchema,
  type ResourceEndpointHttp,
  type ResourceListAndEditEndpointPaths,
} from "./resource-endpoints";
import {
  useResourceViewListHandlers,
  type ListConfirmDeleteOptions,
} from "./resource-list-edit";
import type { ResourceEditState, ResourceListItem, ResourceSchema } from "./types";
import { ResourceMode } from "./types";

export type ResourceListAndEditEndpointsListProps = Omit<
  ResourceEditListConfig,
  "schema" | "items" | "loading" | "error"
> & {
  /** Clear URL/host selection after the selected row is deleted. */
  onClearSelection?: () => void;
};

export type ResourceListAndEditEndpointsProps = Omit<
  ResourceListAndEditProps,
  "editHandlers" | "listHandlers" | "list"
> & {
  http: ResourceEndpointHttp;
  endpoints: ResourceListAndEditEndpointPaths;
  /**
   * Query appended on the initial list load to embed a create template.
   * Default: `fetchSchema=1`. Pass `false` to skip.
   */
  fetchSchemaParam?: string | false;
  /** List UI + selection (selectedId, onSelectItem, onClearSelection, …). */
  list?: ResourceListAndEditEndpointsListProps;
  confirmDelete?: ListConfirmDeleteOptions;
  onLoadingChange?: (loading: boolean) => void;
  uploadImage?: (file: File) => Promise<string>;
};

/**
 * List-and-edit resource cycle driven by standard endpoint URLs:
 * list (+ optional fetchSchema), GET/PATCH/DELETE item, POST create.
 * Replicate uses GET item + create-template overlay (no separate /schema call).
 */
export function ResourceListAndEditEndpoints({
  http,
  endpoints,
  fetchSchemaParam = "fetchSchema=1",
  list: listProp,
  confirmDelete,
  onLoadingChange,
  uploadImage,
  descriptions,
  ...rest
}: ResourceListAndEditEndpointsProps) {
  const createUrl = endpoints.create ?? endpoints.list;
  const [listSchema, setListSchema] = useState<ResourceSchema | null>(null);
  const [items, setItems] = useState<ResourceListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const createTemplateRef = useRef<ResourceEditState | null>(null);
  const onLoadingChangeRef = useRef(onLoadingChange);
  const selectedId = listProp?.selectedId;
  const onSelectItem = listProp?.onSelectItem;
  const onClearSelection = listProp?.onClearSelection;

  useEffect(() => {
    onLoadingChangeRef.current = onLoadingChange;
  });

  const setLoadingState = useCallback((next: boolean) => {
    setLoading(next);
    onLoadingChangeRef.current?.(next);
  }, []);

  const loadItems = useCallback(
    async (options?: { fetchSchema?: boolean }) => {
      const url =
        options?.fetchSchema && fetchSchemaParam !== false
          ? withFetchSchema(endpoints.list, fetchSchemaParam || "fetchSchema=1")
          : endpoints.list;
      const result = await http.getList(url);
      setListSchema(result.listSchema);
      setItems(result.items);
      if (result.createTemplate) {
        createTemplateRef.current = result.createTemplate;
      }
      return result.items;
    },
    [endpoints.list, fetchSchemaParam, http],
  );

  useEffect(() => {
    setLoadingState(true);
    setError(null);
    loadItems({ fetchSchema: true })
      .catch((err) => setError(err instanceof Error ? err.message : "Failed to load"))
      .finally(() => setLoadingState(false));
  }, [loadItems, setLoadingState]);

  const resolvedConfirmDelete = useMemo<ListConfirmDeleteOptions | undefined>(() => {
    const singular = descriptions?.singularName?.trim();
    const defaultFallback = singular ? `this ${singular}` : undefined;
    if (!confirmDelete && !defaultFallback) {
      return undefined;
    }
    return {
      ...confirmDelete,
      fallbackLabel: confirmDelete?.fallbackLabel ?? defaultFallback ?? "this item",
    };
  }, [confirmDelete, descriptions?.singularName]);

  const listLifecycleHandlers = useResourceViewListHandlers({
    items,
    reload: () => loadItems(),
    selectedId,
    onSelectItem,
    onClearSelection,
    confirmDelete: resolvedConfirmDelete,
  });

  const handleLoad = useCallback<ResourceViewEditHandlers["onLoad"]>(
    async ({ mode, itemId }) => {
      const template = createTemplateRef.current;
      if (mode === "create") {
        if (!template) {
          throw new Error("Create template not loaded");
        }
        return cloneResourceEditState(template);
      }
      if (mode === "replicate" && itemId) {
        if (!template) {
          throw new Error("Create template not loaded");
        }
        const source = await http.getEdit(endpoints.item(itemId));
        return buildReplicaEditState(template, source);
      }
      if (mode === "edit" && itemId) {
        return http.getEdit(endpoints.item(itemId));
      }
      throw new Error("Invalid edit request");
    },
    [endpoints, http],
  );

  const handleSubmit = useCallback<ResourceViewEditHandlers["onSubmit"]>(
    async (values, { mode, state }) => {
      if (mode === "create" || mode === "replicate") {
        const nextState = await http.postEdit(createUrl, values);
        const createdId = nextState.values.id;
        if (createdId) {
          onSelectItem?.(createdId);
        }
        return nextState;
      }
      const resourceId = state.values.id;
      if (!resourceId) {
        throw new Error("Missing resource id");
      }
      return http.patchEdit(endpoints.item(resourceId), values);
    },
    [createUrl, endpoints, http, onSelectItem],
  );

  const handleDelete = useCallback(
    async (itemId: string) => {
      await http.deleteResource(endpoints.item(itemId));
    },
    [endpoints, http],
  );

  const editHandlers = useMemo<ResourceViewEditHandlers>(
    () => ({
      onLoad: handleLoad,
      onSubmit: handleSubmit,
      onUploadImage: uploadImage,
    }),
    [handleLoad, handleSubmit, uploadImage],
  );

  const listHandlers = useMemo<ResourceViewListHandlers>(
    () => ({
      ...listLifecycleHandlers,
      onDelete: handleDelete,
    }),
    [handleDelete, listLifecycleHandlers],
  );

  const listConfig = useMemo<ResourceEditListConfig>(() => {
    const { onClearSelection: _clear, ...listUi } = listProp ?? {};
    return {
      ...listUi,
      selectedId,
      onSelectItem,
      schema: listSchema ?? { name: "", label: "", mode: ResourceMode.UNSPECIFIED, fields: [] },
      items,
      loading,
      error,
    };
  }, [error, items, listProp, listSchema, loading, onSelectItem, selectedId]);

  return (
    <ResourceListAndEdit
      {...rest}
      descriptions={descriptions}
      editHandlers={editHandlers}
      listHandlers={listHandlers}
      list={listConfig}
    />
  );
}
