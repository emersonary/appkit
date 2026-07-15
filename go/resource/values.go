package resource

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// Values extracts scalar string values from a struct for editable/readonly visible fields.
func Values(v any) (map[string]string, error) {
	rv, t, err := structValue(v)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		meta := parseResourceTag(sf.Tag.Get("resource"))
		if meta.skip || !meta.visible {
			continue
		}
		key := fieldKey(sf)
		if isUIExcludedKey(key) {
			continue
		}
		value, err := valueToString(rv.Field(i))
		if err != nil {
			return nil, fmt.Errorf("resource: field %q: %w", key, err)
		}
		out[key] = value
	}
	return out, nil
}

func valueToString(v reflect.Value) (string, error) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			if t.IsZero() {
				return "", nil
			}
			return t.UTC().Format(time.RFC3339), nil
		}
	}
	return "", fmt.Errorf("unsupported type %s", v.Type())
}

func structValue(v any) (reflect.Value, reflect.Type, error) {
	if v == nil {
		return reflect.Value{}, nil, fmt.Errorf("resource: nil value")
	}
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return reflect.Value{}, nil, fmt.Errorf("resource: nil pointer")
		}
		rv = rv.Elem()
		rt = rt.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return reflect.Value{}, nil, fmt.Errorf("resource: expected struct, got %s", rv.Kind())
	}
	return rv, rt, nil
}
