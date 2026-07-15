import type { CSSProperties } from "react";
import {
  ResourceRecordState,
  ResourceValidationKind,
  type ResourceEditState,
  type ResourceField,
  type ResourceSchema,
  type ResourceSubmitButtonState,
} from "./types";

export type { ResourceEditState, ResourceRelatedLink, ResourceSubmitButtonState } from "./types";
export { mapResourceRelatedLinks } from "./related-links";

export function emptyResourceSchema(): ResourceSchema {
  return { name: "", label: "", mode: 0, fields: [] };
}

export function isFieldRequired(field: ResourceField): boolean {
  if (field.required) return true;
  return field.validations.some((rule) => rule.kind === ResourceValidationKind.REQUIRED);
}

export function isResourceRecordNew(recordState: ResourceRecordState): boolean {
  return recordState === ResourceRecordState.NEW;
}

export function cloneResourceValues(values: Record<string, string>): Record<string, string> {
  return { ...values };
}

export function mergeServerOnlyValues(
  draft: Record<string, string>,
  serverValues: Record<string, string>,
  schema: ResourceSchema,
): Record<string, string> {
  const editableKeys = new Set(editableFields(schema.fields).map((field) => field.key));
  const next = { ...draft };
  for (const [key, value] of Object.entries(serverValues)) {
    if (!editableKeys.has(key)) {
      next[key] = value;
    }
  }
  return next;
}

export function isResourceDraftDirty(
  schema: ResourceSchema,
  draft: Record<string, string>,
  baseline: Record<string, string>,
): boolean {
  return changedEditableFieldKeys(schema, draft, baseline).size > 0;
}

export function changedEditableFieldKeys(
  schema: ResourceSchema,
  draft: Record<string, string>,
  baseline: Record<string, string>,
): Set<string> {
  const keys = new Set<string>();
  for (const field of editableFields(schema.fields)) {
    const draftValue = draft[field.key] ?? "";
    const baselineValue = baseline[field.key] ?? "";
    if (draftValue !== baselineValue) {
      keys.add(field.key);
    }
  }
  return keys;
}

export function getResourceSubmitButtonState(options: {
  recordState: ResourceRecordState;
  isDirty: boolean;
  saving: boolean;
  isValid: boolean;
}): ResourceSubmitButtonState {
  const { recordState, isDirty, saving, isValid } = options;
  const isNew = isResourceRecordNew(recordState);

  if (saving) {
    return { disabled: true, loading: true, label: isNew ? "Creating…" : "Saving…" };
  }

  if (!isDirty) {
    return {
      disabled: true,
      loading: false,
      label: isNew ? "Create" : "Up to date",
    };
  }

  if (!isValid) {
    return {
      disabled: true,
      loading: false,
      label: isNew ? "Create" : "Save changes",
    };
  }

  return {
    disabled: false,
    loading: false,
    label: isNew ? "Create" : "Save changes",
  };
}

export function groupFieldsBySection(
  fields: ResourceField[],
): Array<{ section: string; fields: ResourceField[] }> {
  const sections = new Map<string, ResourceField[]>();
  for (const field of fields) {
    if (!field.visible) continue;
    const section = field.section || "general";
    const bucket = sections.get(section) ?? [];
    bucket.push(field);
    sections.set(section, bucket);
  }

  return Array.from(sections.entries())
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([section, sectionFields]) => ({
      section,
      fields: [...sectionFields].sort((a, b) => a.order - b.order || a.key.localeCompare(b.key)),
    }));
}

export function formatSectionTitle(section: string): string {
  if (!section || section === "general") return "General";
  return section
    .split(/[_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

export function editableFields(fields: ResourceField[]): ResourceField[] {
  return fields.filter((field) => field.visible && field.editable && !field.readOnly);
}

export function buildSubmitValues(
  schema: ResourceSchema,
  draft: Record<string, string>,
): Record<string, string> {
  const out: Record<string, string> = {};
  for (const field of editableFields(schema.fields)) {
    out[field.key] = draft[field.key] ?? "";
  }
  return out;
}

export const resourceInputClassName = "appkit-resource-edit-field";

export const resourceReadonlyClassName = "appkit-resource-edit-field appkit-resource-edit-field--readonly";

export function imagePreviewStyle(field: ResourceField): CSSProperties {
  const style: CSSProperties = {
    objectFit: "contain",
  };
  if (field.maxWidth > 0) {
    style.maxWidth = field.maxWidth;
  }
  if (field.maxHeight > 0) {
    style.maxHeight = field.maxHeight;
  }
  return style;
}

async function readImageDimensions(file: File): Promise<{ width: number; height: number }> {
  const url = URL.createObjectURL(file);
  try {
    return await new Promise((resolve, reject) => {
      const img = new Image();
      img.onload = () => resolve({ width: img.naturalWidth, height: img.naturalHeight });
      img.onerror = () => reject(new Error("Could not read image dimensions"));
      img.src = url;
    });
  } finally {
    URL.revokeObjectURL(url);
  }
}

export async function validateImageDimensions(field: ResourceField, file: File): Promise<string | null> {
  const maxWidth = field.maxWidth;
  const maxHeight = field.maxHeight;
  if (maxWidth <= 0 && maxHeight <= 0) {
    return null;
  }

  const { width, height } = await readImageDimensions(file);
  if (maxWidth > 0 && width > maxWidth) {
    return `Image width must be ${maxWidth}px or less (got ${width}px)`;
  }
  if (maxHeight > 0 && height > maxHeight) {
    return `Image height must be ${maxHeight}px or less (got ${height}px)`;
  }
  return null;
}
