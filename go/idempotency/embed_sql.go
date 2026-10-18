package idempotency

import _ "embed"

//go:embed migration.gen.sql
var MigrationSQL string

//go:embed migrate_extras.sql
var MigrationExtrasSQL string

//go:embed migrate_post.sql
var MigrationPostSQL string
