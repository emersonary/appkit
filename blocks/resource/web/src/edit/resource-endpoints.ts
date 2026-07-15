import { ResourceRecordState, type ResourceEditState, type ResourceListItem, type ResourceSchema } from "./types";

/** List payload returned by `ResourceEndpointHttp.getList`. */
export type ResourceListResult = {
  listSchema: ResourceSchema;
  items: ResourceListItem[];
  createTemplate?: ResourceEditState;
};

/**
 * Host-provided transport. AppKit owns the resource cycle; the host owns
 * URL auth, content-type, and decoding into AppKit resource types.
 */
export type ResourceEndpointHttp = {
  getEdit: (url: string) => Promise<ResourceEditState>;
  postEdit: (url: string, values: Record<string, string>) => Promise<ResourceEditState>;
  patchEdit: (url: string, values: Record<string, string>) => Promise<ResourceEditState>;
  deleteResource: (url: string) => Promise<void>;
  getList: (url: string) => Promise<ResourceListResult>;
};

export type ResourceEditEndpointPaths = {
  /** GET — load edit form (`ResourceEditState`). */
  get: string;
  /** PATCH — save. */
  patch: string;
};

export type ResourceListAndEditEndpointPaths = {
  /** GET — list rows; optional `?fetchSchema=1` on first load for create template. */
  list: string;
  /** GET / PATCH / DELETE — single item by id. */
  item: (id: string) => string;
  /** POST — create (defaults to `list`). */
  create?: string;
};

/** Append `fetchSchema=1` (or a custom query) to a list URL. */
export function withFetchSchema(listUrl: string, param: string = "fetchSchema=1"): string {
  const trimmed = param.trim();
  if (!trimmed) {
    return listUrl;
  }
  const joiner = listUrl.includes("?") ? "&" : "?";
  return `${listUrl}${joiner}${trimmed}`;
}

export function cloneResourceEditState(state: ResourceEditState): ResourceEditState {
  return {
    ...state,
    values: { ...state.values },
    relatedLinks: state.relatedLinks ? [...state.relatedLinks] : state.relatedLinks,
  };
}

/**
 * Overlay editable source values onto a create template (standard replicate).
 * Sets `name` to `"… (copy)"` / `"Copy"` and `is_primary` to `"false"` when present.
 */
export function buildReplicaEditState(
  template: ResourceEditState,
  source: ResourceEditState,
): ResourceEditState {
  const values = { ...template.values };
  for (const field of template.schema.fields) {
    if (!field.editable || field.readOnly) {
      continue;
    }
    if (Object.prototype.hasOwnProperty.call(source.values, field.key)) {
      values[field.key] = source.values[field.key] ?? "";
    }
  }
  if (Object.prototype.hasOwnProperty.call(values, "name") || Object.prototype.hasOwnProperty.call(source.values, "name")) {
    const name = (values.name ?? "").trim();
    values.name = name ? `${name} (copy)` : "Copy";
  }
  if (Object.prototype.hasOwnProperty.call(values, "is_primary") || Object.prototype.hasOwnProperty.call(source.values, "is_primary")) {
    values.is_primary = "false";
  }
  delete values.id;
  return {
    ...cloneResourceEditState(template),
    values,
    baselineValues: { ...template.values },
    recordState: ResourceRecordState.NEW,
  };
}
