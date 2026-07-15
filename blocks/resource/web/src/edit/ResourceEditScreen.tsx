"use client";

import { useCallback, useEffect, useRef, useState, type ReactNode } from "react";
import { ResourceEdit, type ResourceEditProps } from "./ResourceEdit";
import type { ResourceViewEditHandlers, ResourceEditMode } from "./handlers";
import type { ResourceListAndEditDescriptions } from "./resource-descriptions";
import type { ResourceEditState } from "./types";
import type { ResourceEditOptions } from "./ResourceListAndEdit";

export type ResourceEditScreenProps = {
  editHandlers: ResourceViewEditHandlers;
  form?: ResourceEditOptions;
  /** Resource naming + optional copy overrides (passed through to ResourceEdit). */
  descriptions?: ResourceListAndEditDescriptions;
  /** Passed to `editHandlers.onLoad` on mount and when it changes. Default: `edit`. */
  loadRequest?: { mode: ResourceEditMode; itemId?: string };
  onLoaded?: (state: ResourceEditState) => void;
  onLoadingChange?: (loading: boolean) => void;
  renderForm?: (props: ResourceEditProps & { loadError: string | null }) => ReactNode;
};

export function ResourceEditScreen({
  editHandlers,
  form: formOptions = {},
  descriptions,
  loadRequest = { mode: "edit" },
  onLoaded,
  onLoadingChange,
  renderForm,
}: ResourceEditScreenProps) {
  const { onLoad, onSubmit, onUploadImage } = editHandlers;
  const [state, setState] = useState<ResourceEditState | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const onLoadRef = useRef(onLoad);
  const onLoadedRef = useRef(onLoaded);
  const onLoadingChangeRef = useRef(onLoadingChange);

  useEffect(() => {
    onLoadRef.current = onLoad;
    onLoadedRef.current = onLoaded;
    onLoadingChangeRef.current = onLoadingChange;
  });

  const setLoadingState = useCallback((nextLoading: boolean) => {
    setLoading(nextLoading);
    onLoadingChangeRef.current?.(nextLoading);
  }, []);

  const loadMode = loadRequest.mode;
  const loadItemId = loadRequest.itemId;

  const load = useCallback(async () => {
    setLoadingState(true);
    setLoadError(null);
    try {
      const nextState = await onLoadRef.current({ mode: loadMode, itemId: loadItemId });
      setState(nextState);
      onLoadedRef.current?.(nextState);
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : "Failed to load");
      setState(null);
    } finally {
      setLoadingState(false);
    }
  }, [loadItemId, loadMode, setLoadingState]);

  useEffect(() => {
    void load();
  }, [load]);

  const applyNextState = useCallback((nextState: ResourceEditState | void) => {
    if (!nextState) {
      return nextState;
    }
    setState(nextState);
    onLoadedRef.current?.(nextState);
    return nextState;
  }, []);

  const handleSubmit = useCallback(
    async (values: Record<string, string>) => {
      if (!state) {
        return;
      }
      return applyNextState(await onSubmit(values, { mode: loadMode, state }));
    },
    [applyNextState, loadMode, onSubmit, state],
  );

  if (loading && !state) {
    return <div className="appkit-resource-edit-screen" aria-hidden />;
  }

  if (!state) {
    return loadError ? <div className="appkit-resource-edit-view__error">{loadError}</div> : null;
  }

  const formProps: ResourceEditProps = {
    state,
    twoColumnLayout: formOptions.twoColumnLayout,
    descriptions,
    mode: loadMode,
    description:
      typeof formOptions.description === "function"
        ? formOptions.description(loadMode)
        : formOptions.description,
    handlers: {
      onSubmit: handleSubmit,
      onCreate: handleSubmit,
      onUploadImage,
    },
    LinkComponent: formOptions.LinkComponent,
    renderRelatedLinkIcon: formOptions.renderRelatedLinkIcon,
    renderRelatedLinkChevron: formOptions.renderRelatedLinkChevron,
    relatedLinksTitle: formOptions.relatedLinksTitle,
    backLabel: formOptions.backLabel,
  };

  const formShell = renderForm ? (
    renderForm({ ...formProps, loadError })
  ) : (
    <>
      {loadError ? <div className="appkit-resource-edit-view__error">{loadError}</div> : null}
      <ResourceEdit {...formProps} />
    </>
  );

  return <div className="appkit-resource-edit-screen">{formShell}</div>;
}
