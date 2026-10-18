package migrate

import _ "embed"

//go:embed platform_uuid_v7.sql
var platformUUIDv7SQL string

const InstructionPlatformUUIDv7 = "appkit.platform.uuid_v7.v2"

// PlatformInstructions returns SQL that must run before any table DEFAULT or repo
// function references platform.func_gen_uuid_v7().
func PlatformInstructions() []Instruction {
	return []Instruction{{
		ID:     InstructionPlatformUUIDv7,
		SQL:    []string{platformUUIDv7SQL},
		Source: "migrate/platform_uuid_v7.sql",
	}}
}
