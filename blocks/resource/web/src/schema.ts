import type { ResourceColumn, ResourceField, ResourceFormSection, ResourceSchema } from "./types";

export function idField(schema: ResourceSchema): string {
  return schema.id_field?.trim() || "id";
}

export function nameField(schema: ResourceSchema): string {
  return schema.name_field?.trim() || "name";
}

export function parentIDField(schema: ResourceSchema): string {
  return schema.parent_id_field?.trim() || schema.list?.tree?.parent_id_field?.trim() || "";
}

export function fieldByKey(schema: ResourceSchema): Map<string, ResourceField> {
  return new Map(schema.fields.map((field) => [field.key, field]));
}

export function listColumns(schema: ResourceSchema): ResourceColumn[] {
  const configured = schema.list?.columns?.filter((column) => !column.hidden) ?? [];
  if (configured.length > 0) {
    return [...configured].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0));
  }
  return schema.fields
    .filter((field) => !field.hidden && !field.list_hidden)
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
    .map((field) => ({ field_key: field.key, label: field.label, sort_order: field.sort_order }));
}

export function formSections(schema: ResourceSchema): ResourceFormSection[] {
  const configured = schema.form?.sections ?? [];
  if (configured.length > 0) {
    return [...configured].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0));
  }

  const bySection = new Map<string, string[]>();
  for (const field of schema.fields) {
    if (field.hidden || field.form_hidden) {
      continue;
    }
    const title = field.section?.trim() || "General";
    bySection.set(title, [...(bySection.get(title) ?? []), field.key]);
  }

  return [...bySection.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([title, fields]) => ({
      id: title.toLowerCase().replace(/\s+/g, "_"),
      title,
      fields,
    }));
}

export function itemID(schema: ResourceSchema, item: Record<string, unknown>): string {
  const value = item[idField(schema)];
  return value == null ? "" : String(value);
}

export function itemName(schema: ResourceSchema, item: Record<string, unknown>): string {
  const value = item[nameField(schema)];
  return value == null ? "" : String(value);
}
