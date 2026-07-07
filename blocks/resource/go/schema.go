package resource

import (
	"fmt"
	"sort"
	"strings"
)

const (
	DefaultIDField   = "id"
	DefaultNameField = "name"
)

type FieldType string

const (
	FieldTypeUUID     FieldType = "uuid"
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeMoney    FieldType = "money"
	FieldTypeBool     FieldType = "bool"
	FieldTypeDate     FieldType = "date"
	FieldTypeDateTime FieldType = "datetime"
	FieldTypeEnum     FieldType = "enum"
	FieldTypeRelation FieldType = "relation"
	FieldTypeJSON     FieldType = "json"
)

type FieldOption struct {
	Value string `json:"value" yaml:"value" mapstructure:"value"`
	Label string `json:"label" yaml:"label" mapstructure:"label"`
}

type Relation struct {
	ResourceID   string `json:"resource_id" yaml:"resource_id" mapstructure:"resource_id"`
	ValueField   string `json:"value_field" yaml:"value_field" mapstructure:"value_field"`
	DisplayField string `json:"display_field" yaml:"display_field" mapstructure:"display_field"`
}

type Field struct {
	Key        string        `json:"key" yaml:"key" mapstructure:"key"`
	Label      string        `json:"label" yaml:"label" mapstructure:"label"`
	Type       FieldType     `json:"type" yaml:"type" mapstructure:"type"`
	Section    string        `json:"section,omitempty" yaml:"section,omitempty" mapstructure:"section"`
	HelpText   string        `json:"help_text,omitempty" yaml:"help_text,omitempty" mapstructure:"help_text"`
	Required   bool          `json:"required,omitempty" yaml:"required,omitempty" mapstructure:"required"`
	ReadOnly   bool          `json:"read_only,omitempty" yaml:"read_only,omitempty" mapstructure:"read_only"`
	Hidden     bool          `json:"hidden,omitempty" yaml:"hidden,omitempty" mapstructure:"hidden"`
	ListHidden bool          `json:"list_hidden,omitempty" yaml:"list_hidden,omitempty" mapstructure:"list_hidden"`
	FormHidden bool          `json:"form_hidden,omitempty" yaml:"form_hidden,omitempty" mapstructure:"form_hidden"`
	SortOrder  int           `json:"sort_order,omitempty" yaml:"sort_order,omitempty" mapstructure:"sort_order"`
	Options    []FieldOption `json:"options,omitempty" yaml:"options,omitempty" mapstructure:"options"`
	Relation   *Relation     `json:"relation,omitempty" yaml:"relation,omitempty" mapstructure:"relation"`
}

type Column struct {
	FieldKey  string `json:"field_key" yaml:"field_key" mapstructure:"field_key"`
	Label     string `json:"label,omitempty" yaml:"label,omitempty" mapstructure:"label"`
	Width     string `json:"width,omitempty" yaml:"width,omitempty" mapstructure:"width"`
	SortOrder int    `json:"sort_order,omitempty" yaml:"sort_order,omitempty" mapstructure:"sort_order"`
	Hidden    bool   `json:"hidden,omitempty" yaml:"hidden,omitempty" mapstructure:"hidden"`
}

type TreeConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	ParentIDField string `json:"parent_id_field,omitempty" yaml:"parent_id_field,omitempty" mapstructure:"parent_id_field"`
	NameField     string `json:"name_field,omitempty" yaml:"name_field,omitempty" mapstructure:"name_field"`
	LazyLoad      bool   `json:"lazy_load,omitempty" yaml:"lazy_load,omitempty" mapstructure:"lazy_load"`
}

type SortField struct {
	FieldKey string `json:"field_key" yaml:"field_key" mapstructure:"field_key"`
	Desc     bool   `json:"desc,omitempty" yaml:"desc,omitempty" mapstructure:"desc"`
}

type ListView struct {
	PageSize         int         `json:"page_size,omitempty" yaml:"page_size,omitempty" mapstructure:"page_size"`
	Columns          []Column    `json:"columns,omitempty" yaml:"columns,omitempty" mapstructure:"columns"`
	Tree             TreeConfig  `json:"tree,omitempty" yaml:"tree,omitempty" mapstructure:"tree"`
	SearchableFields []string    `json:"searchable_fields,omitempty" yaml:"searchable_fields,omitempty" mapstructure:"searchable_fields"`
	DefaultSort      []SortField `json:"default_sort,omitempty" yaml:"default_sort,omitempty" mapstructure:"default_sort"`
}

