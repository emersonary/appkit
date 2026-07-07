package resource

import (
	"context"
	"testing"
)

type memoryStore struct{}

func (memoryStore) List(context.Context, ListRequest) (ListResponse, error) {
	return ListResponse{}, nil
}
func (memoryStore) Get(context.Context, string, string) (Item, error)  { return Item{}, nil }
func (memoryStore) Create(context.Context, string, Item) (Item, error) { return Item{}, nil }
func (memoryStore) Update(context.Context, string, string, Item) (Item, error) {
	return Item{}, nil
}
func (memoryStore) Delete(context.Context, string, string) error { return nil }

func TestResourceNormalizeAddsIdentityAndTreeDefaults(t *testing.T) {
	res := Resource{
		ID:            "category",
		Name:          "Category",
		ParentIDField: "parent_id",
		Fields: []Field{
			{Key: "description", Type: FieldTypeTextarea, Section: "Details"},
		},
	}

	res.Normalize()
	if err := res.Validate(); err != nil {
		t.Fatal(err)
	}
	if res.IDField != "id" || res.NameField != "name" {
		t.Fatalf("unexpected identity fields: %+v", res)
	}
	if !res.List.Tree.Enabled || res.List.Tree.ParentIDField != "parent_id" {
		t.Fatalf("expected tree list defaults, got %+v", res.List.Tree)
	}
	if !hasField(res.Fields, "id") || !hasField(res.Fields, "name") || !hasField(res.Fields, "parent_id") {
		t.Fatalf("expected identity and parent fields, got %+v", res.Fields)
	}
}

func TestRegistryRejectsDuplicateResources(t *testing.T) {
	registry, err := NewRegistry(ResourceDefinition{
		Schema: Resource{ID: "product", Name: "Product"},
		Store:  memoryStore{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(ResourceDefinition{
		Schema: Resource{ID: "product", Name: "Product"},
		Store:  memoryStore{},
	}); err == nil {
		t.Fatal("expected duplicate registration error")
	}
}

func TestFieldsFromStructInfersAndReadsResourceTags(t *testing.T) {
	type Product struct {
		ID     string  `json:"id" resource:"type=uuid,readonly"`
		Name   string  `json:"name" resource:"required,section=Identity"`
		Price  float64 `json:"price" resource:"type=money,section=Pricing"`
		Secret string  `json:"secret" resource:"hidden"`
	}

	fields := FieldsFromStruct(Product{})
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %+v", fields)
	}
	if fields[0].Type != FieldTypeUUID || !fields[0].ReadOnly {
		t.Fatalf("unexpected ID field: %+v", fields[0])
	}
	if !fields[1].Required || fields[1].Section != "Identity" {
		t.Fatalf("unexpected name field: %+v", fields[1])
	}
	if fields[2].Type != FieldTypeMoney || fields[2].Section != "Pricing" {
		t.Fatalf("unexpected price field: %+v", fields[2])
	}
	if !fields[3].Hidden {
		t.Fatalf("expected hidden field, got %+v", fields[3])
	}
}
