"use client";

import { ResourceFieldKind, type ResourceField } from "./types";
import { isFieldRequired, resourceInputClassName, resourceReadonlyClassName } from "./resource-edit";
import { isLocationField } from "./resource-location";
import { ResourceFieldLabel } from "./ResourceFieldLabel";
import { ResourceImageField } from "./ResourceImageField";
import { ResourceLocationField } from "./ResourceLocationField";

type ResourceFieldInputProps = {
  field: ResourceField;
  value: string;
  error?: string | null;
  onChange: (key: string, value: string) => void;
  onUploadImage?: (file: File) => Promise<string>;
};

export function ResourceFieldInput({
  field,
  value,
  error,
  onChange,
  onUploadImage,
}: ResourceFieldInputProps) {
  const inputId = `resource-field-${field.key}`;
  const readOnly = field.readOnly || !field.editable;
  const required = isFieldRequired(field);
  const describedBy = [field.helpText ? `${inputId}-help` : null, error ? `${inputId}-error` : null]
    .filter(Boolean)
    .join(" ");

  if (isLocationField(field)) {
    return <ResourceLocationField field={field} value={value} error={error} onChange={onChange} />;
  }

  if (field.kind === ResourceFieldKind.IMAGE) {
    return (
      <ResourceImageField
        field={field}
        value={value}
        error={error}
        readOnly={readOnly}
        onChange={onChange}
        onUploadImage={onUploadImage}
      />
    );
  }

  if (field.kind === ResourceFieldKind.COUNTRY || (field.options.length > 0 && field.kind === ResourceFieldKind.TEXT)) {
    const options = field.options ?? [];
    const hasExplicitUnset = options.some((option) => {
      const optionValue = option.value ?? "";
      return optionValue === "" || optionValue === "none";
    });

    return (
      <div>
        <ResourceFieldLabel field={field} htmlFor={inputId} />
        <select
          id={inputId}
          name={field.key}
          value={value}
          disabled={readOnly}
          required={required}
          aria-describedby={describedBy || undefined}
          className={readOnly ? resourceReadonlyClassName : resourceInputClassName}
          onChange={(event) => onChange(field.key, event.target.value)}
        >
          {!hasExplicitUnset ? (
            <option value="">
              {field.kind === ResourceFieldKind.COUNTRY ? "Select a country" : "Select an option"}
            </option>
          ) : null}
          {options.map((option) => (
            <option key={`${option.value || "__empty"}-${option.label}`} value={option.value ?? ""}>
              {option.label}
            </option>
          ))}
        </select>
        {field.helpText ? (
          <p id={`${inputId}-help`} className="appkit-resource-edit-help">
            {field.helpText}
          </p>
        ) : null}
        {error ? (
          <p id={`${inputId}-error`} className="appkit-resource-edit-error">
            {error}
          </p>
        ) : null}
      </div>
    );
  }

  const commonProps = {
    id: inputId,
    name: field.key,
    value,
    readOnly,
    required,
    placeholder: field.placeholder || undefined,
    "aria-describedby": describedBy || undefined,
    "aria-invalid": error ? true : undefined,
  };

  let control: React.ReactNode;
  switch (field.kind) {
    case ResourceFieldKind.TEXTAREA:
      control = (
        <textarea
          {...commonProps}
          rows={4}
          className={`${readOnly ? resourceReadonlyClassName : resourceInputClassName} appkit-resource-edit-field--textarea`}
          onChange={(event) => onChange(field.key, event.target.value)}
        />
      );
      break;
    case ResourceFieldKind.CHECKBOX:
      control = (
        <label className="appkit-resource-edit-checkbox">
          <input
            type="checkbox"
            id={inputId}
            name={field.key}
            checked={value === "true"}
            disabled={readOnly}
            className="appkit-resource-edit-checkbox__input"
            onChange={(event) => onChange(field.key, event.target.checked ? "true" : "false")}
          />
          <span>
            {field.label}
            {required ? (
              <span className="appkit-resource-edit-label__required" aria-hidden="true">
                {" *"}
              </span>
            ) : null}
          </span>
        </label>
      );
      break;
    default:
      control = (
        <input
          {...commonProps}
          type={inputType(field.kind)}
          className={readOnly ? resourceReadonlyClassName : resourceInputClassName}
          onChange={(event) => onChange(field.key, event.target.value)}
        />
      );
  }

  if (field.kind === ResourceFieldKind.CHECKBOX) {
    return (
      <div>
        {control}
        {field.helpText ? (
          <p id={`${inputId}-help`} className="appkit-resource-edit-help">
            {field.helpText}
          </p>
        ) : null}
        {error ? (
          <p id={`${inputId}-error`} className="appkit-resource-edit-error">
            {error}
          </p>
        ) : null}
      </div>
    );
  }

  return (
    <div>
      <ResourceFieldLabel field={field} htmlFor={inputId} />
      {control}
      {field.helpText ? (
        <p id={`${inputId}-help`} className="appkit-resource-edit-help">
          {field.helpText}
        </p>
      ) : null}
      {error ? (
        <p id={`${inputId}-error`} className="appkit-resource-edit-error">
          {error}
        </p>
      ) : null}
    </div>
  );
}

function inputType(kind: ResourceFieldKind): string {
  switch (kind) {
    case ResourceFieldKind.EMAIL:
      return "email";
    case ResourceFieldKind.PHONE:
      return "tel";
    case ResourceFieldKind.URL:
    case ResourceFieldKind.IMAGE:
      return "url";
    case ResourceFieldKind.NUMBER:
      return "number";
    case ResourceFieldKind.DATETIME:
      return "text";
    default:
      return "text";
  }
}
