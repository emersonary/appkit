import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from "react";
import { FieldRenderer } from "./FieldRenderer";
import { fieldByKey, formSections, itemID } from "./schema";
import type { ResourceField, ResourceItem, ResourceSchema } from "./types";

export type ResourceEditFormProps = {
  schema: ResourceSchema;
  item?: ResourceItem;
  relationOptions?: Record<string, { value: string; label: string }[]>;
  saving?: boolean;
  readOnly?: boolean;
  submitLabel?: string;
  renderField?: (props: {
    field: ResourceField;
    value: unknown;
    item: ResourceItem;
    onChange: (key: string, value: unknown) => void;
    readOnly?: boolean;
  }) => ReactNode;
  onSubmit?: (values: ResourceItem) => void | Promise<void>;
  onCancel?: () => void;
};

export function ResourceEditForm({
  schema,
  item,
  relationOptions,
  saving,
  readOnly,
  submitLabel = "Save",
  renderField,
  onSubmit,
  onCancel,
}: ResourceEditFormProps) {
  const [values, setValues] = useState<ResourceItem>(() => ({ ...(item ?? {}) }));
  const fields = useMemo(() => fieldByKey(schema), [schema]);
  const sections = useMemo(() => formSections(schema), [schema]);
  const editing = Boolean(itemID(schema, values));

  useEffect(() => {
    setValues({ ...(item ?? {}) });
  }, [item]);

  const updateValue = (key: string, value: unknown) => {
    setValues((current) => ({ ...current, [key]: value }));
  };

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    await onSubmit?.(values);
  };

  return (
    <form className="appkit-resource-form" onSubmit={submit}>
      <header className="appkit-resource-form__header">
        <div>
          <p className="appkit-resource-form__eyebrow">{editing ? "Edit" : "Create"}</p>
          <h1 className="appkit-resource-form__title">{schema.name}</h1>
          {schema.description ? (
            <p className="appkit-resource-form__description">{schema.description}</p>
          ) : null}
        </div>
      </header>

      {sections.map((section) => (
        <section key={section.id} className="appkit-resource-form__section">
          <div className="appkit-resource-form__section-head">
            <h2>{section.title}</h2>
            {section.description ? <p>{section.description}</p> : null}
          </div>
          <div className="appkit-resource-form__fields">
            {(section.fields ?? []).map((fieldKey) => {
              const field = fields.get(fieldKey);
              if (!field || field.hidden || field.form_hidden) {
                return null;
              }
              return (
                <label key={field.key} className="appkit-resource-field">
                  {field.type !== "bool" ? (
                    <span className="appkit-resource-field__label">
                      {field.label}
                      {field.required ? <span aria-hidden="true"> *</span> : null}
                    </span>
                  ) : null}
                  {renderField ? (
                    renderField({
                      field,
                      value: values[field.key],
                      item: values,
                      onChange: updateValue,
                      readOnly,
                    })
                  ) : (
                    <FieldRenderer
                      field={field}
                      value={values[field.key]}
                      item={values}
                      onChange={updateValue}
                      relationOptions={relationOptions}
                      readOnly={readOnly}
                    />
                  )}
                  {field.help_text ? (
                    <span className="appkit-resource-field__help">{field.help_text}</span>
                  ) : null}
                </label>
              );
            })}
          </div>
        </section>
      ))}

      <footer className="appkit-resource-form__actions">
        {onCancel ? (
          <button type="button" className="appkit-resource-button appkit-resource-button--ghost" onClick={onCancel}>
            Cancel
          </button>
        ) : null}
        {!readOnly ? (
          <button type="submit" className="appkit-resource-button" disabled={saving}>
            {saving ? "Saving..." : submitLabel}
          </button>
        ) : null}
      </footer>
    </form>
  );
}
