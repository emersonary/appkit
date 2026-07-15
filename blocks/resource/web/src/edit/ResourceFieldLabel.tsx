"use client";

import type { ResourceField } from "./types";
import { isFieldRequired } from "./resource-edit";

type ResourceFieldLabelProps = {
  field: ResourceField;
  htmlFor?: string;
};

export function ResourceFieldLabel({ field, htmlFor }: ResourceFieldLabelProps) {
  const className = "appkit-resource-edit-label";
  const content = (
    <>
      {field.label}
      {isFieldRequired(field) ? (
        <span className="appkit-resource-edit-label__required" aria-hidden="true">
          {" *"}
        </span>
      ) : null}
    </>
  );

  if (htmlFor) {
    return (
      <label htmlFor={htmlFor} className={className}>
        {content}
      </label>
    );
  }

  return <div className={className}>{content}</div>;
}
