"use client";

import type { ResourceField } from "./types";
import {
  buildManualLocationValue,
  locationMode,
  parseLocationValue,
  sourceLabel,
} from "./resource-location";
import { ResourceFieldLabel } from "./ResourceFieldLabel";

type ResourceLocationFieldProps = {
  field: ResourceField;
  value: string;
  error?: string | null;
  onChange: (key: string, value: string) => void;
};

export function ResourceLocationField({ field, value, error, onChange }: ResourceLocationFieldProps) {
  const parsed = parseLocationValue(value);
  const mode = locationMode(field);
  const isManual = mode === "manual" && field.editable && !field.readOnly;
  const lat = parsed?.lat;
  const lng = parsed?.lng;
  const hasCoords = lat != null && lng != null && Number.isFinite(lat) && Number.isFinite(lng);

  const bboxDelta = 0.012;
  const embedUrl = hasCoords
    ? `https://www.openstreetmap.org/export/embed.html?bbox=${lng - bboxDelta}%2C${lat - bboxDelta}%2C${lng + bboxDelta}%2C${lat + bboxDelta}&layer=mapnik&marker=${lat}%2C${lng}`
    : null;

  function updateManualCoords(nextLat: string, nextLng: string) {
    const latNum = Number.parseFloat(nextLat);
    const lngNum = Number.parseFloat(nextLng);
    if (!Number.isFinite(latNum) || !Number.isFinite(lngNum)) {
      onChange(field.key, "");
      return;
    }
    onChange(field.key, buildManualLocationValue(latNum, lngNum, parsed));
  }

  return (
    <div className="appkit-resource-edit-location">
      <ResourceFieldLabel field={field} htmlFor={`resource-field-${field.key}`} />

      {isManual ? (
        <div className="appkit-resource-edit-location__coords">
          <div>
            <label htmlFor={`${field.key}-lat`} className="appkit-resource-edit-location__coord-label">
              Latitude
            </label>
            <input
              id={`${field.key}-lat`}
              type="number"
              step="any"
              className="appkit-resource-edit-field"
              value={hasCoords ? lat : ""}
              onChange={(event) => updateManualCoords(event.target.value, hasCoords ? String(lng) : "")}
            />
          </div>
          <div>
            <label htmlFor={`${field.key}-lng`} className="appkit-resource-edit-location__coord-label">
              Longitude
            </label>
            <input
              id={`${field.key}-lng`}
              type="number"
              step="any"
              className="appkit-resource-edit-field"
              value={hasCoords ? lng : ""}
              onChange={(event) => updateManualCoords(hasCoords ? String(lat) : "", event.target.value)}
            />
          </div>
        </div>
      ) : null}

      {hasCoords && embedUrl ? (
        <>
          <div className="appkit-resource-edit-location__map">
            <iframe
              key={`${lat},${lng}`}
              title={field.label || "Location map"}
              src={embedUrl}
              loading="lazy"
            />
          </div>
          <p className="appkit-resource-edit-help">
            {lat.toFixed(5)}, {lng.toFixed(5)}
            {sourceLabel(parsed?.source) ? ` · ${sourceLabel(parsed?.source)}` : null}
          </p>
        </>
      ) : parsed?.status === "pending" ? (
        <div className="appkit-resource-edit-location__placeholder">
          Address saved. Location coordinates will appear after geocoding completes.
        </div>
      ) : (
        <div className="appkit-resource-edit-location__placeholder">
          {isManual
            ? "Enter coordinates to preview your location on the map."
            : "Add an address above to preview your business on the map."}
        </div>
      )}

      {field.helpText ? <p className="appkit-resource-edit-help">{field.helpText}</p> : null}
      {error ? (
        <p className="appkit-resource-edit-error" role="alert">
          {error}
        </p>
      ) : null}
    </div>
  );
}
