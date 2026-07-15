/** Client-only address geocode for map preview (no backend). Uses Photon (CORS-friendly). */

export type GeocodePreviewResult = {
  lat: number;
  lng: number;
  label?: string;
};

export async function geocodeAddressPreview(
  query: string,
  signal?: AbortSignal,
): Promise<GeocodePreviewResult | null> {
  const q = query.trim();
  if (!q) {
    return null;
  }

  const url =
    "https://photon.komoot.io/api/?" +
    new URLSearchParams({
      q,
      limit: "1",
    }).toString();

  const response = await fetch(url, { signal });
  if (!response.ok) {
    return null;
  }

  const payload = (await response.json()) as {
    features?: Array<{
      geometry?: { coordinates?: number[] };
      properties?: {
        name?: string;
        street?: string;
        housenumber?: string;
        city?: string;
        state?: string;
        country?: string;
      };
    }>;
  };
  const first = payload.features?.[0];
  const coordinates = first?.geometry?.coordinates;
  if (!coordinates || coordinates.length < 2) {
    return null;
  }
  const lng = coordinates[0];
  const lat = coordinates[1];
  if (!Number.isFinite(lat) || !Number.isFinite(lng)) {
    return null;
  }

  const props = first?.properties ?? {};
  const street = [props.housenumber, props.street || props.name].filter(Boolean).join(" ");
  const labelParts = [street || props.name, props.city, props.state, props.country].filter(
    (part): part is string => Boolean(part && String(part).trim()),
  );

  return {
    lat,
    lng,
    label: labelParts.length > 0 ? labelParts.join(", ") : q,
  };
}
