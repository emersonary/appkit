package language

import "time"

type Language struct {
	IDLanguage string
	Name       string
	NativeName string
	Direction  string
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// IDLanguage is an alias for the reference column name used in foreign keys.
const IDLanguageColumn = "id_language"
