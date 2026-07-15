package resource

import "sort"

// ListSchema returns a schema containing only listable fields from Describe, plus optional list-only extras.
func ListSchema(v any, listExtras []Field, opts ...DescribeOption) (Schema, error) {
	full, err := Describe(v, opts...)
	if err != nil {
		return Schema{}, err
	}

	fields := make([]Field, 0, len(full.Fields)+len(listExtras))
	for _, field := range full.Fields {
		if field.Listable {
			fields = append(fields, field)
		}
	}
	for _, field := range listExtras {
		if field.Key == "" || !field.Listable {
			continue
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
		Name:   full.Name,
		Label:  full.Label,
		Mode:   full.Mode,
		Fields: fields,
	}, nil
}
