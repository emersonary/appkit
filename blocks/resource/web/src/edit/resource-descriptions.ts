import type { ReactNode } from "react";
import type { ResourceEditMode } from "./handlers";

/**
 * Optional full-string overrides for generated list/form copy.
 * Where each key appears:
 * - list — list header blurb above the table
 * - emptyTitle — empty-state heading when there are no rows
 * - empty — empty-state body under emptyTitle
 * - create — intro on the create form
 * - edit — intro on the edit form
 * - replicate — intro on the replicate form
 * - createLabel — primary "Add …" button (list header + empty state)
 * - backLabel — back button on the edit/create/replicate form
 * - replicateLabel — "Replicate" item in the row ⋯ menu
 * - deleteLabel — "Delete" item in the row ⋯ menu
 */
export type ResourceListAndEditDescriptionOverrides = {
  list?: ReactNode;
  emptyTitle?: string;
  empty?: ReactNode;
  create?: ReactNode;
  edit?: ReactNode;
  replicate?: ReactNode;
  createLabel?: string;
  backLabel?: string;
  replicateLabel?: string;
  deleteLabel?: string;
};

/** Resource naming + optional full-string overrides for list/form copy. */
export type ResourceListAndEditDescriptions = {
  singularName: string;
  pluralName: string;
  override?: ResourceListAndEditDescriptionOverrides;
};

export type ResolvedResourceListAndEditCopy = {
  list: ReactNode;
  emptyTitle: string;
  empty: ReactNode;
  create: ReactNode;
  edit: ReactNode;
  replicate: ReactNode;
  createLabel: string;
  backLabel: string;
  replicateLabel?: string;
  deleteLabel?: string;
};

function indefiniteArticle(noun: string): "a" | "an" {
  const first = noun.trim().charAt(0).toLowerCase();
  return "aeiou".includes(first) ? "an" : "a";
}

function lowerName(value: string): string {
  if (!value) return value;
  return value.charAt(0).toLowerCase() + value.slice(1);
}

/** Capitalize the first letter (for shell / page titles). */
export function titleCaseName(value: string): string {
  const trimmed = value.trim();
  if (!trimmed) return trimmed;
  return trimmed.charAt(0).toUpperCase() + trimmed.slice(1);
}

/** Default list page title from `descriptions.pluralName`. */
export function resourceListTitle(descriptions: ResourceListAndEditDescriptions): string {
  return titleCaseName(descriptions.pluralName);
}

/** Default singular title from `descriptions.singularName` (edit fallback). */
export function resourceSingularTitle(descriptions: ResourceListAndEditDescriptions): string {
  return titleCaseName(descriptions.singularName);
}

/**
 * Resolve shell/page title while editing vs listing.
 * Host still owns where the title is displayed (e.g. app shell).
 */
export function titleFromEditingState(
  editing: boolean,
  state: { schema?: { label?: string } } | null,
  options: { listTitle: string; fallbackSingular?: string },
): string {
  if (!editing) {
    return options.listTitle;
  }
  const label = state?.schema?.label?.trim();
  if (label) {
    return label;
  }
  return options.fallbackSingular?.trim() || options.listTitle;
}

export function resolveResourceListAndEditCopy(
  descriptions: ResourceListAndEditDescriptions | undefined,
): ResolvedResourceListAndEditCopy | null {
  if (!descriptions) {
    return null;
  }

  const singularRaw = descriptions.singularName.trim();
  const pluralRaw = descriptions.pluralName.trim();
  // Names follow leading text in every default template, so lower-case the first letter.
  const singular = lowerName(singularRaw);
  const plural = lowerName(pluralRaw);
  const override = descriptions.override ?? {};
  const article = indefiniteArticle(singularRaw);

  return {
    list: override.list ?? `Manage ${plural}.`,
    emptyTitle: override.emptyTitle ?? `No ${plural} yet`,
    empty: override.empty ?? `Add your first ${singular}.`,
    create: override.create ?? `Add ${article} ${singular}.`,
    edit: override.edit ?? `Edit this ${singular}.`,
    replicate:
      override.replicate ??
      `Review the copied ${singular} details, change anything you need, then create the new ${singular}.`,
    createLabel: override.createLabel ?? `Add ${singular}`,
    backLabel: override.backLabel ?? `Back to ${plural}`,
    replicateLabel: override.replicateLabel,
    deleteLabel: override.deleteLabel,
  };
}

export function resolveModeDescription(
  copy: ResolvedResourceListAndEditCopy | null,
  mode: ResourceEditMode,
): ReactNode | undefined {
  if (!copy) {
    return undefined;
  }
  return copy[mode];
}