type FormSection struct {
	ID          string   `json:"id" yaml:"id" mapstructure:"id"`
	Title       string   `json:"title" yaml:"title" mapstructure:"title"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	Fields      []string `json:"fields,omitempty" yaml:"fields,omitempty" mapstructure:"fields"`
	SortOrder   int      `json:"sort_order,omitempty" yaml:"sort_order,omitempty" mapstructure:"sort_order"`
}

type FormView struct {
	Sections []FormSection `json:"sections,omitempty" yaml:"sections,omitempty" mapstructure:"sections"`
}

type Resource struct {
	ID            string   `json:"id" yaml:"id" mapstructure:"id"`
	Name          string   `json:"name" yaml:"name" mapstructure:"name"`
	Description   string   `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	IDField       string   `json:"id_field,omitempty" yaml:"id_field,omitempty" mapstructure:"id_field"`
	NameField     string   `json:"name_field,omitempty" yaml:"name_field,omitempty" mapstructure:"name_field"`
	ParentIDField string   `json:"parent_id_field,omitempty" yaml:"parent_id_field,omitempty" mapstructure:"parent_id_field"`
	Fields        []Field  `json:"fields" yaml:"fields" mapstructure:"fields"`
	List          ListView `json:"list,omitempty" yaml:"list,omitempty" mapstructure:"list"`
	Form          FormView `json:"form,omitempty" yaml:"form,omitempty" mapstructure:"form"`
}

func (r *Resource) Normalize() {
	r.ID = strings.TrimSpace(r.ID)
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
	r.IDField = defaultString(strings.TrimSpace(r.IDField), DefaultIDField)
	r.NameField = defaultString(strings.TrimSpace(r.NameField), DefaultNameField)
	r.ParentIDField = strings.TrimSpace(r.ParentIDField)

	for i := range r.Fields {
		normalizeField(&r.Fields[i])
	}
	r.ensureIdentityFields()
	r.normalizeList()
	r.normalizeForm()
}

func (r Resource) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return invalidResource("id", "required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return invalidResource("name", "required")
	}
	if strings.TrimSpace(r.IDField) == "" {
		return invalidResource("id_field", "required")
	}
	if strings.TrimSpace(r.NameField) == "" {
		return invalidResource("name_field", "required")
	}

	fields := make(map[string]Field, len(r.Fields))
	for _, field := range r.Fields {
		if field.Key == "" {
			return invalidResource("fields", "field key is required")
		}
		if _, ok := fields[field.Key]; ok {
			return invalidResourcef("fields.%s", field.Key, "duplicate field key")
		}
		if field.Type == "" {
			return invalidResourcef("fields.%s.type", field.Key, "required")
		}
		if field.Type == FieldTypeRelation {
			if field.Relation == nil || strings.TrimSpace(field.Relation.ResourceID) == "" {
				return invalidResourcef("fields.%s.relation.resource_id", field.Key, "required for relation fields")
			}
		}
		fields[field.Key] = field
	}

	for _, required := range []string{r.IDField, r.NameField} {
		if _, ok := fields[required]; !ok {
			return invalidResourcef("fields.%s", required, "identity field is missing")
		}
	}
	if r.ParentIDField != "" {
		if _, ok := fields[r.ParentIDField]; !ok {
			return invalidResourcef("fields.%s", r.ParentIDField, "parent field is missing")
		}
	}
	for _, col := range r.List.Columns {
		if col.Hidden {
			continue
		}
		if _, ok := fields[col.FieldKey]; !ok {
			return invalidResourcef("list.columns.%s", col.FieldKey, "unknown field")
		}
	}
	for _, section := range r.Form.Sections {
		if section.ID == "" {
			return invalidResource("form.sections", "section id is required")
		}
		for _, key := range section.Fields {
			if _, ok := fields[key]; !ok {
				return invalidResourcef("form.sections.%s.fields.%s", section.ID, key, "unknown field")
			}
		}
	}
	return nil
}

