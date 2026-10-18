package migrate

// GenUUIDSQL is the Postgres expression for platform time-ordered UUID v7.
// Use in DDL DEFAULT clauses and generated repo functions — not gen_random_uuid().
const GenUUIDSQL = "platform.func_gen_uuid_v7()"
