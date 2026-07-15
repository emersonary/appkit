package resource

// UIExcludedKeys are never sent to edit-form clients (schema or values).
var UIExcludedKeys = map[string]struct{}{
	"created_at": {},
	"updated_at": {},
}

func isUIExcludedKey(key string) bool {
	_, ok := UIExcludedKeys[key]
	return ok
}
