# language (Go)

Config-driven language metadata: static catalog, schema bootstrap, and DB seeding for enabled locales.

## Config (main `config.yaml` `language` node)

See `language.example.yaml` for the full shape.

```yaml
language:
  enabled: true
  schema: language
  default_language: en
  languages:
    - en
    - pt
    - fr
```

## Apply schema (migration time)

```go
if err := language.ApplySchemaFromAppConfig(ctx, sqlDB, appCfg.Language); err != nil {
    return err
}
```

## Service

```go
svc, err := language.NewService(sqlDB, cfg, language.Options{})
```

Use `ValidateCode`, `ListLanguages`, and `GetLanguage` from domain handlers. There is no RPC transport in this package.
