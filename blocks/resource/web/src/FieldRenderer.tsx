import type { ResourceField, ResourceItem } from "./types";

export type FieldRendererProps = {
  field: ResourceField;
  value: unknown;
  item: ResourceItem;
  onChange: (key: string, value: unknown) => void;
  relationOptions?: Record<string, { value: string; label: string }[]>;
  readOnly?: boolean;
};

function stringValue(value: unknown): string {
  if (value == null) return "";
  if (typeof value === "string") return value;
  return String(value);
}

function numberValue(value: unknown): string {
  if (typeof value === "number") return String(value);
  return stringValue(value);
}

function boolValue(value: unknown): boolean {
  return value === true || value === "true" || value === 1;
}

export function FieldRenderer({
  field,
  value,
  onChange,
  relationOptions,
  readOnly,
}: FieldRendererProps) {
  const disabled = Boolean(readOnly || field.read_only);
  const commonProps = {
    id: `resource-field-${field.key}`,
    name: field.key,
    disabled,
    required: field.required,
  };

  if (field.type === "textarea" || field.type === "json") {
    return (
      <textarea
        {...commonProps}
        className="appkit-resource-field__control"
        value={field.type === "json" && typeof value !== "string" ? JSON.stringify(value ?? {}, null, 2) : stringValue(value)}
        onChange={(event) => onChange(field.key, event.target.value)}
        rows={field.type === "json" ? 8 : 4}
      />
    );
  }

  if (field.type === "bool") {
    return (
      <label className="appkit-resource-field__checkbox">
        <input
          {...commonProps}
          type="checkbox"
          checked={boolValue(value)}
          onChange={(event) => onChange(field.key, event.target.checked)}
        />
        <span>{field.label}</span>
      </label>
    );
  }

  if (field.type === "enum") {
    return (
      <select
        {...commonProps}
        className="appkit-resource-field__control"
        value={stringValue(value)}
        onChange={(event) => onChange(field.key, event.target.value)}
      >
        <option value="">Select...</option>
        {(field.options ?? []).map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }

  if (field.type === "relation") {
    const options = relationOptions?.[field.key] ?? [];
    return (
      <select
        {...commonProps}
        className="appkit-resource-field__control"
        value={stringValue(value)}
        onChange={(event) => onChange(field.key, event.target.value || null)}
      >
        <option value="">None</option>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    );
  }

  const inputType =
    field.type === "number" || field.type === "money"
      ? "number"
      : field.type === "date"
        ? "date"
        : field.type === "datetime"
          ? "datetime-local"
          : "text";

  return (
    <input
      {...commonProps}
      className="appkit-resource-field__control"
      type={inputType}
      value={inputType === "number" ? numberValue(value) : stringValue(value)}
      onChange={(event) => {
        const next = event.target.value;
        onChange(field.key, inputType === "number" && next !== "" ? Number(next) : next);
      }}
    />
  );
}
