import { ResourceFieldKind, type ResourceField } from "./types";

/** Fields marked listable, sorted by order then key. */
export function listColumnsFromFields(fields: ResourceField[]): ResourceField[] {
  return fields
    .filter((field) => field.listable)
    .slice()
    .sort((a, b) => a.order - b.order || a.key.localeCompare(b.key));
}

export function listCellValue(values: Record<string, string>, key: string): string {
  const value = values[key];
  if (value == null || value.trim() === "") {
    return "—";
  }
  return value;
}

export function listCellBoolValue(values: Record<string, string>, key: string): boolean {
  const value = (values[key] ?? "").trim().toLowerCase();
  return value === "true" || value === "1" || value === "yes";
}

export function isListCheckboxField(field: ResourceField): boolean {
  return field.kind === ResourceFieldKind.CHECKBOX;
}
