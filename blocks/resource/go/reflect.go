package resource

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FieldsFromStruct infers default field metadata from exported struct fields.
// It is a convenience for bootstrapping schemas; apps should still register
// explicit Resource definitions before exposing anything to the UI.
func FieldsFromStruct(sample any) []Field {
	t := reflect.TypeOf(sample)
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]Field, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		key := fieldKey(sf)
		if key == "" || key == "-" {
			continue
		}
		field := Field{
			Key:       key,
			Label:     titleFromKey(key),
			Type:      inferFieldType(sf.Type),
			SortOrder: i,
		}
		applyResourceTag(&field, sf.Tag.Get("resource"))
		fields = append(fields, field)
	}
	return fields
}

func fieldKey(sf reflect.StructField) string {
	for _, tagName := range []string{"json", "db"} {
		tag := strings.TrimSpace(sf.Tag.Get(tagName))
		if tag == "" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	return keyFromTitle(sf.Name)
}

func inferFieldType(t reflect.Type) FieldType {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(time.Time{}) {
		return FieldTypeDateTime
	}
	switch t.Kind() {
	case reflect.Bool:
		return FieldTypeBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return FieldTypeNumber
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return FieldTypeObject
	default:
		return FieldTypeText
	}
}

func applyResourceTag(field *Field, tag string) {
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, hasValue := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "label":
			if hasValue {
				field.Label = value
			}
		case "type":
			if hasValue {
				field.Type = FieldType(value)
			}
		case "section":
			if hasValue {
				field.Section = value
			}
		case "help":
			if hasValue {
				field.HelpText = value
			}
		case "required":
			field.Required = true
		case "readonly", "read_only":
			field.ReadOnly = true
		case "hidden":
			field.Hidden = true
		case "list_hidden":
			field.ListHidden = true
		case "form_hidden":
			field.FormHidden = true
		case "order":
			if hasValue {
				if order, err := strconv.Atoi(value); err == nil {
					field.SortOrder = order
				}
			}
		}
	}
}