func normalizeField(field *Field) {
	field.Key = strings.TrimSpace(field.Key)
	field.Label = defaultString(strings.TrimSpace(field.Label), titleFromKey(field.Key))
	field.Type = FieldType(strings.TrimSpace(string(field.Type)))
	if field.Type == "" {
		field.Type = FieldTypeText
	}
	field.Section = strings.TrimSpace(field.Section)
	field.HelpText = strings.TrimSpace(field.HelpText)
	if field.Relation != nil {
		field.Relation.ResourceID = strings.TrimSpace(field.Relation.ResourceID)
		field.Relation.ValueField = defaultString(strings.TrimSpace(field.Relation.ValueField), DefaultIDField)
		field.Relation.DisplayField = defaultString(strings.TrimSpace(field.Relation.DisplayField), DefaultNameField)
	}
}

func (r *Resource) ensureIdentityFields() {
	if !hasField(r.Fields, r.IDField) {
		r.Fields = append([]Field{{
			Key:      r.IDField,
			Label:    "ID",
			Type:     FieldTypeUUID,
			ReadOnly: true,
		}}, r.Fields...)
	}
	if !hasField(r.Fields, r.NameField) {
		r.Fields = append(r.Fields, Field{
			Key:      r.NameField,
			Label:    "Name",
			Type:     FieldTypeText,
			Required: true,
		})
	}
	if r.ParentIDField != "" && !hasField(r.Fields, r.ParentIDField) {
		r.Fields = append(r.Fields, Field{
			Key:      r.ParentIDField,
			Label:    "Parent",
			Type:     FieldTypeRelation,
			Section:  "Hierarchy",
			Required: false,
			Relation: &Relation{ResourceID: r.ID, ValueField: r.IDField, DisplayField: r.NameField},
		})
	}
}

func (r *Resource) normalizeList() {
	if r.List.PageSize <= 0 {
		r.List.PageSize = 25
	}
	if r.ParentIDField != "" {
		r.List.Tree.Enabled = true
		r.List.Tree.ParentIDField = defaultString(strings.TrimSpace(r.List.Tree.ParentIDField), r.ParentIDField)
		r.List.Tree.NameField = defaultString(strings.TrimSpace(r.List.Tree.NameField), r.NameField)
	}
	if len(r.List.Columns) == 0 {
		for _, field := range sortedFields(r.Fields) {
			if field.Hidden || field.ListHidden {
				continue
			}
			r.List.Columns = append(r.List.Columns, Column{FieldKey: field.Key, Label: field.Label, SortOrder: field.SortOrder})
		}
	}
}

func (r *Resource) normalizeForm() {
	if len(r.Form.Sections) > 0 {
		for i := range r.Form.Sections {
			r.Form.Sections[i].ID = strings.TrimSpace(r.Form.Sections[i].ID)
			r.Form.Sections[i].Title = defaultString(strings.TrimSpace(r.Form.Sections[i].Title), titleFromKey(r.Form.Sections[i].ID))
		}
		return
	}

	bySection := map[string][]string{}
	for _, field := range sortedFields(r.Fields) {
		if field.Hidden || field.FormHidden {
			continue
		}
		section := defaultString(field.Section, "General")
		bySection[section] = append(bySection[section], field.Key)
	}
	sections := make([]string, 0, len(bySection))
	for section := range bySection {
		sections = append(sections, section)
	}
	sort.Strings(sections)
	for _, section := range sections {
		r.Form.Sections = append(r.Form.Sections, FormSection{
			ID:     keyFromTitle(section),
			Title:  section,
			Fields: bySection[section],
		})
	}
}

func sortedFields(fields []Field) []Field {
	out := append([]Field(nil), fields...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortOrder == out[j].SortOrder {
			return out[i].Key < out[j].Key
		}
		return out[i].SortOrder < out[j].SortOrder
	})
	return out
}

func hasField(fields []Field, key string) bool {
	for _, field := range fields {
		if field.Key == key {
			return true
		}
	}
	return false
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func invalidResource(path, msg string) error {
	return fmt.Errorf("resource.%s: %s", path, msg)
}

func invalidResourcef(path, format string, args ...any) error {
	return invalidResource(path, fmt.Sprintf(format, args...))
}

func titleFromKey(key string) string {
	key = strings.TrimSpace(strings.ReplaceAll(key, "_", " "))
	if key == "" {
		return ""
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

func keyFromTitle(title string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(title), " ", "_"))
}
