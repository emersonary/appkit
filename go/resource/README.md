# resource

Reflect on Go structs tagged with `resource:"..."` to produce UI schemas and scalar value maps for schema-driven edit (and future list) screens.

See Solidia `docs/architecture/resource-edit-pattern.md` for the end-to-end flow across API and Command Center.

```go
schema, _ := resource.Describe(BusinessProfile{}, resource.WithName("tenant_business_profile"))
payload, _ := resource.BuildEditPayload(loadedProfile, resource.WithLabel("Business profile"))
_ = resource.ApplyPatch(&loadedProfile, submittedValues)
```
