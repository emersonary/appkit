package resource

import (
	"fmt"
	"reflect"
	"sort"
)

// Describe builds a UI schema from a struct type or pointer.
func Describe(v any, opts ...DescribeOption) (Schema, error) {
	cfg := applyDescribeOptions(opts)
	t, err := structType(v)
	if err != nil {
		return Schema{}, err
	}

	name := cfg.name
	if name == "" {
		name = toSnakeCase(t.Name())
	}
	label := cfg.label
	if label == "" {
		label = t.Name()
	}

	fields := make([]Field, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		meta := parseResourceTag(sf.Tag.Get("resource"))
		if meta.skip {
			continue
		}

		key := fieldKey(sf)
		if isUIExcludedKey(key) {
			continue
		}
		kind := meta.kind
		if kind == "" {
			kind = inferKind(sf)
		}
		fieldLabel := meta.label
		if fieldLabel == "" {
			fieldLabel = defaultLabel(sf)
		}

		field := Field{
			Key:         key,
			Label:       fieldLabel,
			Kind:        kind,
			Section:     meta.section,
			Order:       meta.order,
			Required:    meta.required,
			ReadOnly:    meta.readonly,
			Editable:    meta.editable && !meta.readonly,
			Visible:     meta.visible,
			Listable:    meta.listable,
			Placeholder: meta.placeholder,
			HelpText:     meta.help,
			MaxWidth:     meta.maxWidth,
			MaxHeight:    meta.maxHeight,
			WatchSection: meta.watchSection,
			BindSection:  meta.bindSection,
			LocationMode: meta.locationMode,
		}
		if override, ok := cfg.overrides[key]; ok {
			override(&field)
		}
		finalizeField(&field, meta)
		if !field.Visible {
			continue
		}
		fields = append(fields, field)
	}

	for _, extra := range cfg.extraFields {
		field := extra
		if field.Key == "" {
			continue
		}
		if !field.Visible {
			field.Visible = true
		}
		if field.LocationMode == "" && field.Kind == KindLocation {
			if field.ReadOnly || !field.Editable {
				field.LocationMode = LocationModePreview
			} else {
				field.LocationMode = LocationModeManual
			}
		}
		if field.Kind == KindLocation && field.WatchSection == "" && field.BindSection != "" {
			field.WatchSection = field.BindSection
		}
		if override, ok := cfg.overrides[field.Key]; ok {
			override(&field)
		}
		fields = append(fields, field)
	}

	sort.SliceStable(fields, func(i, j int) bool {
		if fields[i].Order != fields[j].Order {
			return fields[i].Order < fields[j].Order
		}
		if fields[i].Section != fields[j].Section {
			return fields[i].Section < fields[j].Section
		}
		return fields[i].Key < fields[j].Key
	})

	return Schema{
		Name:   name,
		Label:  label,
		Mode:   cfg.mode,
		Fields: fields,
	}, nil
}

// EditPayload builds schema + values from a loaded struct instance.
func BuildEditPayload(v any, opts ...DescribeOption) (EditPayload, error) {
	schema, err := Describe(v, opts...)
	if err != nil {
		return EditPayload{}, err
	}
	values, err := Values(v)
	if err != nil {
		return EditPayload{}, err
	}
	return EditPayload{
		Schema:      schema,
		Values:      values,
		RecordState: RecordStateExisting,
	}, nil
}

// BuildNewEditPayload builds an insert form from a zero-value struct instance.
func BuildNewEditPayload(v any, opts ...DescribeOption) (EditPayload, error) {
	payload, err := BuildEditPayload(v, opts...)
	if err != nil {
		return EditPayload{}, err
	}
	payload.RecordState = RecordStateNew
	return payload, nil
}

func structType(v any) (reflect.Type, error) {
	if v == nil {
		return nil, fmt.Errorf("resource: nil value")
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("resource: expected struct, got %s", t.Kind())
	}
	return t, nil
}
