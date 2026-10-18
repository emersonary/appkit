-- Platform UUID v7 (matches github.com/emersonary/uuidv7 / appkit/uuidv7 Go layout).

CREATE SCHEMA IF NOT EXISTS platform;

CREATE OR REPLACE FUNCTION platform.func_gen_uuid_v7()
RETURNS uuid
LANGUAGE plpgsql
VOLATILE
AS $$
DECLARE
  ts timestamptz := clock_timestamp();
  millis bigint;
  nanos bigint;
  nanos_after bigint;
  b bytea;
  byt int;
  rnd int;
BEGIN
  millis := floor(extract(epoch FROM ts) * 1000)::bigint;
  nanos := floor(extract(epoch FROM ts) * 1000000000)::bigint;
  nanos_after := nanos - millis * 1000000;

  b := gen_random_bytes(16);

  b := set_byte(b, 0, ((millis >> 40) & 255)::int);
  b := set_byte(b, 1, ((millis >> 32) & 255)::int);
  b := set_byte(b, 2, ((millis >> 24) & 255)::int);
  b := set_byte(b, 3, ((millis >> 16) & 255)::int);
  b := set_byte(b, 4, ((millis >> 8) & 255)::int);
  b := set_byte(b, 5, (millis & 255)::int);

  b := set_byte(b, 6, ((nanos_after >> 16) & 255)::int);
  b := set_byte(b, 7, ((nanos_after >> 8) & 255)::int);
  byt := (nanos_after & 255)::int;

  rnd := get_byte(b, 9);
  b := set_byte(b, 8, (get_byte(b, 8) & 15) | ((byt >> 4) & 15));
  b := set_byte(b, 9, (rnd & 15) | ((byt << 4) & 255));

  b := set_byte(b, 6, (get_byte(b, 6) & 15) | 112);
  b := set_byte(b, 8, (get_byte(b, 8) & 63) | 128);

  RETURN encode(b, 'hex')::uuid;
END;
$$;

COMMENT ON FUNCTION platform.func_gen_uuid_v7() IS 'UUID v7 (emersonary/uuidv7 layout); use for all platform PK defaults and repo fallbacks';
