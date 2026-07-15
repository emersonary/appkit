package resource

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ApplyPatch validates and writes submitted values onto dst for editable fields only.
func ApplyPatch(dst any, values map[string]string, opts ...DescribeOption) error {
	rv, t, err := structValue(dst)
	if err != nil {
		return err
	}
	if !rv.CanSet() {
		return fmt.Errorf("resource: destination is not settable")
	}

	schema, err := Describe(dst, opts...)
	if err != nil {
		return err
	}
	if err := ValidateValues(schema, values); err != nil {
		return err
	}

	editable := make(map[string]Field, len(schema.Fields))
	for _, field := range schema.Fields {
		if field.Editable {
			editable[field.Key] = field
		}
	}

	for key := range editable {
		raw, ok := values[key]
		if !ok {
			continue
		}
		if err := setFieldByKey(rv, t, key, raw); err != nil {
			return err
		}
	}

	return nil
}

func setFieldByKey(rv reflect.Value, t reflect.Type, key, raw string) error {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if fieldKey(sf) != key {
			continue
		}
		meta := parseResourceTag(sf.Tag.Get("resource"))
		if meta.skip || !meta.visible || meta.readonly || !meta.editable {
			return fmt.Errorf("resource: field %q is not editable", key)
		}
		return setFieldValue(rv.Field(i), sf.Type, raw)
	}
	return fmt.Errorf("resource: unknown field %q", key)
}

func setFieldValue(field reflect.Value, ft reflect.Type, raw string) error {
	target := field
	isPtr := ft.Kind() == reflect.Pointer
	if isPtr {
		if field.IsNil() {
			field.Set(reflect.New(ft.Elem()))
		}
		target = field.Elem()
		ft = ft.Elem()
	}

	switch ft.Kind() {
	case reflect.String:
		target.SetString(raw)
	case reflect.Bool:
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid bool %q", raw)
		}
		target.SetBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer %q", raw)
		}
		target.SetInt(value)
	case reflect.Struct:
		if ft == reflect.TypeOf(time.Time{}) {
			if raw == "" {
				target.Set(reflect.Zero(ft))
				return nil
			}
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return fmt.Errorf("invalid datetime %q", raw)
			}
			target.Set(reflect.ValueOf(parsed))
			return nil
		}
	}
	if isPtr && raw == "" {
		field.Set(reflect.Zero(field.Type()))
	}
	return nil
}
