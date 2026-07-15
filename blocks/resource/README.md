# appkit resource block

Generic resource metadata and UI building blocks for CRUD-heavy applications.

The block is intentionally schema-driven, not database-reflection-driven. Apps register explicit resources and may use reflection only to bootstrap default field metadata.

## Backend pattern

Register one resource per entity/table:

```go
product := resource.Resource{
	ID:            "product",
	Name:          "Products",
	IDField:       "id",
	NameField:     "name",
	ParentIDField: "", // set this for trees
	Fields: []resource.Field{
		{Key: "sku", Label: "SKU", Type: resource.FieldTypeText, Section: "Identity"},
		{Key: "price", Label: "Price", Type: resource.FieldTypeMoney, Section: "Pricing"},
	},
}

registry, err := resource.NewRegistry(resource.ResourceDefinition{
	Schema: product,
	Store:  productStore,
})
```

Every resource normalizes to an `id` field and a `name` field by default. If `ParentIDField` is set, the list metadata becomes tree-aware.

## Store adapter

Persistence stays in the host app:

```go
type Store interface {
	List(ctx context.Context, req resource.ListRequest) (resource.ListResponse, error)
	Get(ctx context.Context, resourceID, id string) (resource.Item, error)
	Create(ctx context.Context, resourceID string, values resource.Item) (resource.Item, error)
	Update(ctx context.Context, resourceID, id string, values resource.Item) (resource.Item, error)
	Delete(ctx context.Context, resourceID, id string) error
}
```

This lets simple reference data use generic CRUD while complex business objects still go through domain services.

## Reflection helper

`FieldsFromStruct` can infer field defaults from exported struct fields:

```go
type Product struct {
	ID    string  `json:"id" resource:"type=uuid,readonly"`
	Name  string  `json:"name" resource:"required,section=Identity"`
	Price float64 `json:"price" resource:"type=money,section=Pricing"`
}

fields := resource.FieldsFromStruct(Product{})
```

Use this to reduce boilerplate, then explicitly register the final `Resource`.

## Frontend pattern

Import the components and CSS:

```tsx
import {
  ResourceListAndEdit,
  type ResourceSchema,
} from "@emersonary/appkit-resource";
import "@emersonary/appkit-resource/resource.css";
```

`ResourceListAndEdit` combines list and edit flows. While editing, the list is replaced by the form. Back returns to the list.

```tsx
<ResourceListAndEdit
  schema={schema}
  items={items}
  total={total}
  page={page}
  onPageChange={setPage}
  editForm={{
    relationOptions: { category_id: categoryOptions },
    onEditRequest: async (item, mode) => {
      const full = await client.get(schema.id, String(item.id));
      if (mode === "replicate") {
        delete full.id;
      }
      return full;
    },
    onSubmit: async (values, { mode }) => {
      if (mode === "edit") {
        await client.update(schema.id, String(values.id), values);
      } else {
        await client.create(schema.id, values);
      }
      await refreshList();
    },
    onDelete: async (item) => {
      await client.delete(schema.id, String(item.id));
    },
    onDeleted: refreshList,
    confirmDelete: async (item) =>
      window.confirm(`Delete "${item.name}"? This cannot be undone.`),
    interceptField: ({ field, value, onChange, readOnly }) => {
      if (field.key === "address") {
        return <AddressEditor value={value} readOnly={readOnly} onChange={(next) => onChange(field.key, next)} />;
      }
      return null;
    },
  }}
  actions={{ create: true, edit: true, replicate: true, delete: true }}
/>
```

`editForm` receives the same field-rendering options as `ResourceEdit`, except `schema` and `item` come from the view.

- `onEditRequest` loads the full record before opening edit or replicate. List rows may contain fewer fields than the form.
- `onSubmit` receives `{ mode }` so the caller can branch between create, edit, and replicate.
- Throw from `onSubmit` to keep the form open and show the error message.
- Save disables the submit button while `onSubmit` runs.
- Enable delete with `actions.delete` and `editForm.onDelete`. Optional `confirmDelete` and `onDeleted` handle confirmation and list refresh.

Use `actions` to enable create, edit, replicate, and delete row buttons. Create, edit, and replicate are enabled by default; delete is off by default.

## Lower-level components

Use `ResourceList` and `ResourceEdit` directly when you need custom navigation or layout.

List:

```tsx
<ResourceList
  schema={schema}
  items={items}
  total={total}
  page={page}
  onPageChange={setPage}
  rowActions={[
    { id: "edit", label: "Edit", onAction: (item) => openEditor(item) },
  ]}
/>
```

Edit:

```tsx
<ResourceEdit
  schema={schema}
  item={product}
  mode="edit"
  error={saveError}
  relationOptions={{ category_id: categoryOptions }}
  onSubmit={saveProduct}
  onCancel={closeEditor}
  cancelLabel="Back"
/>
```

## Custom field rendering

By default, `ResourceEdit` maps each `field.type` to a built-in control via `FieldRenderer`.

To override rendering for specific fields in a project, pass `interceptField`. Return a React node for fields you want to customize. Return `null` or `undefined` to keep the default renderer.

```tsx
<ResourceEdit
  schema={schema}
  item={product}
  interceptField={({ field, value, onChange, readOnly }) => {
    if (field.type === "json" && field.key === "address") {
      return (
        <AddressEditor
          value={value}
          readOnly={readOnly}
          onChange={(next) => onChange(field.key, next)}
        />
      );
    }

    return null;
  }}
  onSubmit={saveProduct}
/>
```

Match on `field.type`, `field.key`, or both. Use `onChange(field.key, nextValue)` to update form state.

`renderField` is still available when you want to own rendering for every field. If both are passed, `renderField` takes precedence.

## Custom list cells

`ResourceList` uses the same interception pattern via `interceptCell`. Return a React node to override a cell, or `null` / `undefined` to keep the default string formatting.

```tsx
<ResourceList
  schema={schema}
  items={items}
  interceptCell={({ column, value }) => {
    if (column.field_key === "price") {
      return formatMoney(value);
    }

    return null;
  }}
/>
```

`renderCell` remains available for full control over every cell.

## Tree lists

Tree resources set `parent_id_field` on the schema. The generic list renders a treegrid:

- first/name column is indented by depth
- expand/collapse is local for already-loaded rows
- apps can use `onToggle` to lazy-load children later
- other columns remain regular table columns

This avoids separate "tree" and "table" components for categories, locations, accounts, etc.
