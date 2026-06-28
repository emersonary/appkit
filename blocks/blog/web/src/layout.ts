export type BlogLayoutCode = "one_column" | "two_column";

const twoColumnLayouts = new Set(["two_column", "two_column_image_right"]);

export function normalizeBlogLayoutType(layoutType: string): BlogLayoutCode {
  if (twoColumnLayouts.has(layoutType)) {
    return "two_column";
  }
  return "one_column";
}

export function isTwoColumnLayout(layoutType: string): boolean {
  return twoColumnLayouts.has(layoutType);
}
