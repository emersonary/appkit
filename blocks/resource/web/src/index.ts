/**
 * Schema-driven resource edit (protobuf ResourceEditResponse contract).
 * Primary API for host apps like Solidia Command Center.
 */
export * from "./edit";

/**
 * Legacy registry-driven list/edit UI (FieldType schema).
 * Prefer the edit exports above for new work.
 */
export { FieldRenderer, type FieldRendererProps } from "./FieldRenderer";
export {
  ResourceEdit as LegacyRegistryResourceEdit,
  type ResourceEditProps as LegacyRegistryResourceEditProps,
  type ResourceFieldInterceptProps,
} from "./ResourceEdit";
export {
  ResourceList,
  type ResourceCellInterceptProps,
  type ResourceListProps,
  type ResourceRowAction,
} from "./ResourceList";
export {
  ResourceListAndEdit as LegacyRegistryResourceListAndEdit,
  type ResourceEditConfig as LegacyRegistryResourceEditConfig,
  type ResourceListAndEditActions as LegacyRegistryResourceListAndEditActions,
  type ResourceListAndEditProps as LegacyRegistryResourceListAndEditProps,
} from "./ResourceListAndEdit";
export {
  fieldByKey,
  formSections,
  idField,
  itemID,
  itemName,
  listColumns,
  nameField,
  parentIDField,
} from "./schema";
export type {
  FieldOption,
  FieldType,
  Relation,
  ResourceClient,
  ResourceColumn,
  ResourceField as LegacyResourceField,
  ResourceFormSection,
  ResourceFormView,
  ResourceEditMode,
  ResourceItem,
  ResourceListResponse,
  ResourceListView,
  ResourceSchema as LegacyResourceSchema,
  TreeConfig,
} from "./types";
