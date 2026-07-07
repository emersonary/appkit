import { useMemo, useState, type ReactNode } from "react";
import { fieldByKey, itemID, itemName, listColumns, nameField, parentIDField } from "./schema";
import type { ResourceColumn, ResourceItem, ResourceSchema } from "./types";

export type ResourceListProps = {
  schema: ResourceSchema;
  items: ResourceItem[];
  total?: number;
  page?: number;
  pageSize?: number;
  loading?: boolean;
  renderCell?: (props: {
    item: ResourceItem;
    column: ResourceColumn;
    value: unknown;
    depth: number;
  }) => ReactNode;
  onEdit?: (item: ResourceItem) => void;
  onPageChange?: (page: number) => void;
  onToggle?: (item: ResourceItem, expanded: boolean) => void;
};

type Row = {
  item: ResourceItem;
  id: string;
  parentID: string;
  depth: number;
  hasChildren: boolean;
};

function stringValue(value: unknown): string {
  if (value == null) return "";
  if (typeof value === "boolean") return value ? "Yes" : "No";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

function buildRows(schema: ResourceSchema, items: ResourceItem[], expanded: Set<string>): Row[] {
  const parentKey = parentIDField(schema);
  if (!parentKey) {
    return items.map((item) => ({
      item,
      id: itemID(schema, item),
      parentID: "",
      depth: 0,
      hasChildren: false,
    }));
  }

  const byParent = new Map<string, ResourceItem[]>();
  for (const item of items) {
    const parent = item[parentKey];
    const key = parent == null ? "" : String(parent);
    byParent.set(key, [...(byParent.get(key) ?? []), item]);
  }

  const rows: Row[] = [];
  const visit = (parentID: string, depth: number) => {
    for (const item of byParent.get(parentID) ?? []) {
      const id = itemID(schema, item);
      const hasChildren = (byParent.get(id) ?? []).length > 0;
      rows.push({ item, id, parentID, depth, hasChildren });
      if (hasChildren && expanded.has(id)) {
        visit(id, depth + 1);
      }
    }
  };
  visit("", 0);
  return rows;
}

export function ResourceList({
  schema,
  items,
  total,
  page = 1,
  pageSize = schema.list?.page_size ?? 25,
  loading,
  renderCell,
  onEdit,
  onPageChange,
  onToggle,
}: ResourceListProps) {
  const columns = useMemo(() => listColumns(schema), [schema]);
  const fields = useMemo(() => fieldByKey(schema), [schema]);
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set());
  const rows = useMemo(() => buildRows(schema, items, expanded), [schema, items, expanded]);
  const treeEnabled = Boolean(parentIDField(schema));
  const firstColumnKey = nameField(schema);
  const pageCount = total == null ? 1 : Math.max(1, Math.ceil(total / pageSize));

  const toggle = (row: Row) => {
    setExpanded((current) => {
      const next = new Set(current);
      const willExpand = !next.has(row.id);
      if (willExpand) {
        next.add(row.id);
      } else {
        next.delete(row.id);
      }
      onToggle?.(row.item, willExpand);
      return next;
    });
  };

  return (
    <section className="appkit-resource-list">
      <header className="appkit-resource-list__header">
        <div>
          <h1>{schema.name}</h1>
          {schema.description ? <p>{schema.description}</p> : null}
        </div>
        {total != null ? <span className="appkit-resource-list__count">{total} items</span> : null}
      </header>

      <div className="appkit-resource-table-wrap">
        <table className="appkit-resource-table">
          <thead>
            <tr>
              {columns.map((column) => {
                const field = fields.get(column.field_key);
                return (
                  <th key={column.field_key} style={column.width ? { width: column.width } : undefined}>
                    {column.label || field?.label || column.field_key}
                  </th>
                );
              })}
              {onEdit ? <th aria-label="Actions" /> : null}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={columns.length + (onEdit ? 1 : 0)}>Loading...</td>
              </tr>
            ) : null}
            {!loading && rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length + (onEdit ? 1 : 0)}>No {schema.name.toLowerCase()} found.</td>
              </tr>
            ) : null}
            {!loading
              ? rows.map((row) => (
                  <tr key={row.id || itemName(schema, row.item)} className="appkit-resource-table__row">
                    {columns.map((column) => {
                      const value = row.item[column.field_key];
                      const isTreeColumn = treeEnabled && column.field_key === firstColumnKey;
                      return (
                        <td key={column.field_key}>
                          <div
                            className={isTreeColumn ? "appkit-resource-tree-cell" : undefined}
                            style={isTreeColumn ? { paddingLeft: `${row.depth * 1.25}rem` } : undefined}
                          >
                            {isTreeColumn && row.hasChildren ? (
                              <button
                                type="button"
                                className="appkit-resource-tree-toggle"
                                aria-label={`${expanded.has(row.id) ? "Collapse" : "Expand"} ${itemName(schema, row.item)}`}
                                aria-expanded={expanded.has(row.id)}
                                onClick={() => toggle(row)}
                              >
                                {expanded.has(row.id) ? "−" : "+"}
                              </button>
                            ) : isTreeColumn ? (
                              <span className="appkit-resource-tree-spacer" />
                            ) : null}
                            <span>
                              {renderCell
                                ? renderCell({ item: row.item, column, value, depth: row.depth })
                                : stringValue(value)}
                            </span>
                          </div>
                        </td>
                      );
                    })}
                    {onEdit ? (
                      <td className="appkit-resource-table__actions">
                        <button type="button" className="appkit-resource-button appkit-resource-button--ghost" onClick={() => onEdit(row.item)}>
                          Edit
                        </button>
                      </td>
                    ) : null}
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </div>

      {onPageChange && pageCount > 1 ? (
        <footer className="appkit-resource-list__pager">
          <button
            type="button"
            className="appkit-resource-button appkit-resource-button--ghost"
            disabled={page <= 1}
            onClick={() => onPageChange(page - 1)}
          >
            Previous
          </button>
          <span>
            Page {page} of {pageCount}
          </span>
          <button
            type="button"
            className="appkit-resource-button appkit-resource-button--ghost"
            disabled={page >= pageCount}
            onClick={() => onPageChange(page + 1)}
          >
            Next
          </button>
        </footer>
      ) : null}
    </section>
  );
}
