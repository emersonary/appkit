export type FieldType =
  | "uuid"
  | "text"
  | "textarea"
  | "number"
  | "money"
  | "bool"
  | "date"
  | "datetime"
  | "enum"
  | "relation"
  | "object";

export type FieldOption = {
  value: string;
  label: string;
};

export type Relation = {
  resource_id: string;
  value_field?: string;
  display_field?: string;
};

export type ResourceField = {
  key: string;
  label: string;
  type: FieldType;
  section?: string;
  help_text?: string;
  required?: boolean;
  read_only?: boolean;
  hidden?: boolean;
  list_hidden?: boolean;
  form_hidden?: boolean;
  sort_order?: number;
  options?: FieldOption[];
  relation?: Relation;
};

export type ResourceColumn = {
  field_key: string;
  label?: string;
  width?: string;
  sort_order?: number;
  hidden?: boolean;
};

export type TreeConfig = {
  enabled?: boolean;
  parent_id_field?: string;
  name_field?: string;
  lazy_load?: boolean;
};

export type ResourceListView = {
  page_size?: number;
  columns?: ResourceColumn[];
  tree?: TreeConfig;
  searchable_fields?: string[];
};

export type ResourceFormSection = {
  id: string;
  title: string;
  description?: string;
  fields?: string[];
  sort_order?: number;
};

export type ResourceFormView = {
  sections?: ResourceFormSection[];
};

export type ResourceSchema = {
  id: string;
  name: string;
  description?: string;
  id_field?: string;
  name_field?: string;
  parent_id_field?: string;
  fields: ResourceField[];
  list?: ResourceListView;
  form?: ResourceFormView;
};

export type ResourceItem = Record<string, unknown>;

export type ResourceEditMode = "create" | "edit" | "replicate";

export type ResourceListResponse = {
  items: ResourceItem[];
  total: number;
  page: number;
  page_size: number;
  has_more: boolean;
};

export type ResourceClient = {
  list: (resourceID: string, request: {
    page?: number;
    page_size?: number;
    query?: string;
    parent_id?: string | null;
  }) => Promise<ResourceListResponse>;
  get: (resourceID: string, id: string) => Promise<ResourceItem>;
  create: (resourceID: string, values: ResourceItem) => Promise<ResourceItem>;
  update: (resourceID: string, id: string, values: ResourceItem) => Promise<ResourceItem>;
  delete: (resourceID: string, id: string) => Promise<void>;
};
