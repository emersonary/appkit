package resource

// Mode describes how a resource is presented in the UI.
type Mode string

const (
	ModeEditOnly    Mode = "edit_only"
	ModeListAndEdit Mode = "list_and_edit"
)

// RecordState tells the UI whether a row already exists in storage.
type RecordState string

const (
	RecordStateExisting RecordState = "existing"
	RecordStateNew      RecordState = "new"
)

// Kind selects the input widget for a field.
type Kind string

const (
	KindText     Kind = "text"
	KindTextarea Kind = "textarea"
	KindEmail    Kind = "email"
	KindPhone    Kind = "phone"
	KindURL      Kind = "url"
	KindImage    Kind = "image"
	KindCountry  Kind = "country"
	KindCheckbox Kind = "checkbox"
	KindNumber   Kind = "number"
	KindDateTime Kind = "datetime"
	KindLocation Kind = "location"
)

// LocationMode controls how a location field is edited and refreshed.
type LocationMode string

const (
	LocationModePreview LocationMode = "preview"
	LocationModeBound   LocationMode = "bound"
	LocationModeManual  LocationMode = "manual"
)

// ValidationKind selects a server/client validation rule.
type ValidationKind string

const (
	ValidationRequired  ValidationKind = "required"
	ValidationEmail     ValidationKind = "email"
	ValidationURL       ValidationKind = "url"
	ValidationMaxLength ValidationKind = "max_length"
	ValidationPattern   ValidationKind = "pattern"
)

// FieldOption is one selectable value (e.g. country code).
type FieldOption struct {
	Value string
	Label string
}

// ValidationRule is a declarative constraint on a field value.
type ValidationRule struct {
	Kind  ValidationKind
	Param string
}

// Field describes one attribute exposed to the UI.
type Field struct {
	Key          string
	Label        string
	Kind         Kind
	Section      string
	Order        int
	Required     bool
	ReadOnly     bool
	Editable     bool
	Visible      bool
	Listable     bool
	Placeholder  string
	HelpText     string
	MaxWidth     int
	MaxHeight    int
	WatchSection string
	BindSection  string
	LocationMode LocationMode
	Options      []FieldOption
	Validations  []ValidationRule
}

// Schema is the UI contract for one table-backed struct.
type Schema struct {
	Name   string
	Label  string
	Mode   Mode
	Fields []Field
}

// EditPayload combines schema metadata with current scalar values.
type EditPayload struct {
	Schema      Schema
	Values      map[string]string
	RecordState RecordState
}
