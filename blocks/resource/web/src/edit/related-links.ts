import type { ResourceRelatedLink } from "./types";

export function mapResourceRelatedLinks(
  links: Array<{
    label?: string;
    route?: string;
    icon?: string;
    description?: string;
  }> | undefined,
): ResourceRelatedLink[] {
  if (!links?.length) {
    return [];
  }
  return links
    .filter((link) => link.label && link.route)
    .map((link) => ({
      label: link.label!,
      route: link.route!,
      icon: link.icon || undefined,
      description: link.description || undefined,
    }));
}
