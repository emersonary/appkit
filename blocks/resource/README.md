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

## Frontend pattern (schema-driven edit)

Import the protobuf-aligned edit components:

```tsx
import {
  ResourceEdit,
  ResourceFieldInput,
  ResourceFieldKind,
  type ResourceEditState,
} from "@emersonary/appkit-resource";
import "@emersonary/appkit-resource/edit.css";
```

`ResourceFieldInput` maps each `ResourceFieldKind` to a widget (text, textarea, checkbox, country select, image upload, location map, etc.).

Host apps pass navigation via `LinkComponent` and optional `renderRelatedLinkIcon` on `ResourceEdit`.

### Legacy registry UI

The older `FieldType` list/edit components remain available as `LegacyRegistryResourceEdit` and `ResourceList`.

List:

```tsx
<ResourceList
  schema={schema}
  items={items}
  total={total}
  page={page}
  onPageChange={setPage}
  onEdit={(item) => navigate(`/admin/products/${item.id}`)}
/>
```

Edit:

```tsx
<ResourceEdit
  schema={schema}
  item={product}
  relationOptions={{ category_id: categoryOptions }}
  onSubmit={saveProduct}
/>
```

## Tree lists

Tree resources set `parent_id_field` on the schema. The generic list renders a treegrid:

- first/name column is indented by depth
- expand/collapse is local for already-loaded rows
- apps can use `onToggle` to lazy-load children later
- other columns remain regular table columns

This avoids separate "tree" and "table" components for categories, locations, accounts, etc.
