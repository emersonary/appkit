"use client";

import { useRef, useState } from "react";
import type { ResourceField } from "./types";
import { validateImageDimensions } from "./resource-edit";
import { ResourceFieldLabel } from "./ResourceFieldLabel";

type ResourceImageFieldProps = {
  field: ResourceField;
  value: string;
  error?: string | null;
  readOnly: boolean;
  onChange: (key: string, value: string) => void;
  onUploadImage?: (file: File) => Promise<string>;
};

export function ResourceImageField({
  field,
  value,
  error,
  readOnly,
  onChange,
  onUploadImage,
}: ResourceImageFieldProps) {
  const inputId = `resource-field-${field.key}`;
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = "";
    if (!file || readOnly || !onUploadImage) return;

    setUploadError(null);
    setUploading(true);
    try {
      const dimensionError = await validateImageDimensions(field, file);
      if (dimensionError) {
        setUploadError(dimensionError);
        return;
      }
      const url = await onUploadImage(file);
      onChange(field.key, url);
    } catch (err) {
      setUploadError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setUploading(false);
    }
  }

  return (
    <div>
      <ResourceFieldLabel field={field} />
      <div className="appkit-resource-edit-image">
        {value ? (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img src={value} alt={`${field.label} preview`} className="appkit-resource-edit-image__preview" />
        ) : null}
        {!readOnly && onUploadImage ? (
          <>
            <input
              ref={fileRef}
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              className="appkit-resource-edit-image__file"
              onChange={handleFileChange}
            />
            <button
              type="button"
              disabled={uploading}
              className="appkit-resource-edit-button appkit-resource-edit-button--secondary"
              onClick={() => fileRef.current?.click()}
            >
              {uploading ? "Uploading…" : value ? "Replace image" : "Upload image"}
            </button>
          </>
        ) : null}
        {value && !readOnly ? (
          <button
            type="button"
            className="appkit-resource-edit-image__remove"
            onClick={() => onChange(field.key, "")}
          >
            Remove
          </button>
        ) : null}
      </div>
      {uploadError ? <p className="appkit-resource-edit-error">{uploadError}</p> : null}
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
