type ResourceListBoolCellProps = {
  checked: boolean;
  label: string;
};

/** Read-only checkbox for boolean columns on resource list tables. */
export function ResourceListBoolCell({ checked, label }: ResourceListBoolCellProps) {
  return (
    <span className="appkit-resource-edit-list__bool-cell">
      <input
        type="checkbox"
        checked={checked}
        readOnly
        tabIndex={-1}
        aria-label={label}
        aria-readonly="true"
        onChange={() => {}}
        className="appkit-resource-edit-checkbox__input appkit-resource-edit-checkbox__input--list"
      />
    </span>
  );
}
