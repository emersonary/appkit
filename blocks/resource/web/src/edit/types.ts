/** Wire contract for schema-driven resource edit forms (matches solidia.v1.resource proto). */

export enum ResourceFieldKind {
  UNSPECIFIED = 0,
  TEXT = 1,
  TEXTAREA = 2,
  EMAIL = 3,
  PHONE = 4,
  URL = 5,
  IMAGE = 6,
  COUNTRY = 7,
  CHECKBOX = 8,
  NUMBER = 9,
  DATETIME = 10,
  LOCATION = 11,
}

export enum ResourceValidationKind {
  UNSPECIFIED = 0,
  REQUIRED = 1,
  EMAIL = 2,
  URL = 3,
  MAX_LENGTH = 4,
  PATTERN = 5,
}

export enum ResourceRecordState {
  UNSPECIFIED = 0,
  NEW = 1,
  EXISTING = 2,
}

export enum ResourceMode {
  UNSPECIFIED = 0,
  EDIT_ONLY = 1,
  LIST_AND_EDIT = 2,
}

export type ResourceFieldOption = {
  value: string;
  label: string;
};

export type ResourceValidationRule = {
  kind: ResourceValidationKind;
  param: string;
};

export type ResourceField = {
  key: string;
  label: string;
  kind: ResourceFieldKind;
  section: string;
  order: number;
  required: boolean;
  readOnly: boolean;
  editable: boolean;
  visible: boolean;
  listable: boolean;
  placeholder: string;
  helpText: string;
  options: ResourceFieldOption[];
  validations: ResourceValidationRule[];
  maxWidth: number;
  maxHeight: number;
  watchSection: string;
  bindSection: string;
  locationMode: string;
};

export type ResourceSchema = {
  name: string;
  label: string;
  mode: ResourceMode;
  fields: ResourceField[];
};

export type ResourceRelatedLink = {
  label: string;
  route: string;
  icon?: string;
  description?: string;
};

export type ResourceEditState = {
  schema: ResourceSchema;
  values: Record<string, string>;
  recordState: ResourceRecordState;
  relatedLinks: ResourceRelatedLink[];
  /** When set, dirty detection compares draft against these defaults (e.g. replicate-from-source). */
  baselineValues?: Record<string, string>;
};

export type ResourceListItem = {
  id: string;
  values: Record<string, string>;
};

export type ResourceSubmitButtonState = {
  disabled: boolean;
  label: string;
  loading: boolean;
};
