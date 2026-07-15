import { ResourceFieldKind, type ResourceField } from "./types";

export type LocationFieldValue = {
  lat?: number;
  lng?: number;
  status: string;
  source?: string;
  address_label?: string;
};

export function parseLocationValue(raw: string): LocationFieldValue | null {
  const trimmed = raw.trim();
  if (!trimmed) return null;
  try {
    const parsed = JSON.parse(trimmed) as LocationFieldValue;
    if (!parsed || typeof parsed.status !== "string") return null;
    return parsed;
  } catch {
    return null;
  }
}

export function stringifyLocationValue(value: LocationFieldValue): string {
  return JSON.stringify(value);
}

export function isLocationField(field: ResourceField): boolean {
  return field.kind === ResourceFieldKind.LOCATION;
}

export function locationMode(field: ResourceField): "preview" | "bound" | "manual" {
  const mode = field.locationMode?.trim().toLowerCase();
  if (mode === "manual" || mode === "bound" || mode === "preview") {
    return mode;
  }
  if (field.readOnly || !field.editable) {
    return "preview";
  }
  return "manual";
}

/** Sections watched for client-side map preview (not for server autosave). */
export function locationPreviewWatchSections(fields: ResourceField[]): Set<string> {
  const sections = new Set<string>();
  for (const field of fields) {
    if (!isLocationField(field)) continue;
    const mode = locationMode(field);
    if (mode === "manual") continue;
    const watch = field.watchSection || field.bindSection;
    if (watch) sections.add(watch);
  }
  return sections;
}

export function fieldTriggersLocationPreview(fields: ResourceField[], changedKey: string): boolean {
  const changed = fields.find((field) => field.key === changedKey);
  if (!changed || !changed.editable || changed.readOnly) return false;
  return locationPreviewWatchSections(fields).has(changed.section || "general");
}

export function previewLocationFieldKeys(fields: ResourceField[]): string[] {
  return fields
    .filter((field) => isLocationField(field) && locationMode(field) !== "manual")
    .map((field) => field.key);
}

export function buildAddressQueryFromDraft(
  fields: ResourceField[],
  draft: Record<string, string>,
  section: string,
): string {
  const parts: string[] = [];
  for (const field of fields) {
    if (!field.editable || field.readOnly) continue;
    if ((field.section || "general") !== section) continue;
    if (isLocationField(field)) continue;
    const value = (draft[field.key] ?? "").trim();
    if (value) parts.push(value);
  }
  return parts.join(", ");
}

export function buildPreviewLocationValue(
  lat: number,
  lng: number,
  addressLabel: string,
): string {
  return stringifyLocationValue({
    lat,
    lng,
    status: "verified",
    source: "geocoded",
    address_label: addressLabel,
  });
}

export function buildPendingLocationValue(addressLabel: string): string {
  return stringifyLocationValue({
    status: "pending",
    address_label: addressLabel,
  });
}

export function sourceLabel(source?: string): string | null {
  switch (source) {
    case "geocoded":
      return "Geocoded from address";
    case "imported_gbp":
      return "Imported from Google Business Profile";
    case "manual":
      return "Manually placed";
    default:
      return null;
  }
}

export function buildManualLocationValue(
  lat: number,
  lng: number,
  previous?: LocationFieldValue | null,
): string {
  return stringifyLocationValue({
    lat,
    lng,
    status: "verified",
    source: "manual",
    address_label: previous?.address_label,
  });
}

export function validateManualLocationValue(raw: string): string | null {
  const value = parseLocationValue(raw);
  if (!raw.trim()) return null;
  if (!value?.lat || !value?.lng) return "Enter latitude and longitude";
  if (value.lat < -90 || value.lat > 90) return "Latitude must be between -90 and 90";
  if (value.lng < -180 || value.lng > 180) return "Longitude must be between -180 and 180";
  return null;
}
