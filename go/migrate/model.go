package migrate

type Column struct {
	Name         string
	Type         string
	UdtName      string
	IsNullable   bool
	DefaultValue string
	IDAuditField int64
}

type Table struct {
	Schema       string
	Name         string
	PrimaryKey   string
	IDAuditTable int64
	Audit        bool // COMMENT contains AUDIT=true
	Repo         bool // COMMENT contains REPO=true
	Columns      []Column
	Children     []ChildRelation
}

type ChildRelation struct {
	Schema     string
	Table      string
	PrimaryKey string
	FKColumn   string
	JSONKey    string
	IsOneToOne bool
}

type ForeignKey struct {
	ChildSchema  string
	ChildTable   string
	ChildColumn  string
	ParentSchema string
	ParentTable  string
	ParentColumn string
	IsOneToOne   bool
}
