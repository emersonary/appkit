import {
  ResourceFieldKind,
  ResourceValidationKind,
  type ResourceField,
  type ResourceSchema,
} from "./types";
import { changedEditableFieldKeys, editableFields } from "./resource-edit";
import { isLocationField, locationMode, validateManualLocationValue } from "./resource-location";

export function normalizeUrlInput(value: string): string {
  const trimmed = value.trim();
  if (!trimmed) return "";

  try {
    const parsed = new URL(trimmed);
    if (parsed.protocol === "http:" || parsed.protocol === "https:") {
      return trimmed;
    }
    return trimmed;
  } catch {
    return `https://${trimmed.replace(/^\/+/, "")}`;
  }
}

export function isValidHttpUrl(value: string): boolean {
  const trimmed = value.trim();
  if (!trimmed) return true;

  try {
    const candidate = trimmed.includes("://") ? trimmed : normalizeUrlInput(trimmed);
    const parsed = new URL(candidate);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return false;
    return parsed.hostname.length > 0;
  } catch {
    return false;
  }
}

function fieldHasUrlValidation(field: ResourceField): boolean {
  if (field.kind === ResourceFieldKind.URL) return true;
  return field.validations.some((rule) => rule.kind === ResourceValidationKind.URL);
}

export function normalizeSubmitValues(
  schema: ResourceSchema,
  values: Record<string, string>,
): Record<string, string> {
  const out = { ...values };
  for (const field of editableFields(schema.fields)) {
    if (!fieldHasUrlValidation(field)) continue;
    const raw = out[field.key];
    if (raw == null || raw.trim() === "") {
      out[field.key] = "";
      continue;
    }
    out[field.key] = normalizeUrlInput(raw);
  }
  return out;
}

function validateField(field: ResourceField, value: string, errors: Record<string, string>): void {
  if (field.required && !value) {
    errors[field.key] = `${field.label} is required`;
    return;
  }

  for (const rule of field.validations) {
    const message = validateRule(field, value, rule.kind, rule.param);
    if (message) {
      errors[field.key] = message;
      return;
    }
  }

  if (field.kind === ResourceFieldKind.COUNTRY && value) {
    const allowed = new Set(field.options.map((option) => option.value));
    if (!allowed.has(value)) {
      errors[field.key] = "Select a valid country";
    }
    return;
  }

  if (field.options.length > 0 && field.kind === ResourceFieldKind.TEXT && value) {
    const allowed = new Set(field.options.map((option) => option.value ?? ""));
    if (!allowed.has(value)) {
      errors[field.key] = `Select a valid ${field.label.toLowerCase()}`;
    }
    return;
  }

  if (isLocationField(field) && value && locationMode(field) === "manual") {
    const message = validateManualLocationValue(value);
    if (message) errors[field.key] = message;
    return;
  }

  if (fieldHasUrlValidation(field) && value && !isValidHttpUrl(value)) {
    errors[field.key] = `${field.label} must be a valid URL`;
  }
}

export function validateResourceValues(
  schema: ResourceSchema,
  values: Record<string, string>,
  onlyKeys?: Set<string>,
): Record<string, string> {
  const errors: Record<string, string> = {};

  for (const field of editableFields(schema.fields)) {
    if (onlyKeys && !onlyKeys.has(field.key)) continue;
    const value = (values[field.key] ?? "").trim();
    validateField(field, value, errors);
  }

  return errors;
}

export function validateDirtyFields(
  schema: ResourceSchema,
  draft: Record<string, string>,
  baseline: Record<string, string>,
): Record<string, string> {
  const dirtyKeys = changedEditableFieldKeys(schema, draft, baseline);
  if (dirtyKeys.size === 0) return {};
  return validateResourceValues(schema, draft, dirtyKeys);
}

function validateRule(
  field: ResourceField,
  value: string,
  kind: ResourceValidationKind,
  param: string,
): string | null {
  switch (kind) {
    case ResourceValidationKind.REQUIRED:
      return value ? null : `${field.label} is required`;
    case ResourceValidationKind.EMAIL:
      if (!value) return null;
      return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) ? null : `${field.label} must be a valid email`;
    case ResourceValidationKind.URL:
      if (!value) return null;
      return isValidHttpUrl(value) ? null : `${field.label} must be a valid URL`;
    case ResourceValidationKind.MAX_LENGTH: {
      const limit = Number(param);
      if (!Number.isFinite(limit) || limit <= 0) return null;
      return value.length <= limit ? null : `${field.label} must be at most ${limit} characters`;
    }
    case ResourceValidationKind.PATTERN:
      if (!value || !param) return null;
      try {
        return new RegExp(param).test(value) ? null : `${field.label} has invalid format`;
      } catch {
        return null;
      }
    default:
      return null;
  }
}

export function firstValidationError(errors: Record<string, string>): string | null {
  for (const message of Object.values(errors)) {
    if (message) return message;
  }
  return null;
}

export function mapSubmitErrorToFieldErrors(
  schema: ResourceSchema,
  message: string,
): { formError: string | null; fieldErrors: Record<string, string> } {
  const normalized = message.trim().toLowerCase();
  if (
    normalized.includes("name already exists") ||
    (normalized.includes("duplicate") && normalized.includes("name"))
  ) {
    const hasNameField = schema.fields.some((field) => field.key === "name");
    if (hasNameField) {
      return { formError: null, fieldErrors: { name: message.trim() } };
    }
  }
  return { formError: message.trim(), fieldErrors: {} };
}
