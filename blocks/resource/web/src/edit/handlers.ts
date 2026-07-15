import type { ResourceEditState } from "./types";

export type ResourceEditMode = "create" | "edit" | "replicate";

export type ResourceEditLoadRequest = {
  mode: ResourceEditMode;
  itemId?: string;
};

/** `handlers` on ResourceEdit (schema-driven form). */
export type ResourceFormHandlers = {
  onSubmit: (values: Record<string, string>) => Promise<ResourceEditState | void>;
  onCreate?: (values: Record<string, string>) => Promise<ResourceEditState | void>;
  onUploadImage?: (file: File) => Promise<string>;
};

/**
 * `editHandlers` on ResourceListAndEdit / ResourceEditScreen: load items into the form
 * and persist with list-and-edit context (mode, loaded state).
 */
export type ResourceViewEditHandlers = {
  onLoad: (request: ResourceEditLoadRequest) => Promise<ResourceEditState>;
  onSubmit: (
    values: Record<string, string>,
    context: { mode: ResourceEditMode; state: ResourceEditState },
  ) => Promise<ResourceEditState | void>;
  onUploadImage?: ResourceFormHandlers["onUploadImage"];
};

/** `listHandlers` on ResourceListAndEdit; `handlers` on ResourceEditList. */
export type ResourceViewListHandlers = {
  onSaved?: () => void | Promise<void>;
  onDelete?: (itemId: string) => void | Promise<void>;
  onDeleted?: () => void | Promise<void>;
  confirmDelete?: (itemId: string) => boolean | Promise<boolean>;
};
