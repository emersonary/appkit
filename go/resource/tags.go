package resource

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type fieldMeta struct {
	skip              bool
	readonly          bool
	label             string
	kind              Kind
	section           string
	order             int
	required          bool
	listable          bool
	editable          bool
	visible           bool
	placeholder       string
	help              string
	maxLength         int
	maxWidth          int
	maxHeight         int
	watchSection      string
	bindSection       string
	locationMode      LocationMode
	validations       []ValidationKind
	validationParams  map[ValidationKind]string
}

func parseResourceTag(raw string) fieldMeta {
	meta := fieldMeta{
		editable: true,
		visible:  true,
		order:    100,
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return meta
	}
	if raw == "-" {
		meta.skip = true
		return meta
	}

	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		switch {
		case part == "readonly":
			meta.readonly = true
			meta.editable = false
		case part == "required":
			meta.required = true
		case part == "listable":
			meta.listable = true
		case part == "hidden":
			meta.visible = false
		case strings.HasPrefix(part, "label="):
			meta.label = strings.TrimPrefix(part, "label=")
		case strings.HasPrefix(part, "kind="):
			meta.kind = Kind(strings.TrimPrefix(part, "kind="))
		case strings.HasPrefix(part, "section="):
			meta.section = strings.TrimPrefix(part, "section=")
		case strings.HasPrefix(part, "order="):
			if value, err := strconv.Atoi(strings.TrimPrefix(part, "order=")); err == nil {
				meta.order = value
			}
		case strings.HasPrefix(part, "placeholder="):
			meta.placeholder = strings.TrimPrefix(part, "placeholder=")
		case strings.HasPrefix(part, "help="):
			meta.help = strings.TrimPrefix(part, "help=")
		case strings.HasPrefix(part, "validate="):
			meta.validations = append(meta.validations, ValidationKind(strings.TrimPrefix(part, "validate=")))
		case strings.HasPrefix(part, "max="):
			if value, err := strconv.Atoi(strings.TrimPrefix(part, "max=")); err == nil {
				meta.maxLength = value
			}
		case strings.HasPrefix(part, "max_width="):
			if value, err := strconv.Atoi(strings.TrimPrefix(part, "max_width=")); err == nil {
				meta.maxWidth = value
			}
		case strings.HasPrefix(part, "max_height="):
			if value, err := strconv.Atoi(strings.TrimPrefix(part, "max_height=")); err == nil {
				meta.maxHeight = value
			}
		case strings.HasPrefix(part, "watch_section="):
			meta.watchSection = strings.TrimPrefix(part, "watch_section=")
		case strings.HasPrefix(part, "bind_section="):
			meta.bindSection = strings.TrimPrefix(part, "bind_section=")
		case strings.HasPrefix(part, "location_mode="):
			meta.locationMode = LocationMode(strings.TrimPrefix(part, "location_mode="))
		case strings.HasPrefix(part, "pattern="):
			if meta.validationParams == nil {
				meta.validationParams = make(map[ValidationKind]string)
			}
			meta.validationParams[ValidationPattern] = strings.TrimPrefix(part, "pattern=")
			meta.validations = append(meta.validations, ValidationPattern)
		}
	}
	return meta
}

func fieldKey(field reflect.StructField) string {
	if db := field.Tag.Get("db"); db != "" && db != "-" {
		return db
	}
	if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		name, _, _ := strings.Cut(jsonTag, ",")
		if name != "" {
			return name
		}
	}
	return toSnakeCase(field.Name)
}

func defaultLabel(field reflect.StructField) string {
	key := fieldKey(field)
	key = strings.ReplaceAll(key, "_", " ")
	if key == "" {
		return key
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

func inferKind(field reflect.StructField) Kind {
	switch field.Type.Kind() {
	case reflect.Bool:
		return KindCheckbox
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return KindNumber
	case reflect.Struct:
		if field.Type.String() == "time.Time" {
			return KindDateTime
		}
	}
	return KindText
}

func toSnakeCase(value string) string {
	if value == "" {
		return value
	}
	var b strings.Builder
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
