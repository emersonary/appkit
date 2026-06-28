export type BlogContentKind = "text" | "video";

export function normalizeBlogContentKind(kind: string): BlogContentKind {
  return kind === "video" ? "video" : "text";
}
