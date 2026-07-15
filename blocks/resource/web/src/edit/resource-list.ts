export function isResourceListInteractiveTarget(target: EventTarget | null): boolean {
  return Boolean(
    target instanceof Element &&
      target.closest("button, a, input, select, textarea, [role='menu'], [role='menuitem']"),
  );
}

export function resourceListRowClassName(selected: boolean): string {
  return selected
    ? "appkit-resource-edit-list__row appkit-resource-edit-list__row--selected"
    : "appkit-resource-edit-list__row";
}
